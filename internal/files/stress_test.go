package files

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/KhizirFarrukh/local-ai-nas/internal/apperr"
	"github.com/KhizirFarrukh/local-ai-nas/internal/storage"
)

// TestStressConcurrentWriters is the S01.6-T05 acceptance test: 50
// concurrent writers to one name give one consistent result, with each
// policy, and leave no partial or temporary file, while readers see only
// complete versions. The Linux CI job runs it with -race.
func TestStressConcurrentWriters(t *testing.T) {
	const writers, size = 50, 64 << 10
	payload := func(i int) []byte {
		b := bytes.Repeat([]byte{byte(i)}, size)
		copy(b, fmt.Sprintf("writer %02d|", i))
		return b
	}
	complete := map[string]int{} // every writer's whole file
	for i := range writers {
		complete[string(payload(i))] = i
	}
	for _, policy := range []OnConflict{ConflictOverwrite, ConflictFail, ConflictRename} {
		t.Run(string(policy), func(t *testing.T) {
			s, l := newService(t, nil, map[string]string{"up/.keep": ""})
			ctx := t.Context()

			// Readers download the name the whole time.
			stop := make(chan struct{})
			var readers sync.WaitGroup
			var reads, retries atomic.Int64
			for range 4 {
				readers.Go(func() {
					for {
						select {
						case <-stop:
							return
						default:
						}
						_, body, err := s.Download(ctx, owner, "/up/same.bin")
						if err != nil {
							// Not there yet, or replaced on every attempt to
							// open it: both are answers, never partial data.
							if k := apperr.KindOf(err); k != apperr.NotFound && k != apperr.Conflict {
								t.Errorf("reader: %v", err)
							}
							retries.Add(1)
							continue
						}
						b, err := io.ReadAll(body)
						_ = body.Close()
						if _, ok := complete[string(b)]; err != nil || !ok {
							t.Errorf("a reader got %d bytes that are no writer's complete file (%v)", len(b), err)
						}
						reads.Add(1)
					}
				})
			}

			errs := make([]error, writers)
			items := make([]Item, writers)
			created := make([]bool, writers)
			var wg sync.WaitGroup
			for i := range writers {
				wg.Go(func() {
					items[i], created[i], errs[i] = s.Upload(ctx, owner, "/up/same.bin", bytes.NewReader(payload(i)), size, UploadOptions{OnConflict: policy})
				})
			}
			wg.Wait()
			close(stop)
			readers.Wait()

			ok, news := 0, 0
			for i, err := range errs {
				switch {
				case err == nil:
					ok++
					if created[i] {
						news++
					}
				case apperr.KindOf(err) != apperr.Conflict:
					t.Errorf("writer %d: %v", i, err)
				}
			}
			disk := onDisk(t, l)
			for name := range disk {
				if strings.Contains(name, storage.TempPrefix) {
					t.Errorf("a temporary file was left: %s", name)
				}
			}
			final, finalOK := complete[disk["up/same.bin"]]
			switch policy {
			case ConflictOverwrite:
				// Every writer succeeds; the lock makes exactly one of them
				// the creator, and the file is one writer's whole content.
				if ok != writers || news != 1 || !finalOK || len(disk) != 2 {
					t.Errorf("%d succeeded (%d created), final file complete: %v, %d files", ok, news, finalOK, len(disk))
				}
			case ConflictFail:
				winner := -1
				for i, err := range errs {
					if err == nil {
						winner = i
					}
				}
				if ok != 1 || !finalOK || final != winner || len(disk) != 2 {
					t.Errorf("%d succeeded, final file is writer %d's (complete: %v), winner %d, %d files", ok, final, finalOK, winner, len(disk))
				}
			case ConflictRename:
				if ok != writers || len(disk) != writers+1 {
					t.Errorf("%d succeeded, %d files", ok, len(disk))
				}
				for i, it := range items {
					if disk[it.RelPath] != string(payload(i)) {
						t.Errorf("writer %d's file %s does not hold its content", i, it.RelPath)
					}
				}
			}
			t.Logf("%s: %d complete reads and %d not-found or retry answers during the writes", policy, reads.Load(), retries.Load())
		})
	}
}
