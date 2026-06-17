#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")"

PKG_LIBS="libavformat libavcodec libavutil libavfilter libswscale"
CSRC_DIR="."
LIB_NAME="libaddon.a"

if ! command -v pkg-config >/dev/null 2>&1; then
  echo "error: pkg-config not found" >&2
  exit 1
fi

if ! pkg-config --exists ${PKG_LIBS}; then
  echo "error: FFmpeg libraries not found (pkg-config: ${PKG_LIBS})" >&2
  exit 1
fi

CFLAGS="$(pkg-config --cflags ${PKG_LIBS})"
CFLAGS="${CFLAGS} -I. -O2 -Wall -fPIC"

shopt -s nullglob
sources=( "${CSRC_DIR}"/*.c )
shopt -u nullglob

if [[ ${#sources[@]} -eq 0 ]]; then
  echo "error: no .c files found in ${CSRC_DIR}/" >&2
  exit 1
fi

objs=()
for src in "${sources[@]}"; do
  base="$(basename "${src}" .c)"
  obj="${CSRC_DIR}/${base}.o"
  echo "CC  ${src} -> ${obj}"
  gcc ${CFLAGS} -c "${src}" -o "${obj}"
  objs+=("${obj}")
done

echo "AR  ${LIB_NAME}"
rm -f "${LIB_NAME}"
ar rcs "${LIB_NAME}" "${objs[@]}"

echo "done: ${LIB_NAME} (${#objs[@]} object(s): ${objs[*]})"
