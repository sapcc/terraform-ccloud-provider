---
layout: "ccloud"
page_title: "Converged Cloud: ccloud_billing_project_masterdata"
sidebar_current: "docs-ccloud-datasource-billing-project-masterdata"
description: |-
  Get information on the Billing Project Masterdata
---

# ccloud\_billing\_project\_masterdata

Use this data source to get the Billing Project Masterdata.

~> **Note:** You _must_ have admin privileges in your OpenStack cloud to use
this resource for other tenant projects.

## Example Usage

```hcl
data "ccloud_billing_project_masterdata" "masterdata" {
  project_id = "30dd31bcac8748daaa75720dab7e019a"
}

output "cost_object" {
  value = "${data.ccloud_billing_project_masterdata.masterdata.cost_object}"
}
```

## Argument Reference

The following arguments are supported:

* `region` - (Optional) The region in which to obtain the Billing client. If
  omitted, the `region` argument of the provider is used. Changing this forces
  a new resource to be created.

* `project_id` - (Optional) A project ID. Available only for users with an
  admin access. Defaults to the current project scope.

## Attributes Reference

In addition to arguments above, an extra attributes are exported. Please refer
to the `ccloud_billing_project_masterdata` resource arguments and attributes
[documentation](../r/billing_project_masterdata.html) for more information.
