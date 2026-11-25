#!/bin/sh
# Create throwaway signing keys for dry runs and tests (never used for real releases).
# Usage: ephemeral-keys.sh <dir>; prints nothing, writes into <dir>:
#   gpg.key (armored, passphrase "ephemeral"), google-golang.gpg.pub, apk.key, google-golang.rsa.pub
set -eu
d=$1; mkdir -p "$d"; export GNUPGHOME=$(mktemp -d)
gpg --batch --passphrase ephemeral --quick-gen-key "ephemeral <ephemeral@invalid>" rsa3072 sign never 2>/dev/null
gpg --batch --pinentry-mode loopback --passphrase ephemeral --export-secret-keys --armor > "$d/gpg.key"
gpg --export --armor > "$d/google-golang.gpg.pub"
openssl genrsa -out "$d/apk.key" 3072 2>/dev/null
openssl rsa -in "$d/apk.key" -pubout -out "$d/google-golang.rsa.pub" 2>/dev/null
rm -rf "$GNUPGHOME"
