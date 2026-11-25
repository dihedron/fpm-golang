package golang

import "testing"

func release(t *testing.T, v string) Release {
	t.Helper()
	ver, err := ParseVersion(v)
	if err != nil {
		t.Fatal(err)
	}
	return Release{Version: ver, Files: map[string]File{
		"amd64": {Filename: "go" + v + ".linux-amd64.tar.gz", SHA256: "a1"},
		"arm64": {Filename: "go" + v + ".linux-arm64.tar.gz", SHA256: "a2"},
	}}
}

func TestPlanNothingPublished(t *testing.T) { // AC1
	p, err := NewPlan(release(t, "1.27.4"), nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	if p.Tag != "go1.27.4" || len(p.Include) != 2 {
		t.Fatalf("plan %+v", p)
	}
	want := []string{"google-golang-x86-64-1.27.4.deb", "google-golang-x86-64-1.27.4.rpm", "google-golang-x86-64-1.27.4.apk"}
	for i, f := range p.Include[0].Packages { // AC9
		if f != want[i] {
			t.Errorf("package %d = %s, want %s", i, f, want[i])
		}
	}
	if p.Include[1].Arch != "arm64" || p.Include[1].Packages[0] != "google-golang-arm64-1.27.4.deb" || p.Include[1].SHA256 != "a2" {
		t.Errorf("bad arm64 entry %+v", p.Include[1])
	}
}

func TestPlanAlreadyPublished(t *testing.T) { // AC3
	for _, pub := range [][]string{{"go1.27.4"}, {"go1.27.4-2"}} { // AC15
		p, err := NewPlan(release(t, "1.27.4"), pub, 0)
		if err != nil || len(p.Include) != 0 {
			t.Errorf("published %v: plan %+v err %v", pub, p, err)
		}
	}
}

func TestPlanOlderPublishedDoesNotCount(t *testing.T) { // AC2, AC4 (drafts are never passed in)
	p, err := NewPlan(release(t, "1.27.4"), []string{"go1.26.11", "v0.1.0"}, 0)
	if err != nil || len(p.Include) != 2 {
		t.Errorf("plan %+v err %v", p, err)
	}
}

func TestPlanRevision(t *testing.T) { // AC15
	p, err := NewPlan(release(t, "1.27.4"), []string{"go1.27.4"}, 2)
	if err != nil || p.Tag != "go1.27.4-2" || p.Include[0].Packages[0] != "google-golang-x86-64-1.27.4-2.deb" || p.Revision != 2 {
		t.Errorf("plan %+v err %v", p, err)
	}
	if _, err := NewPlan(release(t, "1.27.4"), []string{"go1.27.4", "go1.27.4-2"}, 2); err == nil {
		t.Error("existing revision accepted")
	}
	if _, err := NewPlan(release(t, "1.27.4"), nil, -1); err == nil {
		t.Error("negative revision accepted")
	}
}

func TestParseGHReleases(t *testing.T) {
	tags, err := parseGHReleases([]byte(`[{"tagName":"go1.27.4","isDraft":false},{"tagName":"go1.27.5","isDraft":true}]`))
	if err != nil || len(tags) != 1 || tags[0] != "go1.27.4" {
		t.Errorf("tags %v err %v", tags, err)
	}
	if _, err := parseGHReleases([]byte("nope")); err == nil {
		t.Error("expected error")
	}
}
