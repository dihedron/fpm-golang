package golang

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// VerifySHA256 checks that r hashes to the expected hex digest.
func VerifySHA256(r io.Reader, want string) error {
	h := sha256.New()
	if _, err := io.Copy(h, r); err != nil {
		return fmt.Errorf("hashing: %w", err)
	}
	if got := hex.EncodeToString(h.Sum(nil)); got != want {
		return fmt.Errorf("sha256 mismatch: got %s, want %s", got, want)
	}
	return nil
}

// VerifyContent cross-checks a Go tarball against the golang.org/toolchain module
// directory, whose h1 hash the checksum database has vouched for (via go mod download).
// Every module file must exist byte-identical in the tarball (the module stores
// go.mod as _go.mod), bin/ and pkg/ must not contain any file the module lacks, and
// the tarball may only hold plain files and directories, each once, below go/.
func VerifyContent(tgz io.Reader, moduleDir string) error {
	want := map[string]string{} // path in tarball -> sha256
	err := filepath.WalkDir(moduleDir, func(p string, d fs.DirEntry, err error) error {
		if err != nil || !d.Type().IsRegular() {
			return err
		}
		rel, _ := filepath.Rel(moduleDir, p)
		rel = filepath.ToSlash(rel)
		if filepath.Base(rel) == "_go.mod" {
			rel = strings.TrimSuffix(rel, "_go.mod") + "go.mod"
		}
		f, err := os.Open(p)
		if err != nil {
			return err
		}
		defer f.Close()
		h := sha256.New()
		if _, err := io.Copy(h, f); err != nil {
			return err
		}
		want[rel] = hex.EncodeToString(h.Sum(nil))
		return nil
	})
	if err != nil {
		return fmt.Errorf("reading module: %w", err)
	}
	if len(want) == 0 {
		return fmt.Errorf("module directory %s holds no files", moduleDir)
	}
	gz, err := gzip.NewReader(tgz)
	if err != nil {
		return fmt.Errorf("reading tarball: %w", err)
	}
	tr := tar.NewReader(gz)
	seen := map[string]bool{}
	for {
		h, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("reading tarball: %w", err)
		}
		name := strings.TrimSuffix(h.Name, "/")
		rel, below := strings.CutPrefix(name, "go/")
		if name != "go" && (!below || path.Clean(name) != name) {
			return fmt.Errorf("unexpected entry %q", h.Name)
		}
		if h.Typeflag == tar.TypeDir {
			continue
		}
		if h.Typeflag != tar.TypeReg || !below {
			return fmt.Errorf("%q is not a plain file", h.Name)
		}
		if seen[rel] {
			return fmt.Errorf("%q appears twice", h.Name)
		}
		seen[rel] = true
		w, inModule := want[rel]
		if !inModule {
			if strings.HasPrefix(rel, "bin/") || strings.HasPrefix(rel, "pkg/") {
				return fmt.Errorf("%s is in the tarball but not in the verified module", rel)
			}
			continue
		}
		hh := sha256.New()
		if _, err := io.Copy(hh, tr); err != nil {
			return fmt.Errorf("reading tarball: %w", err)
		}
		if hex.EncodeToString(hh.Sum(nil)) != w {
			return fmt.Errorf("%s differs from the verified module", rel)
		}
		delete(want, rel)
	}
	if len(want) > 0 {
		for rel := range want {
			return fmt.Errorf("%s (and %d more) is in the verified module but not in the tarball", rel, len(want)-1)
		}
	}
	return nil
}
