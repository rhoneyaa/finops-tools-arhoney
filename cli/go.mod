module github.com/openshift-online/finops-tools/cli

go 1.24

require (
	github.com/aws/aws-sdk-go-v2 v1.41.7
	github.com/aws/aws-sdk-go-v2/config v1.32.18
	github.com/aws/aws-sdk-go-v2/credentials v1.19.17
	github.com/aws/aws-sdk-go-v2/service/organizations v1.51.3
	github.com/aws/aws-sdk-go-v2/service/sts v1.42.1
	github.com/aws/smithy-go v1.25.1
	github.com/mattn/go-isatty v0.0.20
	github.com/olekukonko/tablewriter v0.0.5
	github.com/openshift-online/finops-tools/core v0.0.0
	github.com/spf13/cobra v1.9.1
	golang.org/x/term v0.30.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.18.23 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.4.23 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.7.23 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.4.24 // indirect
	github.com/aws/aws-sdk-go-v2/service/costexplorer v1.63.8 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.13.9 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.13.23 // indirect
	github.com/aws/aws-sdk-go-v2/service/signin v1.0.11 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.30.17 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.36.0 // indirect
	github.com/clipperhouse/uax29/v2 v2.6.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/mattn/go-runewidth v0.0.19 // indirect
	github.com/spf13/pflag v1.0.6 // indirect
	golang.org/x/sys v0.31.0 // indirect
)

replace github.com/openshift-online/finops-tools/core => ../core
