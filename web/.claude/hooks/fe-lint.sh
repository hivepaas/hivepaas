#!/usr/bin/env bash
# Formats and lints a front-end file right after Claude Code edits it, so a formatting or lint
# problem comes back in the same turn instead of surfacing on a later manual run.
#
# Wired up as a PostToolUse hook for Edit|Write. Anything outside a project that carries the local
# tooling is a no-op, so this stays silent in other repositories.
#
# NOTE: `tsc --noEmit` is deliberately left out. It type-checks the whole project (~7s) and reports
# every pre-existing error in the repo on every single edit, so it belongs in `yarn lint:ci`, not
# in a per-edit hook.

set -uo pipefail

payload=$(cat)
file=$(printf '%s' "$payload" | jq -r '.tool_input.file_path // empty' 2>/dev/null)

[[ -n "${file}" && -f "${file}" ]] || exit 0
[[ "${file}" == */node_modules/* ]] && exit 0

case "${file}" in
    *.ts|*.tsx|*.js|*.jsx|*.mjs|*.cjs|*.json|*.css|*.scss|*.md) ;;
    *) exit 0 ;;
esac

dir=$(cd "$(dirname "${file}")" 2>/dev/null && pwd) || exit 0

# The tooling is the project's own, not whatever happens to be on PATH, so find the root that
# actually carries it rather than trusting the current directory.
root="${dir}"
while [[ "${root}" != "/" && ! -d "${root}/node_modules/.bin" ]]; do
    root=$(dirname "${root}")
done
[[ -d "${root}/node_modules/.bin" ]] || exit 0

prettier="${root}/node_modules/.bin/prettier"
eslint="${root}/node_modules/.bin/eslint"

# Prettier runs in ~0.3s and settles formatting before eslint looks at the file.
[[ -x "${prettier}" ]] && "${prettier}" --write --log-level=warn "${file}" >/dev/null 2>&1

case "${file}" in
    *.ts|*.tsx|*.js|*.jsx|*.mjs|*.cjs) ;;
    *) exit 0 ;;
esac

[[ -x "${eslint}" ]] || exit 0

output=$(cd "${root}" && "${eslint}" --fix "${file}" 2>&1)
status=$?

[[ ${status} -eq 0 ]] && exit 0

# Exit 2 puts this in front of Claude. That is the point: it fixes the finding now, instead of
# handing over a file that fails `yarn lint`.
{
    echo "eslint found issues in ${file} (auto-fixable ones were already applied):"
    echo "${output}"
} >&2
exit 2
