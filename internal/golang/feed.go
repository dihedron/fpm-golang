package golang

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// FeedURL is the upstream list of supported stable releases.
const FeedURL = "https://go.dev/dl/?mode=json"

// Arch maps the upstream architecture name to packaging names.
type Arch struct {
	Feed string // go.dev arch, also nFPM arch
	Name string // arch part of our asset filenames
}

// Arches lists the supported architectures; add a line to support a new one.
var Arches = []Arch{{"amd64", "x86-64"}, {"arm64", "arm64"}}

// Formats lists the package formats built for every architecture.
var Formats = []string{"deb", "rpm", "apk"}

// File is one upstream linux archive.
type File struct{ Filename, SHA256 string }

// Release is an upstream stable release with its linux archives keyed by feed arch.
type Release struct {
	Version Version
	Files   map[string]File
}

type feedRelease struct {
	Version string `json:"version"`
	Stable  bool   `json:"stable"`
	Files   []struct {
		Filename string `json:"filename"`
		OS       string `json:"os"`
		Arch     string `json:"arch"`
		SHA256   string `json:"sha256"`
		Kind     string `json:"kind"`
	} `json:"files"`
}

// FetchLatest returns the highest stable release in the feed; older lines are ignored.
func FetchLatest(ctx context.Context, c *http.Client, url string) (Release, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return Release{}, fmt.Errorf("building feed request: %w", err)
	}
	resp, err := c.Do(req)
	if err != nil {
		return Release{}, fmt.Errorf("fetching feed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return Release{}, fmt.Errorf("fetching feed: HTTP %d", resp.StatusCode)
	}
	var all []feedRelease
	if err := json.NewDecoder(io.LimitReader(resp.Body, 64<<20)).Decode(&all); err != nil {
		return Release{}, fmt.Errorf("decoding feed: %w", err)
	}
	var best *feedRelease
	var bestV Version
	for i := range all {
		if !all[i].Stable {
			continue
		}
		v, err := ParseVersion(all[i].Version)
		if err != nil {
			return Release{}, fmt.Errorf("feed: %w", err)
		}
		if best == nil || v.Compare(bestV) > 0 {
			best, bestV = &all[i], v
		}
	}
	if best == nil {
		return Release{}, fmt.Errorf("feed lists no stable release")
	}
	rel := Release{Version: bestV, Files: map[string]File{}}
	for _, f := range best.Files {
		if f.OS == "linux" && f.Kind == "archive" {
			rel.Files[f.Arch] = File{Filename: f.Filename, SHA256: f.SHA256}
		}
	}
	for _, a := range Arches {
		f, ok := rel.Files[a.Feed]
		if !ok || f.SHA256 == "" {
			return Release{}, fmt.Errorf("feed: go%s has no linux/%s archive", bestV, a.Feed)
		}
		// we build every name ourselves; only accept the one upstream naming scheme
		if f.Filename != fmt.Sprintf("go%s.linux-%s.tar.gz", bestV, a.Feed) {
			return Release{}, fmt.Errorf("feed: unexpected filename %q", f.Filename)
		}
	}
	return rel, nil
}
