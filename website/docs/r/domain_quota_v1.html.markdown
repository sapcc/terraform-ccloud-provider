---
layout: "ccloud"
page_title: "Converged Cloud: ccloud_domain_quota_v1"
sidebar_current: "docs-ccloud-resource-domain-quota-v1"
description: |-
  Manages Domain Quota Resources
---

# ccloud\_domain\_quota\_v1

Manages Limes (Quota) Domain resources.

~> **Note:** The `terraform destroy` command destroys the
`ccloud_domain_quota_v1` state, but not the actual Limes domain quota.

~> **Note:** You _must_ have admin privileges in your OpenStack cloud to use
this resource.

## Example Usage

```hcl
data "openstack_identity_project_v3" "demo" {
  name = "demo"
}

resource "ccloud_domain_quota_v1" "quota" {
  domain_id = "${openstack_identity_project_v3.demo.domain_id}"

  compute {
    instances = 8
    cores     = 32
    ram       = 81920
  }

  volumev2 {
    capacity  = 1024
    snapshots = 0
    volumes   = 128
  }

  network {
    floating_ips         = 4
    networks             = 1
    ports                = 512
    routers              = 2
    security_group_rules = 64
    security_groups      = 4
    subnets              = 1
    healthmonitors       = 10
    l7policies           = 8
    listeners            = 16
    loadbalancers        = 8
    pools                = 8
    pool_members         = 10
  }

  dns {
    zones      = 1
    recordsets = 16
  }

  sharev2 {
    share_networks    = 1
    share_capacity    = 1024
    shares            = 16
    snapshot_capacity = 512
    share_snapshots   = 8
  }

  objectstore {
    capacity = 1073741824
  }

  database {
    cfm_share_capacity = 1073741824
  }
}
```

## Argument Reference

The following arguments are supported:

* `region` - (Optional) The region in which to obtain the Limes client. If
  omitted, the `region` argument of the provider is used. Changing this forces
  a new resource to be created.

* `domain_id` – (Required) The ID of the domain to manage the quota. Changing
  this forces a new resource to be created.

* `compute` - (Optional) The list of compute resources quota. Consists of
  `cores`, `instances`, `ram` (Mebibytes), `server_groups` and
  `server_group_members`.

* `volumev2` - (Optional) The list of block storage resources quota. Consists of
  `capacity` (Gibibytes), `snapshots` and `volumes`.

* `network` - (Optional) The list of network resources quota. Consists of
  `floating_ips`, `networks`, `ports`, `rbac_policies`, `routers`,
  `security_group_rules`, `security_groups`, `subnet_pools`, `subnets`,
  `healthmonitors`, `l7policies`, `listeners`, `loadbalancers`, `pools` and
  `pool_members`.

* `dns` - (Optional) The list of DNS resources quota. Consists of `zones` and
  `recordsets`.

* `sharev2` - (Optional) The list of Shared File Systems resources quota.
  Consists of `share_networks`, `share_capacity` (Gibibytes), `shares`,
  `snapshot_capacity` (Gibibytes) and `share_snapshots`.

* `objectstore` - (Optional) The list of Object Storage resources quota.
  Consists of `capacity` (Bytes).

* `database` - (Optional) The list of CFM Storage resources quota. Consists of
  `cfm_share_capacity` (Bytes).

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The domain ID.

## Import

Limes Domain Quota can be imported using the domain ID as an argument, e.g.

```
$ terraform import ccloud_domain_quota_v1.demo bf2273b5-2926-4495-9fb7-f28c3abed5f6
```
