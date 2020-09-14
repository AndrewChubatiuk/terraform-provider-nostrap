package main

import (
	"github.com/AndrewChubatiuk/terraform-provider-nostrap/provider"
	"github.com/hashicorp/terraform/plugin"
)

func main() {
	plugin.Serve(
		&plugin.ServeOpts{
			ProviderFunc: provider.Provider})
}
