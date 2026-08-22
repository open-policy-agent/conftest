package main

import rego.v1

risky_shell_calls contains call if {
	walk(input, [_, call])
	call.kind == "CallExpr"
	call.children[0].kind == "IdentExpr"
	call.children[0].children[0].kind == "Identifier"
	call.children[0].children[0].text in {"bat", "powershell", "pwsh", "sh"}
}

# Join static fragments so slashy strings and split GStrings cannot hide a command.
shell_command(call) := command if {
	fragments := [token.text |
		walk(call, [_, token])
		token.kind in {
			"DollarSlashyStringLit",
			"GStringBegin",
			"GStringEnd",
			"GStringFull",
			"GStringPart",
			"SlashyStringLit",
			"StringLiteral",
		}
	]
	command := regex.replace(lower(concat("", fragments)), `["'$]`, "")
}

deny contains violation if {
	some call in risky_shell_calls
	command := shell_command(call)
	regex.match(`(^|[^[:alnum:]_])(curl|wget)[[:space:]]`, command)
	regex.match(`\|[[:space:]]*(bash|powershell|pwsh|sh)([^[:alnum:]_]|$)`, command)
	violation := {
		"msg": "Pipeline downloads a script and pipes it directly to an interpreter",
		"_loc": {
			"file": data.conftest.file.name,
			"line": call.line,
		},
	}
}
