package main

# Check that no name attribute exists twice among all resources
deny[msg] {
	name := input[_].contents.metadata.name
	occurrences := [name | some i; input[i].contents.metadata.name == name; name := input[i].metadata.name]
	count(occurrences) > 1
	msg = sprintf("Error duplicate name : %s", [name])
}

deny[msg] {
	kind := input[_].contents.kind
	name := input[_].contents.metadata.name
	kind == "team"

	some i, j

	# list all existing users
	existing_users = {email | some i; input[i].contents.kind == "user"; email := input[i].contents.metadata.email}

	# gather all configured users in teams
	configured_owner_users_array = [user | input[i].contents.kind == "team"; user := input[i].contents.spec.owner]
	configured_member_users_array = [user | input[i].contents.kind == "team"; user := input[i].contents.spec.member]

	configured_users_array = array.concat(configured_owner_users_array, configured_member_users_array)

	# create a set to remove duplicates
	configured_users = {team | team := configured_users_array[i][j]}

	# sets can be substracted
	missing_users := configured_users - existing_users

	# missing users are the ones configured in teams but not in Github
	count(missing_users) > 0

	msg = sprintf(
		"Existing users %s | Configured users %s | Missing users %s",
		[sort(existing_users), sort(configured_users), sort(missing_users)],
	)
}
