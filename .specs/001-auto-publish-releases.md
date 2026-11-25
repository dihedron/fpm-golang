# SPEC-001: Automatic publication of Go packages tracking go.dev

Status: DRAFT v2, open decisions resolved (awaiting final approval) — no implementation code may be written until approved.

## 1. Goal

Whenever go.dev publishes a new Go release, this repository automatically repackages the
official Linux tarball as DEB, RPM and APK packages for every supported architecture and
publishes them in a GitHub Release dedicated to that Go version. Package versions track
upstream exactly: `go1.27.4` becomes package version `1.27.4`.

Non-goals (this spec): MSI packages (see §10), other architectures and operating systems
(to be evaluated later), publishing to apt/yum/apk repositories, release candidates and betas.

Decisions taken (from review of v1): only the latest stable version is packaged, no
coexistence of Go lines (`1.27.0` simply replaces `1.26.11`), tags `go<version>`, packaging
revisions allowed, GPG-signed packages, digests cross-checked against the Go checksum
database, amd64 + arm64 only, no PDF guides (README only; deliberate deviation from
AGENTS.md rule 12 for this project).

## 2. Upstream source of truth

`https://go.dev/dl/?mode=json` lists the currently supported stable releases with, for each
file, `filename`, `os`, `arch`, `kind`, `sha256` and `size`. Observed today (2026-11-25): it
lists only `go1.27.4` and `go1.26.11`, because the Go project supports only the two latest
minor versions. **Only the highest stable version in the feed is packaged**; older lines still listed
(e.g. `go1.26.11`) are ignored. Consequence: security point releases of an older line are
not published. The selection rule lives in one function, so widening it later is trivial.

- Only `stable: true` entries are packaged. RCs/betas are out of scope (`?include=all` is not used).
- Per release, the file used is the one with `kind=archive`, `os=linux`, and `arch` in the
  architecture table (§5). Its `sha256` from the feed is the expected digest.

## 3. Design overview

```
cron (every 6h) / manual dispatch
        │
   [plan job]  fpm-golang release plan
        │   fetch feed → list existing GitHub releases → emit JSON matrix of MISSING
        │   (version × arch × format) combinations; empty ⇒ stop
        ▼
   [build job, matrix: version × arch]
        │   download tarball → verify sha256 → nfpm deb/rpm/apk → smoke-test in container
        ▼
   [publish job, one per version]
        │   create DRAFT release go<ver> → upload all packages + SHA256SUMS → attest → publish
```

Idempotency is by state-in-GitHub: a Go version is "done" iff a published release tagged
`go<version>` exists. No database, no state file in the repo.

### 3.1 Components

1. **GitHub Actions workflow** `.github/workflows/track-go.yml` — triggers: `schedule`
   (every 6 hours), `workflow_dispatch` (optional input `version` to force one version,
   e.g. a rebuild). `concurrency: track-go` (no cancel) so runs never overlap.
2. **`cmd/fpm-golang`** (Go, `go-flags`, per AGENTS.md CLI convention) with:
   - `release plan` — prints the JSON build matrix of missing releases (reads the feed,
     lists releases via `gh`/GitHub API; both injectable for tests).
   - `release show` — shows what upstream offers vs what is already published.
   - `version` — compact (`-v`) and verbose (`-vv`) modes per AGENTS.md.
   Packaging itself stays in `make` + nFPM; Go code only does what shell cannot do safely
   (parse, validate, diff). Logging with `log/slog`, `context.Context` everywhere.
3. **Makefile / nfpm.yaml** — parameterised by `VERSION` and `ARCH`; sha256 verification
   added; nFPM pinned to an explicit version instead of `@latest`.

## 4. Naming and versioning

