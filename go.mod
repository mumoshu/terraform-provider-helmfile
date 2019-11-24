module github.com/mumoshu/terraform-provider-helmfile

go 1.12

require (
	github.com/armon/circbuf v0.0.0-20190214190532-5111143e8da2
	github.com/hashicorp/terraform-plugin-sdk v1.0.0
	github.com/mitchellh/go-linereader v0.0.0-20190213213312-1b945b3263eb
	github.com/posener/complete v1.2.1
	github.com/rs/xid v1.2.1
)

replace git.apache.org/thrift.git => github.com/apache/thrift v0.0.0-20180902110319-2566ecd5d999
