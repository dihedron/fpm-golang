#!/bin/sh
# Smoke test of one signed package, run INSIDE a container of the target distro.
# Usage: smoke.sh <package> <expected-go-version> <keydir>
# <keydir> holds google-golang.gpg.pub (deb, rpm) and google-golang.rsa.pub (apk).
# Checks: a tampered copy is rejected, the genuine package verifies and installs,
# `go version` reports the packaged version, removal leaves nothing behind.
set -eu
pkg=$1 want=$2 keys=$3
tmp=$(mktemp -d)
bad="$tmp/tampered-$(basename "$pkg")"
cp "$pkg" "$bad"
# flip one byte in the middle of the file
size=$(wc -c < "$bad")
printf 'X' | dd of="$bad" bs=1 seek=$((size / 2)) conv=notrunc 2>/dev/null

die() { echo "SMOKE FAIL: $*" >&2; exit 1; }

case "$pkg" in
*.deb)
	apt-get update -qq >/dev/null && apt-get install -y -qq gnupg binutils >/dev/null
	export GNUPGHOME="$tmp/gnupg"; mkdir -m700 "$GNUPGHOME"
	gpg --batch --import "$keys/google-golang.gpg.pub" 2>/dev/null
	verify() { # dpkg does not check _gpgorigin: verify it over the three members
		ar p "$1" _gpgorigin > "$tmp/sig" &&
		ar p "$1" debian-binary control.tar.gz data.tar.gz | gpg --batch --verify "$tmp/sig" - 2>/dev/null
	}
	install() { dpkg -i "$1" >/dev/null; }
	remove() { dpkg -r google-golang >/dev/null; }
	;;
*.rpm)
	rpm --import "$keys/google-golang.gpg.pub"
	verify() { rpm -K "$1" | grep -q 'digests signatures OK\|pgp.*OK'; }
	install() { rpm -i "$1"; }
	remove() { rpm -e google-golang; }
	;;
*.apk)
	cp "$keys/google-golang.rsa.pub" /etc/apk/keys/
	# a full install into a scratch root: apk checks signature and data hashes while unpacking
	verify() { rm -rf "$tmp/root"; apk add --root "$tmp/root" --keys-dir /etc/apk/keys --initdb --no-scripts --no-progress "$1" >/dev/null 2>&1; }
	install() { apk add --no-progress "$1" >/dev/null; }
	remove() { apk del google-golang >/dev/null; }
	;;
*) die "unknown package type $pkg" ;;
esac

verify "$bad" && die "tampered package was accepted"
verify "$pkg" || die "genuine package failed verification"
install "$pkg"
got=$(/opt/go/bin/go version) || die "go does not run"
case "$got" in *"go$want "*) ;; *) die "got '$got', want go$want" ;; esac
remove
[ ! -e /opt/go ] || die "/opt/go left behind"
[ ! -e /etc/profile.d/google-golang.sh ] || die "profile script left behind"
echo "SMOKE OK: $pkg"