| Item | Value |
|---|---|
| Git tag / release name | `go1.27.4` (distinct from this tool's own `vX.Y.Z` tags, which `make release-*` keeps using) |
| Package version | `1.27.4` (epoch 1 kept from current `nfpm.yaml`) |
| Asset filename | `google-golang-<arch>-<version>.<ext>`, `<arch>` ∈ `x86-64`, `arm64` (as requested), `<ext>` ∈ `deb`, `rpm`, `apk` |
| Package name inside the package | `google-golang` |
| Native arch fields | deb `amd64`/`arm64`, rpm `x86_64`/`aarch64`, apk `x86_64`/`aarch64` (nFPM maps these from one `arch` value) |
| Extra release assets | `SHA256SUMS` (+ detached `.asc`), public key `google-golang.gpg.pub`, build provenance attestation |

**Packaging revisions.** Revision 1 (the automatic one) has no suffix. A packaging-only fix
is released by `workflow_dispatch` with inputs `version=1.27.4` and `revision=2`: tag
`go1.27.4-2`, package version `1.27.4-2` (deb/apk `-2`, rpm `Release: 2`), asset names
`google-golang-x86-64-1.27.4-2.deb`. Revisions are manual only; the cron never creates one.
The plan job treats `go1.27.4` as "done" regardless of revisions.

**Single line.** One package name, installs into `/opt/go`; a newer version replaces the
older through the package manager. No `Conflicts`/per-line names.

## 4.1 Signing

Packages are signed with a dedicated GPG key through nFPM's native signing (deb, rpm, apk).
The private key and passphrase are GitHub secrets scoped to a protected `release`
environment; the public key is committed in the repo and attached to every release.
Signing happens in the build job (nFPM signs while packaging), so that job is the only one
besides publish with the environment. `SHA256SUMS` is signed detached (`.asc`).
Key generation, expiry (proposed: 2 years) and rotation are covered in README.

Implementation findings (amend v2): (a) apk packages cannot be signed with GPG: apk uses an
RSA PEM key, so there are two keys, GPG (deb, rpm, `SHA256SUMS`) and RSA (apk), both secrets
in the `release` environment and both public halves committed in `keys/`. (b) dpkg does not
verify the embedded deb signature (`_gpgorigin`); consumers verify it with `gpg`, documented
in the README. (c) nFPM must be pinned at >= v2.49.0: older builds emit apks that apk-tools
rejects (`package file format error`), including the previous Makefile's output. (d) The
old nfpm.yaml packaged the whole `/etc/profile.d` directory (conflicts with distro
packages) and declared `provides: google-golang` equal to its own name (apk self-conflict);
both removed.

## 5. Architecture / format matrix

Initial: `amd64`, `arm64` × `deb`, `rpm`, `apk` = 6 assets per Go version. The matrix is a
table in the Go source (arch → feed arch, nFPM arch, filename arch), so adding `riscv64`,
`ppc64le`, `s390x`, etc. is a one-line change. Formats are a second list so that `msi` can
later be added as a format with its own builder (§10).

## 6. Threat model

Assets: integrity of published packages (they are installed as root, into `/opt/go`, and a
profile script is dropped in `/etc/profile.d`); the repo's write token; availability of the pipeline.

| # | Threat | Mitigation |
|---|---|---|
| T1 | Tarball tampered in transit or at the CDN | HTTPS only; sha256 from the feed must match; mismatch ⇒ hard fail, nothing published |
| T2 | Feed and tarball compromised together (same origin, so T1 check is not independent) | Mandatory second check, content-based (the sumdb stores the h1 hash of the `golang.org/toolchain` module zip, not the tarball digest; spike on go1.27.4 showed the zip holds 12072 of the tarball's 15659 files, all of `bin/`, `pkg/`, `lib/`, `src/`, byte-identical except `go.mod` stored as `_go.mod`; `test/`, `api/`, `misc/` are not in the module). The build job runs `go mod download` (which verifies the h1 hash against `sum.golang.org`) and requires every module file to be byte-identical to the tarball file, with `bin/` and `pkg/` fully covered; any difference ⇒ hard fail. Known gap: files outside the module are only covered by T1. Published `SHA256SUMS` and attestation let consumers verify what we built |
| T3 | Malicious/malformed version string from the feed flows into shell, tag name, matrix | Version must match `^[0-9]+\.[0-9]+(\.[0-9]+)?$` in Go before use; values passed to steps via env vars, never interpolated into `run:` scripts; filenames are built by us, not taken from the feed |
| T4 | Compromised third-party action or tool | Actions pinned to commit SHA; nFPM pinned to a version and verified by checksum; `govulncheck` and secret scanning in CI (AGENTS.md rules 6–7) |
| T5 | Over-privileged token | Workflow default `permissions: {}`; only the publish job gets `contents: write` (+ `id-token`/`attestations: write`); plan/build jobs read-only |
| T6 | Existing release overwritten or silently altered | Existing tags are never modified; publish refuses if tag exists; draft-then-publish; `--clobber` never used |
| T7 | Partial release (some assets missing) visible to users | Release created as draft, published only after all expected assets uploaded and counted |
| T8 | Unattended-run abuse: workflow spam or cost | Fixed cron, single concurrency group, max N versions per run (default 4), job timeouts |
| T9 | Package scripts run arbitrary code as root | Packages ship no maintainer scripts; contents are only the upstream tree plus `etc/google-golang.sh` (the PATH handling in the profile script is reviewed in this spec's audit) |
| T10 | Upstream retracts a release | Not auto-deleted; operator decision (documented in Operator's Guide) |
| T11 | Scheduled workflow auto-disabled after 60 days of repo inactivity | Documented; the plan job's own releases do not count as activity, so a weekly health check or a keep-alive is listed in Q6 |
| T12 | Package signing absent ⇒ no trust anchor | GPG signatures (§4.1), checksums, attestation |
| T13 | Signing key leaked or abused by a modified workflow | Key only in a protected `release` environment (branch-restricted to `main`, optional required reviewer); passphrase-protected; never written to logs; key imported to a tmpfs path and removed at job end; the build job runs only pinned tools on a digest-verified tarball; documented revocation/rotation procedure |

## 7. Acceptance criteria (tests are written before implementation)

Unit (Go, offline via `httptest` and a fake release lister):
- AC1 Feed with `go1.27.4` and `go1.26.11`, no releases ⇒ plan lists both, 2 arch × 3 formats each.
- AC2 Release `go1.27.4` already published ⇒ plan lists only `go1.26.11`.
- AC3 Everything published ⇒ plan is empty, exit 0.
- AC4 Draft/unpublished release with the same tag counts as NOT done (resumed, not duplicated).
- AC5 `stable: false` entries are ignored.
- AC6 Version strings failing the regex (`1.27.4; rm -rf`, `../x`, empty) are rejected with an error; fuzz test on the parser.
- AC7 Feed missing the linux archive for an arch ⇒ that version fails the plan loudly (non-zero exit), others still emitted? **No**: fail the whole plan, never publish a partial arch set.
- AC8 Feed HTTP error, timeout or invalid JSON ⇒ non-zero exit, nothing published.
- AC9 Filename builder yields exactly `google-golang-x86-64-1.27.4.deb`, and `…-1.27.4-2.deb` for revision 2.
- AC14 Plan with feed `[1.27.4, 1.26.11]` selects only `1.27.4`.
- AC15 Plan treats `go1.27.4` as done even if only a `-2` revision release exists; automatic runs never emit revision ≥ 2; manual `revision=2` emits tag `go1.27.4-2` and refuses if it exists.
- AC16 `release verify`: a tarball where any file differs from the sumdb-verified module (e.g. one byte flipped in `bin/go`), or where a `bin/`/`pkg/` file is missing from the module or tarball ⇒ non-zero exit, nothing built.
- AC17 Each produced package verifies against the committed public key (`dpkg-sig`/`rpm -K`/`apk verify`) in the smoke-test containers; a tampered package fails verification.

Integration / lab:
- AC10 `make deb|rpm|apk VERSION=… ARCH=…` verifies sha256 and fails on a corrupted tarball.
- AC11 Smoke test per format in a container (Debian/Ubuntu, AlmaLinux, Alpine; arm64 via QEMU): install the package, `go version` reports the packaged version, uninstall removes `/opt/go` and the profile script.
- AC12 Dry-run mode (`workflow_dispatch` with `dry_run=true`) runs plan + build + smoke tests but creates no release; used to validate the pipeline before the first real run.
- AC13 Publish job re-run after partial failure completes the release without duplicating assets.

Benchmarks: not warranted (I/O-bound). Fuzzing: version parser and feed decoder (AC6).
Chaos/long soak: none proposed; the pipeline is stateless and run-to-completion. A 2-day
observation of the real cron after enabling is the soak (no overnight session needed).

## 8. Documentation (AGENTS.md rules 11–12)

Only `README.md` is updated: how to install the packages (deb/rpm/apk, importing the
public key, verifying signatures), how to build them with this tool (`make`, `ARCH`,
`VERSION`), how the workflow works (cron, dispatch, dry run, revisions), key rotation.
No PDF guides: explicit owner decision, deviating from rule 12.

## 9. Backward compatibility (AGENTS.md rule 8)

Product is pre-1.0, so: output filenames change from nFPM's default
(`google-golang_1.27.4_amd64.deb`) to the §4 scheme; `make` targets gain `ARCH`; the
`VERSION=1.24rc2` override keeps working locally. No compatibility shims.

## 10. Later: MSI

Out of scope. The format list in §5 is the extension point. Tooling candidates: WiX v4 on
a `windows-latest` runner (most standard) or `msitools` (`wixl`) on Linux. Needs its own
spec: install location, PATH/GOROOT handling, code signing (unsigned MSIs trigger SmartScreen).

## 11. Remaining open items

- **Q6 Cron keep-alive.** Scheduled workflows are disabled after 60 days without repo
  activity. Proposed: accept the risk and document it (a manual dispatch re-enables it);
  alternative is a keep-alive commit/API call. Default: accept.
- **Q10 Signing key.** Who generates the GPG key and when (needed before first real run;
  dry run works without it)? Default: owner generates it, adds secrets, I document the procedure.
