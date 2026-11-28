package golang

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"testing"
)

// The tarball may only hold plain files and directories, each once, below go/.
func TestVerifyContentRejectsOddEntries(t *testing.T) {
	good := []*tar.Header{{Name: "go/", Typeflag: tar.TypeDir}, {Name: "go/bin/go", Typeflag: tar.TypeReg}, {Name: "go/src/a.go", Typeflag: tar.TypeReg}}
	content := map[string]string{"go/bin/go": "BIN", "go/src/a.go": "A"}
	module := mkModule(t, fl{"bin/go": "BIN", "src/a.go": "A"})
	verify := func(extra ...*tar.Header) error {
		var b bytes.Buffer
		gz := gzip.NewWriter(&b)
		tw := tar.NewWriter(gz)
		for _, h := range append(good[:len(good):len(good)], extra...) {
			c := ""
			if h.Typeflag == tar.TypeReg {
				if c = content[h.Name]; c == "" {
					c = "evil"
				}
			}
			h.Mode, h.Size = 0o644, int64(len(c))
			if err := tw.WriteHeader(h); err != nil {
				t.Fatal(err)
			}
			_, _ = tw.Write([]byte(c))
		}
		_ = tw.Close()
		_ = gz.Close()
		return VerifyContent(&b, module)
	}
	if err := verify(); err != nil {
		t.Fatalf("good tarball rejected: %v", err)
	}
	for name, h := range map[string]*tar.Header{
		"symlink":      {Name: "go/bin/evil", Typeflag: tar.TypeSymlink, Linkname: "/bin/sh"},
		"hard link":    {Name: "go/bin/evil", Typeflag: tar.TypeLink, Linkname: "go/bin/go"},
		"duplicate":    {Name: "go/src/a.go", Typeflag: tar.TypeReg},
		"unclean path": {Name: "go/src/../bin/evil", Typeflag: tar.TypeReg},
		"outside go/":  {Name: "etc/passwd", Typeflag: tar.TypeReg},
	} {
		if err := verify(h); err == nil {
			t.Errorf("%s: accepted", name)
		}
	}
}
