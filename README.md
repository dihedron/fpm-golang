# fpm-golang

Repackages the official Google Go compiler (the Linux tarballs from https://go.dev/dl/) as signed `.deb`, `.rpm` and `.apk` packages, and publishes them automatically as soon as a new Go version is released upstream.

Design: [`.specs/001-auto-publish-releases.md`](.specs/001-auto-publish-releases.md).

## Installing a package

Every Go version has a [release](../../releases) tagged `go<version>` (e.g. `go1.27.4`) with these assets for `x86-64` and `arm64`:

```
google-golang-<arch>-<version>.deb | .rpm | .apk     the packages (signed)
SHA256SUMS, SHA256SUMS.asc                           checksums and their detached GPG signature
google-golang.gpg.pub                                GPG public key (deb, rpm, SHA256SUMS)
google-golang.rsa.pub                                RSA public key (apk)
```

Only the latest stable Go version is published; installing it replaces any previous one. Go is installed in `/opt/go`, and `/etc/profile.d/google-golang.sh` adds `/opt/go/bin` to the `PATH` of login shells.

Verify the checksums first (all formats):

```bash
$> gpg --import google-golang.gpg.pub
$> gpg --verify SHA256SUMS.asc SHA256SUMS
$> sha256sum --ignore-missing -c SHA256SUMS
```

**Debian / Ubuntu.** `dpkg` does not check the embedded signature; verify it by hand, then install:

```bash
$> ar p google-golang-x86-64-1.27.4.deb _gpgorigin > sig
$> ar p google-golang-x86-64-1.27.4.deb debian-binary control.tar.gz data.tar.gz | gpg --verify sig -
$> sudo dpkg -i google-golang-x86-64-1.27.4.deb
```

**RHEL / Alma / Fedora.**

```bash
$> sudo rpm --import google-golang.gpg.pub
$> rpm -K google-golang-x86-64-1.27.4.rpm
$> sudo dnf install ./google-golang-x86-64-1.27.4.rpm
```

**Alpine.**

```bash
$> sudo cp google-golang.rsa.pub /etc/apk/keys/
$> sudo apk add ./google-golang-x86-64-1.27.4.apk
```

Packaging-only fixes are released as revisions of the same Go version: `go1.27.4-2` ships `google-golang-x86-64-1.27.4-2.deb`, and so on.

## How releases are produced

The workflow [`track-go`](.github/workflows/track-go.yml) runs every 6 hours (and on demand):

1. **plan** - `fpm-golang release plan` reads the upstream feed, picks the latest stable Go version and checks whether a release `go<version>` already exists; if so, nothing happens.
2. **build** - per architecture it downloads the tarball, checks its sha256 against the feed, cross-checks its content against the `golang.org/toolchain` module whose hash is vouched for by the Go checksum database (`sum.golang.org`), builds the three packages with [nFPM](https://nfpm.goreleaser.com/), signs them, and smoke-tests them in Debian, AlmaLinux and Alpine containers (signature, tampering, install, `go version`, removal).
3. **publish** - creates a draft release, uploads all assets plus `SHA256SUMS` (signed) and a build provenance attestation, and only then makes it public.

Manual runs (`Actions` -> `track-go` -> `Run workflow`):

- `dry_run=true`: builds and smoke-tests with throwaway keys and publishes nothing; use it to validate changes.
- `version=1.27.4 revision=2`: publishes a packaging revision of the latest version (tag `go1.27.4-2`).

GitHub disables scheduled workflows after 60 days without repository activity; if releases stop appearing, re-enable the workflow in the Actions tab.

### One-time setup of the signing keys

```bash
# GPG key for deb, rpm and SHA256SUMS (give it a passphrase)
$> gpg --quick-gen-key "google-golang packages <you@example.com>" rsa4096 sign 2y
$> gpg --armor --export-secret-keys <KEYID> > gpg.key
$> gpg --armor --export <KEYID> > keys/google-golang.gpg.pub
# RSA key for apk
$> openssl genrsa -out apk.key 4096
$> openssl rsa -in apk.key -pubout -out keys/google-golang.rsa.pub
```

Commit the two public keys in `keys/`. Create a GitHub environment named `release` (restrict it to the `main` branch) with the secrets `GPG_PRIVATE_KEY` (content of `gpg.key`), `GPG_PASSPHRASE` and `APK_PRIVATE_KEY` (content of `apk.key`), then delete the local private key files. To rotate keys, repeat the procedure, replace the public keys and the secrets, and say so in the next release notes; if a key leaks, revoke it and rotate immediately.

## Building packages locally

Prerequisites: `make`, `go`, `curl`; nFPM is installed at the pinned version into `bin/` automatically.

```bash
$> make               # build bin/fpm-golang
$> make deb           # package the latest Go for amd64 (also: rpm, apk, packages)
$> make packages ARCH=arm64
$> make packages VERSION=1.26.11 SHA256=<digest> REVISION=2
$> make help
```

`VERSION` defaults to the latest stable release; the tarball digest is taken from the upstream feed (latest release only), so pass `SHA256` for any other version. Packages are signed when `NFPM_SIGN_KEY` (GPG key file) and `NFPM_PASSPHRASE` (deb, rpm) and `NFPM_APK_SIGN_KEY` (RSA key, apk) are set, and unsigned otherwise. `make clean` removes downloaded tarballs and packages.

Test the signed packages in containers with `lab/smoke.sh` (see the script header); `lab/ephemeral-keys.sh <dir>` creates throwaway keys for that.

## The `fpm-golang` command

```bash
$> fpm-golang release show            # latest upstream version vs published releases
$> fpm-golang release plan            # JSON build matrix of what is missing
$> fpm-golang release verify --tarball go1.27.4.linux-amd64.tar.gz [--module <toolchain module dir>]
$> fpm-golang version [-v]
```

`release plan` and `release show` need the GitHub CLI (`gh`) authenticated for the repository.

## Development

```bash
$> go test ./...
$> go test ./internal/golang -fuzz FuzzParseVersion
```
