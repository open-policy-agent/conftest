module github.com/instrumenta/conftest

go 1.13

require (
	cuelang.org/go v0.0.11
	github.com/BurntSushi/toml v0.3.1
	github.com/containerd/containerd v1.3.0
	github.com/deislabs/oras v0.7.0
	github.com/ghodss/yaml v1.0.0
	github.com/go-ini/ini v1.49.0
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/google/go-cmp v0.3.1 // indirect
	github.com/hashicorp/hcl v1.0.0
	github.com/hashicorp/hcl2 v0.0.0-20191002203319-fb75b3253c80
	github.com/kami-zh/go-capturer v0.0.0-20171211120116-e492ea43421d
	github.com/logrusorgru/aurora v0.0.0-20191017060258-dc85c304c434
	github.com/moby/buildkit v0.3.3
	github.com/open-policy-agent/opa v0.14.2
	github.com/opencontainers/image-spec v1.0.1
	github.com/rcrowley/go-metrics v0.0.0-20190826022208-cac0b30c2563 // indirect
	github.com/sirupsen/logrus v1.4.2
	github.com/smartystreets/goconvey v0.0.0-20190731233626-505e41936337 // indirect
	github.com/spf13/cobra v0.0.5
	github.com/spf13/viper v1.4.0
	github.com/stretchr/testify v1.4.0
	github.com/yashtewari/glob-intersection v0.0.0-20180916065949-5c77d914dd0b // indirect
	github.com/zclconf/go-cty v1.1.0
	gopkg.in/ini.v1 v1.49.0 // indirect
	gotest.tools v2.2.0+incompatible
)

replace github.com/docker/docker v0.0.0-00010101000000-000000000000 => github.com/moby/moby v0.7.3-0.20190826074503-38ab9da00309
