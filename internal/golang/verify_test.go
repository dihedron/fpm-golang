package golang

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"
)

type fl map[string]string

func mkTar(t *testing.T, files fl) []byte {
	t.Helper()
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	tw := tar.NewWriter(gz)
	for n, c := range files {
		if err := tw.WriteHeader(&tar.Header{Name: "go/" + n, Mode: 0o644, Size: int64(len(c)), Typeflag: tar.TypeReg}); err != nil {
			t.Fatal(err)
		}
		_, _ = tw.Write([]byte(c))
	}
	_ = tw.Close()
	_ = gz.Close()
	return b.Bytes()
}

func mkModule(t *testing.T, files fl) string {
	t.Helper()
	d := t.TempDir()
	for n, c := range files {
		p := filepath.Join(d, n)
		_ = os.MkdirAll(filepath.Dir(p), 0o755)
		if err := os.WriteFile(p, []byte(c), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return d
}

func TestVerifyContent(t *testing.T) { // AC16
	mod := fl{"bin/go": "BIN", "pkg/tool": "T", "src/a.go": "A", "src/_go.mod": "M"}
	tarball := fl{"bin/go": "BIN", "pkg/tool": "T", "src/a.go": "A", "src/go.mod": "M", "test/x.go": "only in tarball"}
	if err := VerifyContent(bytes.NewReader(mkTar(t, tarball)), mkModule(t, mod)); err != nil {
		t.Fatalf("good tarball rejected: %v", err)
	}
	cases := map[string]fl{
		"flipped byte":        {"bin/go": "BIX", "pkg/tool": "T", "src/a.go": "A", "src/go.mod": "M"},
		"module file missing": {"bin/go": "BIN", "pkg/tool": "T", "src/go.mod": "M"},
		"extra bin file":      {"bin/go": "BIN", "bin/evil": "E", "pkg/tool": "T", "src/a.go": "A", "src/go.mod": "M"},
		"extra pkg file":      {"bin/go": "BIN", "pkg/tool": "T", "pkg/evil": "E", "src/a.go": "A", "src/go.mod": "M"},
		"go.mod changed":      {"bin/go": "BIN", "pkg/tool": "T", "src/a.go": "A", "src/go.mod": "X"},
	}
	for name, tb := range cases {
		if err := VerifyContent(bytes.NewReader(mkTar(t, tb)), mkModule(t, mod)); err == nil {
			t.Errorf("%s: accepted", name)
		}
	}
	if err := VerifyContent(bytes.NewReader([]byte("not gzip")), mkModule(t, mod)); err == nil {
		t.Error("garbage tarball accepted")
	}
}

func TestVerifySHA256(t *testing.T) { // AC10
	// sha256("abc")
	const abc = "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"
	if err := VerifySHA256(bytes.NewReader([]byte("abc")), abc); err != nil {
		t.Error(err)
	}
	if err := VerifySHA256(bytes.NewReader([]byte("abd")), abc); err == nil {
		t.Error("mismatch accepted")
	}
}
