module github.com/sapcc/terraform-provider-ccloud

go 1.14

require (
	github.com/ghodss/yaml v1.0.0
	github.com/go-openapi/runtime v0.19.5
	github.com/go-openapi/strfmt v0.19.3
	github.com/go-openapi/validate v0.19.3
	github.com/gophercloud/gophercloud v0.13.1-0.20201031160030-9e602f1dcf38
	github.com/gophercloud/utils v0.0.0-20201101202656-8677e053dcf1
	github.com/hashicorp/terraform-plugin-sdk v1.15.0
	github.com/sapcc/gophercloud-sapcc v0.0.0-20201030120356-149286e82617
	github.com/sapcc/kubernikus v1.0.1-0.20200130145221-13142045b03f
	github.com/sapcc/limes v0.0.0-20200420152328-facea01fd1ab
	k8s.io/apimachinery v0.0.0-20190831074630-461753078381 // indirect
	k8s.io/client-go v11.0.0+incompatible
)
