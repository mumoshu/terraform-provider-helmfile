module github.com/mumoshu/terraform-provider-helmfile

go 1.13

require (
	github.com/Masterminds/semver v1.5.0
	github.com/hashicorp/terraform-plugin-sdk v1.0.0
	github.com/mitchellh/go-linereader v0.0.0-20190213213312-1b945b3263eb
	github.com/mumoshu/shoal v0.2.13
	github.com/pkg/profile v1.5.0
	github.com/rs/xid v1.2.1
)

replace github.com/fishworks/gofish => github.com/mumoshu/gofish v0.13.1-0.20200908033248-ab2d494fb15c

replace git.apache.org/thrift.git => github.com/apache/thrift v0.0.0-20180902110319-2566ecd5d999
