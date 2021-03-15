module github.com/open-policy-agent/conftest

go 1.13

require (
	cuelang.org/go v0.0.15
	github.com/BurntSushi/toml v0.3.1
	github.com/KeisukeYamashita/go-vcl v0.4.0
	github.com/basgys/goxml2json v1.1.0
	github.com/deislabs/oras v0.10.0
	github.com/ghodss/yaml v1.0.0
	github.com/go-akka/configuration v0.0.0-20200606091224-a002c0330665
	github.com/go-ini/ini v1.62.0
	github.com/google/go-jsonnet v0.17.0
	github.com/hashicorp/go-getter v1.5.2
	github.com/hashicorp/hcl v1.0.0
	github.com/jstemmer/go-junit-report v0.9.1
	github.com/logrusorgru/aurora v2.0.3+incompatible
	github.com/moby/buildkit v0.8.2
	github.com/olekukonko/tablewriter v0.0.5
	github.com/open-policy-agent/opa v0.27.1
	github.com/opencontainers/image-spec v1.0.1
	github.com/shteou/go-ignore v0.3.0
	github.com/spf13/cobra v1.1.3
	github.com/spf13/viper v1.7.1
	github.com/tmccombs/hcl2json v0.3.1
	olympos.io/encoding/edn v0.0.0-20200308123125-93e3b8dd0e24
)

replace (
	github.com/docker/distribution => github.com/docker/distribution v0.0.0-20191216044856-a8371794149d
	github.com/docker/docker => github.com/moby/moby v17.12.0-ce-rc1.0.20200618181300-9dc6525e6118+incompatible
)
