Converged Cloud Enterprise Edition - Terraform Provider
=======================================================

- Website: https://registry.terraform.io/providers/sapcc/ccloud/latest/docs

<img src="https://cdn.rawgit.com/hashicorp/terraform-website/master/content/source/assets/images/logo-hashicorp.svg" width="600px">

Maintainers
-----------

This provider plugin is maintained by:

  * Anton Khramov (@kayrus)
  * Michael Schmidt (@bugroger)

Requirements
------------

-	[Terraform](https://www.terraform.io/downloads.html) 0.14.x
-	[Go](https://golang.org/doc/install) 1.15 (to build the provider plugin)

Usage
---------------------


The CCloudEE provider is an extension to the [OpenStack Terraform
Provider](https://github.com/terraform-provider-openstack/terraform-provider-openstack).
It provides resources that allow to use Terraform for Converged Cloud's
additional services:

  * Limes for Quota Management
  * Kubernikus (Kubernetes as a Service)
  * Arc for Arc resources management
  * Lyra for Automation management
  * Billing for Billing management

The provider needs to be configured with the proper OpenStack credentials
before it can be used. For details see the OpenStack provider.

Building The Provider
---------------------

Clone the repository

```sh
$ git clone git@github.com:sapcc/terraform-ccloud-provider
```

Enter the provider directory and build the provider

```sh
$ cd terraform-ccloud-provider
$ make build
```

Installing the provider
-----------------------

To install this provider, copy and paste this code into your Terraform configuration. Then, run `terraform init`.

```hcl
terraform {
  required_providers {
    ccloud = {
      source = "sapcc/ccloud"
    }
  }
}

provider "ccloud" {
  # Configuration options
}
```

Using the provider
----------------------
You can browse the documentation [here](https://registry.terraform.io/providers/sapcc/ccloud/latest/docs).

Developing the Provider
---------------------------

If you wish to work on the provider, you'll first need [Go](https://golang.org) installed on your machine (version 1.15+ is *required*).

To compile the provider, run `make build`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

```sh
$ make build
```

In order to test the provider, you can simply run `make test`.

```sh
$ make test
```
