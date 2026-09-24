package files

import (
	"fmt"
	"testing"
)

// BenchmarkListLargeFolder lists the first page of a 10,000-entry folder
// through the files service, for each sort key (S01.7-T04, NFR-003). Every
// page reads and sorts the whole folder.
func BenchmarkListLargeFolder(b *testing.B) {
	files := make(map[string]string, 10000)
	for i := range 10000 {
		files[fmt.Sprintf("big/file-%05d.txt", i)] = ""
	}
	s, _ := newService(b, nil, files)
	for _, sort := range []SortKey{SortName, SortSize, SortModTime} {
		b.Run(string(sort), func(b *testing.B) {
			for b.Loop() {
				if _, err := s.List(b.Context(), owner, "/big", ListOptions{Limit: 100, Sort: sort}); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
