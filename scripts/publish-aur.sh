#!/usr/bin/env bash
set -euo pipefail

usage() {
    echo "Usage: $0 [--dry-run]"
    echo ""
    echo "Publishes current PKGBUILD to AUR."
    echo "Downloads the source tarball, computes sha256sum,"
    echo "updates PKGBUILD, generates .SRCINFO, and pushes to AUR."
    echo ""
    echo "Options:"
    echo "  --dry-run    Show what would happen without making changes"
    exit 1
}

DRY_RUN=false
for arg in "$@"; do
    case "$arg" in
        --dry-run) DRY_RUN=true ;;
        -h|--help) usage ;;
        *) echo "Unknown option: $arg"; usage ;;
    esac
done

REPO_ROOT="$(git rev-parse --show-toplevel)"
cd "$REPO_ROOT"

if [[ ! -f PKGBUILD ]]; then
    echo "ERROR: PKGBUILD not found in repo root."
    exit 1
fi

PKGNAME=$(grep -oP '^pkgname=\K.*' PKGBUILD)
PKGVER=$(grep -oP '^pkgver=\K.*' PKGBUILD)
URL=$(grep -oP '^url="\K[^"]+' PKGBUILD)
TARBALL_URL="${URL}/archive/v${PKGVER}.tar.gz"

echo "=== AUR Publish Plan ==="
echo "  Package:    ${PKGNAME}"
echo "  Version:    ${PKGVER}"
echo "  Tarball:    ${TARBALL_URL}"
echo ""

TMPDIR=$(mktemp -d)
trap 'rm -rf "$TMPDIR"' EXIT

echo "Downloading source tarball..."
if ! curl -fsSL -o "${TMPDIR}/source.tar.gz" "$TARBALL_URL"; then
    echo "ERROR: Failed to download tarball. Is v${PKGVER} released on GitHub?"
    echo "Run ./scripts/release.sh first."
    exit 1
fi

SHA256=$(sha256sum "${TMPDIR}/source.tar.gz" | awk '{print $1}')
echo "  sha256sum:  ${SHA256}"
echo ""

if [[ "$DRY_RUN" == "true" ]]; then
    echo "[dry-run] Would update PKGBUILD sha256sums to ${SHA256}"
    echo "[dry-run] Would clone aur@aur.archlinux.org:${PKGNAME}.git"
    echo "[dry-run] Would generate .SRCINFO and push to AUR"
    exit 0
fi

AUR_DIR="${TMPDIR}/aur"
echo "Cloning AUR repository..."
git clone "ssh://aur@aur.archlinux.org/${PKGNAME}.git" "$AUR_DIR"

cp PKGBUILD "${AUR_DIR}/PKGBUILD"

sed -i "s/^sha256sums=.*/sha256sums=('${SHA256}')/" "${AUR_DIR}/PKGBUILD"
echo "Updated sha256sums in AUR PKGBUILD"

echo "Generating .SRCINFO..."
(cd "$AUR_DIR" && makepkg --printsrcinfo > .SRCINFO)

cd "$AUR_DIR"
git add PKGBUILD .SRCINFO
if git diff --cached --quiet; then
    echo "No changes to publish — AUR is already up to date."
    exit 0
fi

git commit -m "Update to ${PKGVER}"
echo ""
echo "Pushing to AUR..."
git push

echo ""
echo "=== Done ==="
echo "Package ${PKGNAME} ${PKGVER} published to AUR."
echo "https://aur.archlinux.org/packages/${PKGNAME}"
