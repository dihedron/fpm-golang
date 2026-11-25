package golang

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// Entry is one build-matrix row: all packages for one architecture.
type Entry struct {
	Arch     string   `json:"arch"`      // feed/nFPM arch, e.g. amd64
	Name     string   `json:"arch_name"` // filename arch, e.g. x86-64
	Tarball  string   `json:"tarball"`
	SHA256   string   `json:"sha256"`
	Packages []string `json:"packages"`
}

// Plan is what the workflow must build; empty Include means nothing to do.
type Plan struct {
	Version  string  `json:"version"`
	Revision int     `json:"revision"`
	Tag      string  `json:"tag"`
	Include  []Entry `json:"include"`
}

// PackageName builds an asset filename; revision 1 carries no suffix.
func PackageName(arch, version string, revision int, format string) string {
	if revision > 1 {
		return fmt.Sprintf("google-golang-%s-%s-%d.%s", arch, version, revision, format)
	}
	return fmt.Sprintf("google-golang-%s-%s.%s", arch, version, format)
}

// NewPlan computes the work for rel given the tags of published releases.
// revision 0 means automatic (revision 1, only if no release of that version exists);
// revision >= 2 is a manual packaging re-release whose tag must not exist yet.
func NewPlan(rel Release, published []string, revision int) (Plan, error) {
	if revision < 0 {
		return Plan{}, fmt.Errorf("invalid revision %d", revision)
	}
	v := rel.Version.String()
	base := "go" + v
	p := Plan{Version: v, Revision: 1, Tag: base}
	if revision >= 2 {
		p.Revision, p.Tag = revision, fmt.Sprintf("%s-%d", base, revision)
	}
	for _, t := range published {
		if t == p.Tag {
			if revision >= 2 {
				return Plan{}, fmt.Errorf("release %s already exists", t)
			}
			return p, nil
		}
		if revision < 2 && strings.HasPrefix(t, base+"-") { // a revision implies base was done
			return p, nil
		}
	}
	for _, a := range Arches {
		f := rel.Files[a.Feed]
		e := Entry{Arch: a.Feed, Name: a.Name, Tarball: f.Filename, SHA256: f.SHA256}
		for _, fm := range Formats {
			e.Packages = append(e.Packages, PackageName(a.Name, v, p.Revision, fm))
		}
		p.Include = append(p.Include, e)
	}
	return p, nil
}

// PublishedTags lists the tags of non-draft releases of the current repository via gh.
func PublishedTags(ctx context.Context) ([]string, error) {
	out, err := exec.CommandContext(ctx, "gh", "release", "list", "--limit", "500", "--json", "tagName,isDraft").Output()
	if err != nil {
		return nil, fmt.Errorf("listing releases with gh: %w", err)
	}
	return parseGHReleases(out)
}

func parseGHReleases(b []byte) ([]string, error) {
	var rs []struct {
		TagName string `json:"tagName"`
		IsDraft bool   `json:"isDraft"`
	}
	if err := json.Unmarshal(b, &rs); err != nil {
		return nil, fmt.Errorf("decoding gh output: %w", err)
	}
	var tags []string
	for _, r := range rs {
		if !r.IsDraft {
			tags = append(tags, r.TagName)
		}
	}
	return tags, nil
}
