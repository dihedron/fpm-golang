// Package golang discovers upstream Go releases and plans their repackaging.
package golang

import (
	"fmt"
	"regexp"
	"strconv"
)

// versionRE is the only shape of version allowed to reach tags, filenames and
// shell steps: stable releases only (no rc/beta suffixes), "go" prefix optional.
var versionRE = regexp.MustCompile(`^(?:go)?([0-9]{1,4})\.([0-9]{1,4})(?:\.([0-9]{1,4}))?$`)

// Version is an upstream stable Go version; go1.27 (patch absent) is kept as such.
type Version struct {
	Major, Minor, Patch int
	hasPatch            bool
}

// ParseVersion parses "go1.27.4", "1.27.4" or "go1.27".
func ParseVersion(s string) (Version, error) {
	m := versionRE.FindStringSubmatch(s)
	if m == nil {
		return Version{}, fmt.Errorf("invalid Go version %q", s)
	}
	v := Version{hasPatch: m[3] != ""}
	v.Major, _ = strconv.Atoi(m[1])
	v.Minor, _ = strconv.Atoi(m[2])
	if v.hasPatch {
		v.Patch, _ = strconv.Atoi(m[3])
	}
	return v, nil
}

func (v Version) String() string {
	if !v.hasPatch {
		return fmt.Sprintf("%d.%d", v.Major, v.Minor)
	}
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

// Compare returns <0, 0, >0; go1.27 equals go1.27.0.
func (v Version) Compare(o Version) int {
	for _, d := range [3]int{v.Major - o.Major, v.Minor - o.Minor, v.Patch - o.Patch} {
		if d != 0 {
			return d
		}
	}
	return 0
}
