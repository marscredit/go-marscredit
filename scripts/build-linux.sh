#!/usr/bin/env bash
# Build reproducible Linux geth binaries for Mars Credit (1.10.18-mars-fast1).
#
# Usage:
#   ./scripts/build-linux.sh           # both amd64 and arm64 + SHA256SUMS
#   ./scripts/build-linux.sh amd64     # linux-amd64 only
#   ./scripts/build-linux.sh arm64     # linux-arm64 only
#
# Cross-builds use CGO_ENABLED=0 (pure-Go secp256k1), matching MARS_FAST.md.
# Pin Go 1.17.x in CI; a newer local toolchain usually works for this tree.

set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

ARCHES="${*:-amd64 arm64}"
for a in $ARCHES; do
	case "$a" in
	amd64|arm64) ;;
	*)
		echo "usage: $0 [amd64] [arm64]" >&2
		exit 2
		;;
	esac
done

COMMIT="$(git rev-parse HEAD)"
DATE="$(git show -s --format=%cs HEAD | tr -d '-')"
VERSION_META="1.10.18-mars-fast1"
OUTDIR="${ROOT}/dist"
mkdir -p "$OUTDIR"

LDFLAGS="-s -w -X main.gitCommit=${COMMIT} -X main.gitDate=${DATE}"

echo "Mars Credit geth ${VERSION_META}"
echo "  commit  ${COMMIT}"
echo "  date    ${DATE}"
echo "  arches  ${ARCHES}"
echo

build_one() {
	local arch="$1"
	local name="geth-linux-${arch}"
	local out="${OUTDIR}/${name}"
	echo ">> GOOS=linux GOARCH=${arch} CGO_ENABLED=0 go build ./cmd/geth"
	GOOS=linux GOARCH="${arch}" CGO_ENABLED=0 GO111MODULE=on \
		go build -trimpath -ldflags "${LDFLAGS}" -o "${out}" ./cmd/geth
	chmod +x "${out}"
	echo "   wrote ${out} ($(wc -c < "${out}") bytes)"
}

for arch in $ARCHES; do
	build_one "$arch"
done

# Native binary: confirm it is the mars-fast1 fork, not stock geth.
HOST_ARCH="$(uname -m)"
NATIVE=""
case "$HOST_ARCH" in
x86_64|amd64) NATIVE="amd64" ;;
aarch64|arm64) NATIVE="arm64" ;;
esac
if [ -n "$NATIVE" ] && [ -x "${OUTDIR}/geth-linux-${NATIVE}" ]; then
	echo
	echo ">> ${OUTDIR}/geth-linux-${NATIVE} version"
	"${OUTDIR}/geth-linux-${NATIVE}" version
	"${OUTDIR}/geth-linux-${NATIVE}" version | grep -q "1.10.18-mars-fast1"
fi

echo
(
	cd "$OUTDIR"
	# Checksums of whatever linux binaries are in dist/ (this arch plus any
	# sibling from a previous run). Skip missing files so a single-arch
	# build does not fail under set -o pipefail.
	files=()
	for f in geth-linux-amd64 geth-linux-arm64; do
		[ -f "$f" ] && files+=("$f")
	done
	if [ "${#files[@]}" -eq 0 ]; then
		echo "no geth-linux-* binaries in ${OUTDIR}" >&2
		exit 1
	fi
	sha256sum "${files[@]}" | tee SHA256SUMS
)

echo
echo "Done. Binaries and SHA256SUMS are in ${OUTDIR}/"
