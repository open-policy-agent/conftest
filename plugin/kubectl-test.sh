#!/bin/bash

# kubectl test allows for testing resources in your cluster using Open Policy Agent
# It uses the conftest utility and expects to find associated policy files in
# a directory called policy


# Check if a specified command exists on the path and is executable
function check_command () {
    if ! [[ -x $(command -v $1) ]] ; then
        echo "$1 not installed"
        exit 1
    fi
}

function usage () {
    echo "A Kubectl plugin for using Conftest to test objects in Kubernetes using Open Policy Agent"
    echo
    echo "See https://github.com/instrumenta/conftest for more information"
    echo
    echo "Usage:"
    echo "   kubectl test (TYPE[.VERSION][.GROUP] [NAME] | TYPE[.VERSION][.GROUP]/NAME)"
}

# Check the required commands are available on the PATH
check_command "conftest"
check_command "kubectl"


if [[ ($# -eq 0) || ($1 == "--help") ]]; then
    # No commands or the --help flag passed and we'll show the usage instructions
    usage
elif [[ ($# -eq 1) && $1 =~ ^[a-z\.]+$ ]]; then
    # If we have one argument we get the list of objects from kubectl
    # parse our the individual items and then pass those one by one into conftest
    check_command "jq"
    if output=$(kubectl get $1 $2 -o json); then
        echo $output | jq -cj '.items[] | tostring+"\u0000"' | xargs -n1 -0 -I@ bash -c "echo '@' | conftest test -"
    fi
elif [[ ($# -eq 1 ) ]]; then
    # Support the / variant for getting an individual resource
    if output=$(kubectl get $1 -o json); then
        echo $output | conftest test -
    fi
elif [[ ($# -eq 2 ) && $1 =~ ^[a-z]+$ ]]; then
    # if we have two arguments then we assume the first is the type and the second the resource name
    if output=$(kubectl get $1 $2 -o json); then
        echo $output | conftest test -
    fi
else
    echo "Please check the arguments to kubectl test"
    echo
    usage
    exit 1
fi
