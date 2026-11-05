package workspacedeps

import "strings"

// toolkitBinDir is the workspace image's toolkit bin directory. It is
// appended to PATH when present and consulted by discovery as the fallback
// location for image-provided commands. Remote targets do not have it.
const toolkitBinDir = "/opt/ami/toolkit/bin"

// exitCodeLocked is the status returned by the OS locker on contention.
// The runner also checks that the prelude never reported acquisition.
const (
	exitCodeLocked     = 75
	lockAcquiredMarker = "__AMI_DEP_LOCK_ACQUIRED__"
)

// lockProbeHelpers asks the kernel whether the stable lock file is held.
// Lock files are never unlinked: doing so would create a second lock inode.
// Legacy directory locks are refused rather than reclaimed behind a live run.
const lockProbeHelpers = `ami_lock_active() {
  [ -d "$1" ] && return 0
  [ -f "$1" ] || return 1
  case "$(uname -s)" in
    Darwin) ! lockf -k -t 0 "$1" true >/dev/null 2>&1 ;;
    *) ! flock -n "$1" true >/dev/null 2>&1 ;;
  esac
}
`

// scriptExecWrapper acquires an OS advisory lock before sh reads the script.
// flock and lockf retain the same inode until the entire command exits, and
// the kernel releases it on crash. There is no user-space stale-lock deletion.
const scriptExecWrapper = `ami_lock="$(dirname "$AMI_DEP_HOME")/.locks/$AMI_DEP_ID.lock"
mkdir -p "$(dirname "$ami_lock")"
if [ -d "$ami_lock" ]; then
  printf '%s\n' 'legacy dependency lock directory requires operator recovery' >&2
  exit 73
fi
case "${AMI_DEP_OS:-$(uname -s)}" in
  darwin|Darwin)
    # FD mode keeps the lock in the shell and its descendants. Command mode
    # unlocks when the lockf supervisor exits, even if its child is still alive.
    exec 9>> "$ami_lock"
    lockf -k -t 0 9 || exit "$?"
    exec sh -s ;;
  *) exec flock -n -E 75 "$ami_lock" sh -s ;;
esac`

// prelude is the POSIX sh text the runner feeds to `sh -s` ahead of every
// catalog script. It must stay POSIX and must pass
// `shellcheck -s sh`; TestPreludeShellcheck enforces the latter.
//
// The text ends with the opening of ami_dep_main so the script body is
// parsed as a function and later invoked with stdin redirected from
// /dev/null. Anything in the body that reads stdin (`read`, npm prompts,
// apt) therefore sees EOF instead of eating the rest of the script.
const prelude = `# ---- ami dependency runner prelude (injected) ----
set -eu
export DEBIAN_FRONTEND=noninteractive CI=1
export PATH="$AMI_DEP_BIN:$PATH"
if [ -d ` + toolkitBinDir + ` ]; then
  export PATH="$PATH:` + toolkitBinDir + `"
fi

# Public recipe helpers are intentionally usable from arbitrary script bodies.
# shellcheck disable=SC2329
dep_log()    { printf '%s\n' "$*" >&2; }
# shellcheck disable=SC2329
dep_result() { printf '%s' "$1" > "$AMI_DEP_RESULT"; }
# shellcheck disable=SC2329
dep_switch() {
  case "$AMI_DEP_OS" in
    darwin)
      # BSD mv has no -T. ln -sfh is unlink+create, close enough to atomic for
      # the serialized foreground dependency operation.
      ln -sfh "$1" "$AMI_DEP_HOME/current" ;;
    *)
      ln -sfn "$1" "$AMI_DEP_HOME/current.tmp"
      mv -Tf "$AMI_DEP_HOME/current.tmp" "$AMI_DEP_HOME/current" ;;
  esac
}

# Reaching the prelude proves the kernel lock was acquired, even when the
# recipe itself later exits 75.
printf '%s\n' '` + lockAcquiredMarker + `' >&2
if [ -n "${AMI_DEP_OPERATION_DIR:-}" ]; then
  ami_operation_parent="$(dirname "$AMI_DEP_OPERATION_DIR")"
  # A reaper can cancel a claimed operation before this process ever starts.
  # Check its durable fence while holding the same kernel lock as the reaper.
  if [ -f "$ami_operation_parent/.cancelled-${AMI_DEP_OPERATION_DIR##*/}" ]; then
    printf '%s\n' 'dependency operation was cancelled before execution' >&2
    exit 76
  fi
  printf '%s\n' "${AMI_DEP_OPERATION_DIR##*/}" > "$ami_operation_parent/current.$$"
  mv -f "$ami_operation_parent/current.$$" "$ami_operation_parent/current"
fi

ami_dep_main() {
`

// preludeEpilogue closes the function opened by prelude and runs it with
// stdin detached from the script source.
const preludeEpilogue = `}
set +e
(set -e; ami_dep_main < /dev/null)
ami_exit=$?
set -e
if [ -n "${AMI_DEP_OPERATION_DIR:-}" ]; then
  printf '%s\n' "$ami_exit" > "$AMI_DEP_OPERATION_DIR/exit-code.tmp"
  mv -f "$AMI_DEP_OPERATION_DIR/exit-code.tmp" "$AMI_DEP_OPERATION_DIR/exit-code"
fi
exit "$ami_exit"
`

// preludeLines is the number of stdin lines that precede the first line of
// the script body. Shells report syntax and runtime errors with line numbers
// counted from the start of stdin, so the runner subtracts this offset before
// forwarding stderr to the user.
var preludeLines = strings.Count(prelude, "\n")

// PreludeLines returns the line offset the prelude adds in front of a script
// body. Line k of the body is line PreludeLines()+k of what the shell reads.
func PreludeLines() int {
	return preludeLines
}

// WrapScript returns the full stdin text for `sh -s`: the prelude, the body
// as the contents of ami_dep_main, and the call that runs it with stdin
// redirected from /dev/null.
func WrapScript(body string) string {
	var b strings.Builder
	b.Grow(len(prelude) + len(body) + len(preludeEpilogue) + 1)
	b.WriteString(prelude)
	b.WriteString(body)
	if !strings.HasSuffix(body, "\n") {
		b.WriteByte('\n')
	}
	b.WriteString(preludeEpilogue)
	return b.String()
}

// shellQuote returns s as a single-quoted POSIX sh word.
func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}
