#!/bin/bash
# Release the tool: scripts/release.sh major|minor|patch  (DRY_RUN=1 only prints the next version).
# Tags vX.Y.Z on origin/main through a GitHub release; go<version> tags are not ours to bump.
set -euo pipefail
kind=${1:-}
case "$kind" in major | minor | patch) ;; *) echo "usage: $0 major|minor|patch" >&2; exit 2 ;; esac
die() { echo "release: $*" >&2; exit 1; }

[ "$(git rev-parse --abbrev-ref HEAD)" = main ] || die "not on main"
[ -z "$(git status --porcelain)" ] || die "working tree not clean"
git fetch -q origin main --tags
[ "$(git rev-parse HEAD)" = "$(git rev-parse origin/main)" ] || die "main differs from origin/main (push or pull first)"

last=$(git tag --list 'v[0-9]*.[0-9]*.[0-9]*' --sort=-v:refname | head -n 1)
IFS=. read -r major minor patch <<<"${last:-v0.0.0}"
major=${major#v}
case "$kind" in
major) major=$((major + 1)) minor=0 patch=0 ;;
minor) minor=$((minor + 1)) patch=0 ;;
patch) patch=$((patch + 1)) ;;
esac
next="v$major.$minor.$patch"
echo "release: ${last:-none} -> $next"
[ -z "${DRY_RUN:-}" ] || exit 0

go vet ./... && go test ./...
command -v govulncheck >/dev/null || die "govulncheck not installed"
govulncheck ./...
command -v gitleaks >/dev/null || die "gitleaks not installed"
gitleaks detect --no-banner

gh release create "$next" --target main --latest --title "autopackager-golang $next" --generate-notes
