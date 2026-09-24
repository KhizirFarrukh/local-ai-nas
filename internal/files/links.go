package files

import (
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/KhizirFarrukh/local-ai-nas/internal/apperr"
)

// refuseLinkParents refuses a path whose parent folders include a
// symbolic link, or another special entry such as a Windows junction:
// links in the area are never followed (S01.6-T03). os.Root follows a link
// that stays inside the area, so this check comes first; a link that
// leaves the area is refused by os.Root as well. The last element is not
// checked: an operation on a link acts on the link itself. A missing
// element or a file ends the check; the operation reports those as usual.
func refuseLinkParents(root *os.Root, rel, apiPath string) error {
	dir := path.Dir(rel)
	if dir == "." {
		return nil
	}
	cur := ""
	for _, name := range strings.Split(dir, "/") {
		cur = path.Join(cur, name)
		info, err := root.Lstat(filepath.FromSlash(cur))
		if err != nil {
			return nil
		}
		if info.Mode()&(fs.ModeSymlink|fs.ModeIrregular) != 0 {
			return apperr.Newf(apperr.InvalidRequest, "%s goes through the symbolic link /%s; links are not followed", apiPath, cur)
		}
		if !info.IsDir() {
			return nil
		}
	}
	return nil
}
