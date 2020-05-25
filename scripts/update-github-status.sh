#!/bin/bash
set -euox pipefail

export BASE_DIR=$(git rev-parse --show-superproject-working-tree --show-toplevel)
export COMMIT_ID=$(git rev-parse HEAD)
export POLICY_DIR="$BASE_DIR/policy/$POLICY_NAME"
export POLICY_RESULT=""


policy_dir_exist() {
    if [ ! -e "$POLICY_DIR" ]; then
        echo "No OPA files (*.rego) present, skipping OPA check!"
        exit 0;
    fi
}

eval_policy_name() {
    update_github_status "pending" $POLICY_NAME
}

run_policy_check() {
    if POLICY_RESULT=$(conftest test --no-color -p $POLICY_DIR $PLANFILE.json)
    then
        state="success"
    else
        state="failure"
    fi

    update_github_status $state $POLICY_NAME
    update_github_comment $POLICY_NAME $POLICY_RESULT
}

update_github_status() {
    curl --header "Authorization: token $ATLANTIS_GH_TOKEN" \
        --header "Content-Type: application/json" \
        --data \
        '{"state": "'$1'", "context": "'open-policy-agent/$2'", "description": "'"OPA Policy Check"'"}' \
        --request POST \
        "https://api.github.com/repos/$BASE_REPO_OWNER/$BASE_REPO_NAME/statuses/$COMMIT_ID" \
        >/dev/null 2>/dev/null
}

update_github_comment() {
    curl --header "Authorization: token $ATLANTIS_GH_TOKEN" \
        --header "Content-Type: application/json" \
        --data-binary \
        '{"body": "'"${POLICY_RESULT//$'\n'/'\n'}"'"}' \
        --request POST \
        "https://api.github.com/repos/$BASE_REPO_OWNER/$BASE_REPO_NAME/issues/$PULL_NUM/comments" \
        >/dev/null 2>/dev/null
}

policy_dir_exist
eval_policy_name
run_policy_check