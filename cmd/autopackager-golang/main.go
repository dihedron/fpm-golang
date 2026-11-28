// Command autopackager-golang plans, verifies and describes the repackaging of upstream Go releases.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/dihedron/autopackager-golang/internal/golang"
	"github.com/dihedron/autopackager-golang/pkg/metadata"
	flags "github.com/jessevdk/go-flags"
)

var httpClient = &http.Client{Timeout: 60 * time.Second}

type options struct {
	Release releaseCmd `command:"release" description:"Upstream Go releases"`
	Version versionCmd `command:"version" description:"Show version information"`
}

type releaseCmd struct {
	Plan   planCmd   `command:"plan" description:"Print the JSON build matrix of what is missing"`
	Show   showCmd   `command:"show" description:"Show the latest upstream release and what is published"`
	Verify verifyCmd `command:"verify" description:"Verify a downloaded tarball"`
}

type versionCmd struct {
	Verbose bool `short:"v" long:"verbose" description:"Show detailed build information"`
}

func (c *versionCmd) Execute([]string) error {
	if c.Verbose {
		fmt.Print(metadata.Long())
	} else {
		fmt.Println(metadata.Short())
	}
	return nil
}

type planCmd struct {
	Version  string `long:"version" description:"Must equal the latest stable version; required with --revision"`
	Revision int    `long:"revision" description:"Manual packaging revision (>= 2) of the latest version"`
}

func (c *planCmd) Execute([]string) error {
	ctx := context.Background()
	rel, err := golang.FetchLatest(ctx, httpClient, golang.FeedURL)
	if err != nil {
		return err
	}
	if c.Revision != 0 && c.Version == "" {
		return errors.New("--revision requires --version")
	}
	if c.Version != "" {
		v, err := golang.ParseVersion(c.Version)
		if err != nil {
			return err
		}
		if v.Compare(rel.Version) != 0 {
			return fmt.Errorf("version %s is not the latest stable (%s)", v, rel.Version)
		}
	}
	tags, err := golang.PublishedTags(ctx)
	if err != nil {
		return err
	}
	p, err := golang.NewPlan(rel, tags, c.Revision)
	if err != nil {
		return err
	}
	slog.Info("plan computed", "version", p.Version, "tag", p.Tag, "builds", len(p.Include))
	return json.NewEncoder(os.Stdout).Encode(p)
}

type showCmd struct{}

func (c *showCmd) Execute([]string) error {
	ctx := context.Background()
	rel, err := golang.FetchLatest(ctx, httpClient, golang.FeedURL)
	if err != nil {
		return err
	}
	tags, err := golang.PublishedTags(ctx)
	if err != nil {
		return err
	}
	p, err := golang.NewPlan(rel, tags, 0)
	if err != nil {
		return err
	}
	fmt.Printf("latest upstream: go%s\npublished tags:  %v\nto build:        %d architecture(s)\n", rel.Version, tags, len(p.Include))
	return nil
}

type verifyCmd struct {
	Tarball string `long:"tarball" required:"true" description:"Downloaded go<ver>.linux-<arch>.tar.gz"`
	SHA256  string `long:"sha256" description:"Expected digest (default: from the upstream feed)"`
	Module  string `long:"module" description:"Extracted golang.org/toolchain module dir (from go mod download) to cross-check content"`
}

func (c *verifyCmd) Execute([]string) error {
	ctx := context.Background()
	want := c.SHA256
	if want == "" {
		rel, err := golang.FetchLatest(ctx, httpClient, golang.FeedURL)
		if err != nil {
			return err
		}
		for _, f := range rel.Files {
			if f.Filename == filepath.Base(c.Tarball) {
				want = f.SHA256
			}
		}
		if want == "" {
			return fmt.Errorf("%s is not a latest-release archive; pass --sha256", filepath.Base(c.Tarball))
		}
	}
	f, err := os.Open(c.Tarball)
	if err != nil {
		return fmt.Errorf("opening tarball: %w", err)
	}
	defer f.Close()
	if err := golang.VerifySHA256(f, want); err != nil {
		return err
	}
	slog.Info("sha256 ok", "tarball", c.Tarball)
	if c.Module == "" {
		return nil
	}
	if _, err := f.Seek(0, 0); err != nil {
		return fmt.Errorf("rewinding tarball: %w", err)
	}
	if err := golang.VerifyContent(f, c.Module); err != nil {
		return err
	}
	slog.Info("content matches the checksum-database-verified module", "module", c.Module)
	return nil
}

func main() {
	var o options
	if _, err := flags.NewParser(&o, flags.Default).Parse(); err != nil {
		var fe *flags.Error
		if errors.As(err, &fe) && fe.Type == flags.ErrHelp {
			return
		}
		slog.Error("command failed", "error", err)
		os.Exit(1)
	}
}
