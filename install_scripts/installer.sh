#!/bin/sh

# references: https://github.com/mamba-org/micromamba-releases
# references: https://raw.githubusercontent.com/mamba-org/micromamba-releases/main/install.sh

# set -eu

# Detect the shell from which the script was called
parent=$(ps -o comm $PPID |tail -1)
parent=${parent#-}  # remove the leading dash that login shells have
case "$parent" in
  bash|fish|zsh)
    shell=$parent
    ;;
  *)
    # use the login shell (basename of $SHELL) as a fallback
    shell=${SHELL##*/}
    ;;
esac

echo "Parent shell: $parent"

find_writable_path_dir() {
    comfycli_path=$(command -v comfycli >/dev/null 2>&1 || echo "")

    if [ -n "$comfycli_path" ]; then
        comfycli_dir=$(dirname "$comfycli_path")

        if [ -w "$comfycli_dir" ]; then
            echo "$comfycli_dir"
            return 0
        fi
    fi

    path=$PATH
    IFS=':' read -r dir1 dir2 dir3 rest <<< "$PATH"

    for dir in "${dirs[@]}"; do
        if [ -w "$dir" ]; then
            echo "$dir"
            return 0
        fi
    done

    mkdir -p ~/.local/bin
    echo "~/.local/bin"
}

# Call the function and store the result in a variable
writable_path_dir=$(find_writable_path_dir)

# Parsing arguments
if [ -t 0 ] ; then
  printf "Install comficli to: [$writable_path_dir] "
  read BIN_FOLDER

  # if BIN_FOLDER is empty, set it to the presented value
  BIN_FOLDER="${BIN_FOLDER:-$writable_path_dir}"
fi

# Computing artifact location
case "$(uname)" in
  Linux)
    PLATFORM="linux" ;;
  Darwin)
    PLATFORM="osx" ;;
  *NT*)
    PLATFORM="win" ;;
esac

ARCH="$(uname -m)"
case "$ARCH" in
  aarch64|ppc64le|arm64)
      ;;  # pass
  *)
    ARCH="64" ;;
esac

case "$PLATFORM-$ARCH" in
  linux-aarch64|linux-ppc64le|linux-64|osx-arm64|osx-64|win-64)
      ;;  # pass
  *)
    echo "Failed to detect your OS" >&2
    exit 1
    ;;
esac

if [ "${VERSION:-}" = "" ]; then
  # https://github.com/richinsley/comfycli/releases/latest/download/comfycli-osx-arm64
  RELEASE_URL="https://github.com/richinsley/comfycli/releases/latest/download/comfycli-${PLATFORM}-${ARCH}"
else
  # https://github.com/richinsley/comfycli/releases/download/v0.0.1/comfycli-osx-arm64
  RELEASE_URL="https://github.com/richinsley/comfycli/releases/download/comfycli-${VERSION}/comfycli-${PLATFORM}-${ARCH}"
fi

echo $BIN_FOLDER
echo $PLATFORM-$ARCH
echo $RELEASE_URL

# Downloading artifact
mkdir -p "${BIN_FOLDER}"
if hash curl >/dev/null 2>&1; then
  curl "${RELEASE_URL}" -o "${BIN_FOLDER}/comfycli" -fsSL --compressed ${CURL_OPTS:-}
elif hash wget >/dev/null 2>&1; then
  wget ${WGET_OPTS:-} -qO "${BIN_FOLDER}/comfycli" "${RELEASE_URL}"
else
  echo "Neither curl nor wget was found" >&2
  exit 1
fi
chmod +x "${BIN_FOLDER}/comfycli"

# test run the binary to initialize the home directory
"${BIN_FOLDER}/comfycli" --help
