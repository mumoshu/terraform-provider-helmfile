package helmfile

import (
	"github.com/mumoshu/terraform-provider-eksctl/pkg/sdk"
	"github.com/mumoshu/terraform-provider-eksctl/pkg/sdk/api"
	"github.com/mumoshu/terraform-provider-eksctl/pkg/sdk/tfsdk"
)

func newContext(d api.Getter) *sdk.Context {
	conf := tfsdk.ConfigFromResourceData(d,
		tfsdk.SchemaOptionAWSRegionKey(KeyAWSRegion),
		tfsdk.SchemaOptionAWSProfileKey(KeyAWSProfile),
		tfsdk.SchemaOptionAWSAssumeRole(KeyAWSAssumeRole),
	)

	ctx := sdk.ContextConfig(conf)

	return ctx
}
