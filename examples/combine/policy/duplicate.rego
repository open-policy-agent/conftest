package main

# Check that no name attribute exists twice among all resources
deny[msg] {
    name := input[_].metadata.name
    occurrences := [name | input[i].metadata.name == name; name := input[i].metadata.name]
    count(occurrences) > 1
    msg = sprintf("Error duplicate name : %s", [name])
}

deny[msg] {
    kind = input[_].kind
    name = input[_].metadata.name
    kind = "team"
    # list all existing users
    existing_users = { email | input[i].kind = "user" ; email := input[i].metadata.email }

    # gather all configured users in teams
    configured_owner_users_array = [ user | input[i].kind = "team" ; user := input[i].spec.owner ]
    configured_member_users_array = [ user | input[i].kind = "team" ; user := input[i].spec.member ]

    configured_users_array = array.concat(configured_owner_users_array, configured_member_users_array)
    # create a set to remove duplicates
    configured_users = { team | team := configured_users_array[i][j] }

    # sets can be substracted
    missing_users := configured_users - existing_users
    # missing users are the ones configured in teams but not in Github
    count(missing_users) > 0

    msg = sprintf("\nExisting users %s \nConfigured users %s \nMissing users %s", [sort(existing_users), sort(configured_users), sort(missing_users)])
}