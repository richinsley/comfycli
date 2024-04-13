#!/bin/sh

# references: https://github.com/mamba-org/micromamba-releases
# references: https://raw.githubusercontent.com/mamba-org/micromamba-releases/main/install.sh

# set -eu

if [ -n "${COMFYCLI_PARENT_PATH}" ]; then
    PATH="$COMFYCLI_PARENT_PATH"
else
    COMFYCLI_PARENT_PATH="$PATH"
fi

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

find_writable_path_dir() {
    comfycli_path=$(command -v comfycli >/dev/null 2>&1 || echo "")

    if [ -n "$comfycli_path" ]; then
        comfycli_dir=$(dirname "$comfycli_path")

        if [ -w "$comfycli_dir" ]; then
            echo "$comfycli_dir"
            return 0
        fi
    fi

    common_install_paths=(
        "$HOME/.local/bin"
        "$HOME/bin"
        "$HOME/.bin"
        "/usr/local/bin"
        "/opt/local/bin"
        "/usr/bin"
        "/opt/homebrew/bin"
    )

    IFS=':' path_dirs=("${(@s/:/)COMFYCLI_PARENT_PATH}")

    for dir in "${common_install_paths[@]}"; do
        if [[ "${path_dirs[@]}" =~ "$dir" ]] && [ -w "$dir" ]; then
            echo "$dir"
            return 0
        fi
    done

    echo "${HOME}/.local/bin"
}


# Call the function and store the result in a variable
writable_path_dir=$(find_writable_path_dir)

if [ $# -eq 1 ]; then
  BIN_FOLDER="$1"
else
  if [ -t 0 ] ; then
    printf "Install comficli to: [$writable_path_dir] "
    read BIN_FOLDER

    # if BIN_FOLDER is empty, set it to the presented value
    BIN_FOLDER="${BIN_FOLDER:-$writable_path_dir}"
  else
    BIN_FOLDER="$writable_path_dir"
  fi
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
  RELEASE_URL="https://github.com/richinsley/comfycli/releases/latest/download/comfycli-${PLATFORM}-${ARCH}"
else
  RELEASE_URL="https://github.com/richinsley/comfycli/releases/download/comfycli-${VERSION}/comfycli-${PLATFORM}-${ARCH}"
fi

# Downloading artifact
echo "Downloading comfycli from: ${RELEASE_URL}"
if ! mkdir -p "${BIN_FOLDER}"; then
  echo "Failed to create directory: ${BIN_FOLDER}" >&2
  exit 1
fi

if hash curl >/dev/null 2>&1; then
  if ! curl "${RELEASE_URL}" -o "${BIN_FOLDER}/comfycli" -fsSL --compressed ${CURL_OPTS:-}; then
    echo "Failed to download comfycli using curl" >&2
    exit 1
  fi
elif hash wget >/dev/null 2>&1; then
  if ! wget ${WGET_OPTS:-} -qO "${BIN_FOLDER}/comfycli" "${RELEASE_URL}"; then
    echo "Failed to download comfycli using wget" >&2
    exit 1
  fi
else
  echo "Neither curl nor wget was found" >&2
  exit 1
fi

chmod +x "${BIN_FOLDER}/comfycli"

echo "comfycli has been successfully installed to: ${BIN_FOLDER}"
echo "You can now use the 'comfycli' command."

case ":$PATH:" in
  *":${BIN_FOLDER}:"*) ;;
  *)
    echo
    echo "Please add ${BIN_FOLDER} to your PATH to use comfycli from anywhere."
    echo "You can do this by adding the following line to your shell profile (e.g., ~/.bashrc, ~/.zshrc):"
    echo "export PATH=\"${BIN_FOLDER}:\$PATH\""
    ;;
esac

echo
echo "╭─────────────────────────────────────────────────────────────╮"
echo "                         🛠️  comfycli 🛠️"
echo "╰─────────────────────────────────────────────────────────────╯"

# test run the binary to initialize the home directory
help=$("${BIN_FOLDER}/comfycli" --help)
if [ $? -ne 0 ]; then
  echo "Failed to run comfycli" >&2
  exit 1
fi
echo $help
