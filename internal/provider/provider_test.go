// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// testAccProtoV6ProviderFactories is used to instantiate a provider during acceptance testing.
// The factory function is called for each Terraform CLI command to create a provider
// server that the CLI can connect to and interact with.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"elemento": providerserver.NewProtocol6WithError(New("test")()),
}

func TestProviderServer(t *testing.T) {
	t.Parallel()

	factory, ok := testAccProtoV6ProviderFactories["elemento"]
	if !ok {
		t.Fatal("missing elemento provider factory")
	}

	server, err := factory()
	if err != nil {
		t.Fatalf("create provider server: %s", err)
	}
	if server == nil {
		t.Fatal("expected a provider server")
	}
}
