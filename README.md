Converged Cloud Enterprise Edition - Terraform Provider
=======================================================

- Website: https://www.terraform.io
- [![Gitter chat](https://badges.gitter.im/hashicorp-terraform/Lobby.png)](https://gitter.im/hashicorp-terraform/Lobby)
- Mailing list: [Google Groups](http://groups.google.com/group/terraform-tool)

<img src="https://cdn.rawgit.com/hashicorp/terraform-website/master/content/source/assets/images/logo-hashicorp.svg" width="600px">

Maintainers
-----------

This provider plugin is maintained by:

  * Anton Khramov (@kayrus)
  * Michael Schmidt (@bugroger)

Requirements
------------

-	[Terraform](https://www.terraform.io/downloads.html) 0.12.x
-	[Go](https://golang.org/doc/install) 1.12 (to build the provider plugin)

Usage
---------------------


The CCloudEE provider is an extension to the [OpenStack Terraform
Provider](https://github.com/terraform-providers/terraform-provider-openstack).
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

Clone repository to: `$GOPATH/src/github.com/sapcc/terraform-ccloud-provider`

```sh
$ mkdir -p $GOPATH/src/github.com/sapcc; cd $GOPATH/src/github.com/sapcc
$ git clone git@github.com:sapcc/terraform-ccloud-provider
```

Enter the provider directory and build the provider

```sh
$ cd $GOPATH/src/github.com/sapcc/terraform-ccloud-provider
$ make build
```

Installing the provider
-----------------------
Download a [binary release](https://github.com/sapcc/terraform-provider-ccloud/releases) for your OS, unpack the archive and follow the official Terraform documentation to know how to [install third-party providers](https://www.terraform.io/docs/configuration/providers.html#third-party-plugins).

Using the provider
----------------------
You can browse the documentation within this repo [here](https://github.com/sapcc/terraform-provider-ccloud/tree/master/website/docs).

Developing the Provider
---------------------------

If you wish to work on the provider, you'll first need [Go](https://golang.org) installed on your machine (version 1.12+ is *required*). You'll also need to correctly setup a [GOPATH](https://golang.org/doc/code.html#GOPATH), as well as adding `$GOPATH/bin` to your `$PATH`.

To compile the provider, run `make build`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

```sh
$ make build
...
$ $GOPATH/bin/terraform-provider-ccloud
...
```

In order to test the provider, you can simply run `make test`.

```sh
$ make test
```
