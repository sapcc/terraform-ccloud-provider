module github.com/sapcc/terraform-provider-ccloud

go 1.15

require (
	github.com/ghodss/yaml v1.0.0
	github.com/go-openapi/runtime v0.19.5
	github.com/go-openapi/strfmt v0.19.3
	github.com/go-openapi/validate v0.19.3
	github.com/gophercloud/gophercloud v0.15.1-0.20210205220151-18b16b34db5c
	github.com/gophercloud/utils v0.0.0-20210209042946-13abf2251886
	github.com/hashicorp/terraform-plugin-sdk v1.16.0
	github.com/sapcc/gophercloud-sapcc v0.0.0-20201205201040-3739d487e866
	github.com/sapcc/kubernikus v1.5.1-0.20210209164035-81def994cb3c
	github.com/sapcc/limes v0.0.0-20210202142824-88364f9c65af
	k8s.io/apimachinery v0.21.0-alpha.3 // indirect
	k8s.io/client-go v0.0.0-20210209155049-20732a1bc198
)
