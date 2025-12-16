#!/usr/bin/env bash

set -eo pipefail

code=0
while read -r commit; do
    match=$(echo "${commit}" | grep -o -h -E "^[a-z]+(\([a-z-]+\))?: " || true)
    if [[ -z "${match}" ]]; then
        echo "::error::Commit \"${commit}\" does not have the required conventional commit prefix. See https://www.conventionalcommits.org/ for more info."
        code=1
    else
        echo "Commit \"${commit}\" has conventional commit prefix \"${match}\"."
    fi
done < <(git --no-pager log --pretty=format:%s && echo "") # git log does not include newline after last commit

exit ${code}
