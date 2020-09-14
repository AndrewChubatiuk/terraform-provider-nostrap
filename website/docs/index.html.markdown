---
layout: "nostrap"
page_title: "Provider: Nomad/Consul Bootstrap Provider"
sidebar_current: "docs-nostrap-index"
description: |-
  The Nomad/Consul bootstrap provider is used to bootstrap ACL in Nomad/Consul cluster.
---

# Nomad/Consul Bootstrap Provider

The Nomad/Consul bootstrap provider is used to bootstrap ACL in Nomad/Consul cluster.

Use the navigation to the left to read about the available resources.

## Example Usage

```hcl
provider "nostrap" {
}

resource "nostrap_acl_token" "acl" {
   address    = "http://localhost:4646
   ssm_prefix = "/some/ssm/prefix"
   aws_region = "us-east-1"
}
```
