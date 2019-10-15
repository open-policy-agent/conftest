module github.com/instrumenta/conftest

go 1.13

replace (
	git.apache.org/thrift.git => github.com/apache/thrift v0.0.0-20180902110319-2566ecd5d999

	//go: github.com/moby/buildkit@v0.6.1 requires
	//        github.com/containerd/containerd@v1.3.0-0.20190507210959-7c1e88399ec0: invalid pseudo-version: version before v1.3.0 would have negative patch number
	github.com/containerd/containerd v1.3.0-0.20190507210959-7c1e88399ec0 => github.com/containerd/containerd v1.2.1-0.20190507210959-7c1e88399ec0

	// go: github.com/deislabs/oras@v0.7.0 requires
	//        github.com/docker/docker@v0.0.0-00010101000000-000000000000: invalid version: unknown revision 000000000000
	github.com/docker/docker v0.0.0-00010101000000-000000000000 => github.com/moby/moby v0.7.3-0.20190826074503-38ab9da00309

	//go: github.com/moby/buildkit@v0.6.1 requires
	//        github.com/docker/docker@v1.14.0-0.20190319215453-e7b5f7dbe98c: invalid pseudo-version: version before v1.14.0 would have negative patch number
	github.com/docker/docker v1.14.0-0.20190319215453-e7b5f7dbe98c => github.com/docker/docker v1.4.2-0.20190319215453-e7b5f7dbe98c

	//go: github.com/moby/buildkit@v0.6.1 requires
	//        github.com/tonistiigi/fsutil@v0.0.0-20190327153851-3bbb99cdbd76 requires
	//        golang.org/x/crypto@v0.0.0-20190129210102-0709b304e793: invalid pseudo-version: does not match version-control timestamp (2018-09-04T16:38:35Z)
	golang.org/x/crypto v0.0.0-20190129210102-0709b304e793 => golang.org/x/crypto v0.0.0-20180904163835-0709b304e793
)

require (
	cuelang.org/go v0.0.9
	github.com/BurntSushi/toml v0.3.1
	github.com/containerd/containerd v1.3.0-beta.2.0.20190823190603-4a2f61c4f2b4
	github.com/deislabs/oras v0.7.0
	github.com/ghodss/yaml v1.0.0
	github.com/go-ini/ini v1.46.0
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/hashicorp/hcl v1.0.0
	github.com/hashicorp/hcl2 v0.0.0-20190909202536-66c59f909e25
	github.com/kami-zh/go-capturer v0.0.0-20171211120116-e492ea43421d
	github.com/logrusorgru/aurora v0.0.0-20190803045625-94edacc10f9b
	github.com/moby/buildkit v0.6.1
	github.com/open-policy-agent/opa v0.14.0
	github.com/opencontainers/image-spec v1.0.1
	github.com/rcrowley/go-metrics v0.0.0-20190826022208-cac0b30c2563 // indirect
	github.com/sirupsen/logrus v1.4.2
	github.com/smartystreets/goconvey v0.0.0-20190731233626-505e41936337 // indirect
	github.com/spf13/cobra v0.0.3
	github.com/spf13/viper v1.4.0
	github.com/stretchr/testify v1.3.0
	github.com/yashtewari/glob-intersection v0.0.0-20180916065949-5c77d914dd0b // indirect
	github.com/yvasiyarov/newrelic_platform_go v0.0.0-20160601141957-9c099fbc30e9 // indirect
	github.com/zclconf/go-cty v1.0.0
	golang.org/x/crypto v0.0.0-20190513172903-22d7a77e9e5f // indirect
	google.golang.org/genproto v0.0.0-20190620144150-6af8c5fc6601 // indirect
	gopkg.in/ini.v1 v1.46.0 // indirect
	gotest.tools v2.2.0+incompatible
)
