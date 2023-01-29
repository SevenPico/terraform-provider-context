package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestEnabled(t *testing.T) {
	const tfConfig = `
	data "context" "default" {
	}

	data "context" "base_true" {
	  	enabled = true
	}

	data "context" "base_false" {
		enabled = false
	}

	data "context" "inherit_true" {
		context = data.context.base_true
	}

	data "context" "inherit_false" {
		context = data.context.base_false
	}

	data "context" "inherit_true_override_false" {
		context = data.context.base_true
		enabled = false
	}

	data "context" "inherit_false_override_true" {
		context = data.context.base_false
		enabled = true
	}
	`

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: tfConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.context.default", "enabled", "true"),
					resource.TestCheckResourceAttr("data.context.base_true", "enabled", "true"),
					resource.TestCheckResourceAttr("data.context.base_false", "enabled", "false"),
					resource.TestCheckResourceAttr("data.context.inherit_true", "enabled", "true"),
					resource.TestCheckResourceAttr("data.context.inherit_false", "enabled", "false"),
					resource.TestCheckResourceAttr("data.context.inherit_true_override_false", "enabled", "false"),
					resource.TestCheckResourceAttr("data.context.inherit_false_override_true", "enabled", "false"),
				),
			},
		},
	})
}

func TestDescriptors(t *testing.T) {
	const tfConfig = `
	data "context" "base" {
		attributes_map = {
			project = "foo"
			stage   = "dev"
			env     = "bar"
			domain  = "example.com"
		}

		descriptors = {
			id = {
				order     = ["stage", "env", "project", "attributes"]
				delimiter = "-"
				upper     = false
				lower     = true
				title     = false
				reverse   = false
				limit     = 64
			}
			namespace = {
				order     = ["stage", "env", "project"]
				delimiter = "-"
				upper     = false
				lower     = true
				title     = false
				reverse   = false
				limit     = 64
			}
			fqdn = {
				order     = ["attributes", "stage", "env", "project", "domain"]
				delimiter = "."
				upper     = false
				lower     = true
				title     = false
				reverse   = false
				limit     = 64
			}
			fqdn2_prefix = {
				order     = ["attributes"]
				delimiter = "-"
			}
			fqdn2 = {
				order     = ["$fqdn2_prefix", "domain"]
				delimiter = "."
			}
			fqdn3_prefix = {
				order     = ["attributes"]
				delimiter = "."
				reverse   = true
			}
			fqdn3 = {
				order     = ["$fqdn3_prefix", "domain"]
				delimiter = "."
			}
		}
	}

	data "context" "fizz" {
		context = data.context.base
		attributes = ["fizz"]
	}

	data "context" "buzz" {
		context = data.context.fizz
		attributes = ["buzz"]
	}

	data "context" "too_long" {
		attributes = ["buzzzzzzzzzzzzzzz"]
		descriptors = {
			id = {
				order = ["attributes"]
				limit = 10
			}
		}
	}
	`

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: tfConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.context.base", "id", "dev-bar-foo"),
					resource.TestCheckResourceAttr("data.context.base", "namespace", "dev-bar-foo"),
					resource.TestCheckResourceAttr("data.context.base", "descriptors.fqdn.value", "dev.bar.foo.example.com"),
					resource.TestCheckResourceAttr("data.context.base", "descriptors.fqdn2.value", "example.com"),
					resource.TestCheckResourceAttr("data.context.base", "descriptors.fqdn3.value", "example.com"),

					resource.TestCheckResourceAttr("data.context.fizz", "id", "dev-bar-foo-fizz"),
					resource.TestCheckResourceAttr("data.context.fizz", "namespace", "dev-bar-foo"),
					resource.TestCheckResourceAttr("data.context.fizz", "descriptors.fqdn.value", "fizz.dev.bar.foo.example.com"),
					resource.TestCheckResourceAttr("data.context.fizz", "descriptors.fqdn2.value", "fizz.example.com"),
					resource.TestCheckResourceAttr("data.context.fizz", "descriptors.fqdn3.value", "fizz.example.com"),

					resource.TestCheckResourceAttr("data.context.buzz", "id", "dev-bar-foo-fizz-buzz"),
					resource.TestCheckResourceAttr("data.context.buzz", "namespace", "dev-bar-foo"),
					resource.TestCheckResourceAttr("data.context.buzz", "descriptors.fqdn.value", "fizz.buzz.dev.bar.foo.example.com"),
					resource.TestCheckResourceAttr("data.context.buzz", "descriptors.fqdn2.value", "fizz-buzz.example.com"),
					resource.TestCheckResourceAttr("data.context.buzz", "descriptors.fqdn3.value", "buzz.fizz.example.com"),

					resource.TestCheckResourceAttr("data.context.too_long", "id", "buzzzzzzzz"),
				),
			},
		},
	})
}
