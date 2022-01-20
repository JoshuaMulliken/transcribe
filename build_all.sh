#!/bin/bash

VERSION=0.0.1

echo "[INFO] Building version $VERSION"

echo "[INFO] Cleaning up"
make clean

echo "[INFO] Creating tar archive..."
make archive VERSION_STR="-v${VERSION}"

echo "[INFO] Creating build directory"
mkdir -p build

echo "[INFO] Copying archive to build directory"
cp /tmp/transcribe-v$VERSION.tar.gz build/

echo "[INFO] Building binaries for all architectures..."
make build_all VERSION_STR="-v${VERSION}" builddir=build

echo "[INFO] Creating github release..."
gh release create ${VERSION} -d \
  -t "Transcribe v$VERSION" \
  -F changelog/$VERSION.md \
  build/*