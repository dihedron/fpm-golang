package golang

import "testing"

func TestParseVersion(t *testing.T) {
	ok := map[string]string{"go1.27.4": "1.27.4", "1.26.11": "1.26.11", "go1.27": "1.27"}
	for in, want := range ok {
		v, err := ParseVersion(in)
		if err != nil || v.String() != want {
			t.Errorf("ParseVersion(%q) = %v, %v; want %s", in, v, err, want)
		}
	}
	bad := []string{"", "go", "1", "1.27.4; rm -rf /", "../x", "go1.27rc1", "go1.27.4\n", "1.2.3.4", "go-1.2", "99999.1.1", " 1.2.3"}
	for _, in := range bad {
		if _, err := ParseVersion(in); err == nil {
			t.Errorf("ParseVersion(%q) accepted", in)
		}
	}
}

func TestVersionCompare(t *testing.T) {
	v := func(s string) Version { x, _ := ParseVersion(s); return x }
	if !(v("1.27.4").Compare(v("1.26.11")) > 0) || !(v("1.27.10").Compare(v("1.27.9")) > 0) ||
		!(v("1.27").Compare(v("1.26.11")) > 0) || v("1.27").Compare(v("1.27.0")) != 0 {
		t.Error("bad ordering")
	}
}

func FuzzParseVersion(f *testing.F) {
	for _, s := range []string{"go1.27.4", "1.27", "", "go1.2.3.4", "1.27.4;x"} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, s string) {
		v, err := ParseVersion(s)
		if err != nil {
			return
		}
		// accepted values must be safe to embed in tags, filenames and shell
		for _, r := range v.String() {
			if !(r == '.' || r >= '0' && r <= '9') {
				t.Fatalf("unsafe rune %q in %q", r, v.String())
			}
		}
	})
}
