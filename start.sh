#!/usr/bin/env bash
# One-shot setup for MorphoWheel: checks the toolchain, installs converters,
# runs tests, builds and installs the app, and enables start at login.
#
#   ./start.sh                 full setup
#   ./start.sh --no-autostart  install but do not start at login
#   ./start.sh --skip-deps     do not touch Homebrew
#   ./start.sh --uninstall     remove app, command and autostart
set -euo pipefail

cd "$(dirname "$0")"

AUTOSTART=1
DEPS=1
for arg in "$@"; do
	case "$arg" in
	--no-autostart) AUTOSTART=0 ;;
	--skip-deps) DEPS=0 ;;
	--uninstall)
		make uninstall
		echo "MorphoWheel removed."
		exit 0
		;;
	-h | --help)
		sed -n '2,8p' "$0" | sed 's/^# \{0,1\}//'
		exit 0
		;;
	*)
		echo "unknown option: $arg" >&2
		exit 2
		;;
	esac
done

step() { printf '\n==> %s\n' "$*"; }
die() {
	echo "error: $*" >&2
	exit 1
}

[ "$(uname)" = "Darwin" ] || die "MorphoWheel works on macOS only"

step "Toolchain"
xcode-select -p >/dev/null 2>&1 || die "Command Line Tools missing: run 'xcode-select --install' and rerun"
command -v go >/dev/null 2>&1 || die "Go missing: brew install go"
echo "go: $(go version)"

# Converters. Format: command|brew install args.
if [ "$DEPS" = 1 ]; then
	step "Converters"
	command -v brew >/dev/null 2>&1 || die "Homebrew missing (https://brew.sh) or rerun with --skip-deps"
	for entry in "ffmpeg|ffmpeg" "magick|imagemagick" "pandoc|pandoc" "soffice|--cask libreoffice"; do
		cmd="${entry%%|*}"
		pkg="${entry#*|}"
		if command -v "$cmd" >/dev/null 2>&1 || [ -x "/opt/homebrew/bin/$cmd" ] || [ -x "/usr/local/bin/$cmd" ]; then
			echo "$cmd: found"
		else
			echo "$cmd: installing ($pkg)"
			# shellcheck disable=SC2086
			brew install $pkg
		fi
	done
fi

# An app started at login gets a minimal PATH, so a converter living in a
# custom directory (e.g. a hand-built ffmpeg) would not be found. Link such
# tools into the app's own tools directory, which it always searches.
step "Converters outside standard locations"
TOOLS_DIR="$HOME/Library/Application Support/MorphoWheel/tools"
for cmd in ffmpeg magick pandoc soffice; do
	path="$(command -v "$cmd" 2>/dev/null || true)"
	[ -n "$path" ] || continue
	case "$path" in
	/opt/homebrew/bin/* | /usr/local/bin/* | /usr/bin/*) continue ;;
	esac
	mkdir -p "$TOOLS_DIR"
	ln -sf "$path" "$TOOLS_DIR/$cmd"
	echo "$cmd: linked $path"
done

step "Tests"
make test

step "Build and install"
if [ "$AUTOSTART" = 1 ]; then
	make autostart-on
else
	make autostart-off
	make install
fi

cat <<'EOF'

Done.

  Menu bar icon appears in a few seconds.
  Drag a file with Shift / Control / Option held, or select files in Finder
  and press Control+Option+Space.

First launch: macOS may ask for Input Monitoring and Automation (Finder)
permissions. Grant them in System Settings -> Privacy & Security.
EOF
