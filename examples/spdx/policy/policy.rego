package main
import rego.v1

deny contains msg if {
	expected_data_license := "conftest-demo"
	input.CreationInfo.DataLicense != expected_data_license
	msg := sprintf("DataLicense should be %d, but found %d", [expected_data_license, input.CreationInfo.DataLicense])
}
