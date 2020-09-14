---
layout: "nostrap"
page_title: "Nomad/Consul Bootstrap: nostrap_acl_token"
sidebar_current: "docs-nostrap-resource-acl-token"
description: |-
  Bootstraps ACL for Nomad/Consul 
---

# nostrap_acl_token

This resource bootstraps initial ACL token for Nomad/Consul and puts it to AWS SSM.

## Example Usage

```hcl
resource "nostrap_acl_token" "acl" {
   address    = "http://localhost:4646
   ssm_prefix = "/some/ssm/prefix"
   aws_region = "us-east-1"
}
```

## Argument Reference

The following arguments are supported:

* `address` - (Required) Nomad/Consul cluster address.
* `ssm_prefix` - (Required) AWS SSM prefix.
* `aws_region` - (Required) AWS SSM Region

## Attribute Reference

* `ancestor_id` - ACL Token ancestor ID
* `secret_id` - ACL Token secret ID
