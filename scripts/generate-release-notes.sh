#!/usr/bin/env bash
# Generate release notes from git history between tags
# Usage: ./generate-release-notes.sh [PREV_TAG] [CURRENT_TAG]
# If PREV_TAG is not provided, uses first commit
# If CURRENT_TAG is not provided, uses HEAD

set -euo pipefail

REPO_OWNER="ChocolateNao"
REPO_NAME="tuichart"

PREV="${1:-}"
CURRENT="${2:-HEAD}"

if [[ -z "$PREV" ]]; then
  if git tag -l "v*" | sort -V | tail -2 | head -1 | grep -q .; then
    PREV=$(git tag -l "v*" | sort -V | tail -2 | head -1)
  else
    PREV=$(git rev-list --max-parents=0 HEAD)
  fi
fi

if [[ "$CURRENT" == refs/tags/* ]]; then
  CURRENT="${CURRENT#refs/tags/}"
fi

DATE=$(date +%Y-%m-%d)

if [[ "$PREV" == v* ]]; then
  COMPARE_URL="https://github.com/$REPO_OWNER/$REPO_NAME/compare/$PREV...$CURRENT"
else
  COMPARE_URL="https://github.com/$REPO_OWNER/$REPO_NAME/commits/$CURRENT"
fi

AWK_FEATURES='
  {
    hash=$1
    rest=substr($0, length($1)+2)
    if (rest ~ /^feat: /) {
      sub(/^feat: /, "", rest)
      type="feat"
      print "- [" hash "] **" type ":** " rest
    } else if (rest ~ /^[^:]+: /) {
      # Skip other commit tags
    } else {
      # Treat non-conventional commit as feature
      type="feat"
      print "- [" hash "] **" type ":** " rest
    }
  }'

AWK_FIXES='
  {
    hash=$1
    rest=substr($0, length($1)+2)
    if (rest ~ /^fix: /) {
      sub(/^fix: /, "", rest)
      type="fix"
      print "- [" hash "] **" type ":** " rest
    }
  }'

AWK_OTHER='
  {
    hash=$1
    rest=substr($0, length($1)+2)
    if (rest ~ /^feat: / || rest ~ /^fix: /) {
      # Skip for other changes
    } else if (rest ~ /^[^:]+: /) {
      split(rest, t, /: /)
      type=t[1]
      rest=substr(rest, length(t[1])+3)
      print "- [" hash "] **" type ":** " rest
    } else {
      # Skip non-conventional for other changes
    }
  }'

{
  echo "# [$CURRENT]($COMPARE_URL) ($DATE)"
  echo ""
  echo "## Features"
  echo ""
  git log --oneline --no-merges "$PREV..$CURRENT" | awk "$AWK_FEATURES"
  echo ""
  echo "## Bug Fixes"
  echo ""
  git log --oneline --no-merges "$PREV..$CURRENT" | awk "$AWK_FIXES"
  echo ""
  echo "## Other Changes"
  echo ""
  git log --oneline --no-merges "$PREV..$CURRENT" | awk "$AWK_OTHER"
  echo ""
  echo "## Contributors"
  echo ""
  git shortlog -sne "$PREV..$CURRENT" | sed 's/^ *\([0-9]*\)	\(.*\) <\(.*\)>/- \1 commits authored by \2 (<\3>)/'
}
