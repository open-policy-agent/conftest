package main

deny[msg] {
	expectedDataLicense := "conftest-demo"
	input.CreationInfo.DataLicense != expectedDataLicense
	msg := sprintf("DataLicense should be %d, but found %d", [expectedDataLicense, input.CreationInfo.DataLicense])
}
