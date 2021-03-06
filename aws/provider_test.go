package aws

import (
	"log"
	"os"
	"testing"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	"github.com/terraform-providers/terraform-provider-template/template"
	"github.com/terraform-providers/terraform-provider-tls/tls"
)

var testAccProviders map[string]terraform.ResourceProvider
var testAccProvidersWithTLS map[string]terraform.ResourceProvider
var testAccProvider *schema.Provider
var testAccTemplateProvider *schema.Provider

func init() {
	testAccProvider = Provider().(*schema.Provider)
	testAccTemplateProvider = template.Provider().(*schema.Provider)
	testAccProviders = map[string]terraform.ResourceProvider{
		"aws":      testAccProvider,
		"template": testAccTemplateProvider,
	}
	testAccProvidersWithTLS = map[string]terraform.ResourceProvider{
		"tls": tls.Provider(),
	}

	for k, v := range testAccProviders {
		testAccProvidersWithTLS[k] = v
	}
}

func TestProvider(t *testing.T) {
	if err := Provider().(*schema.Provider).InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvider_impl(t *testing.T) {
	var _ terraform.ResourceProvider = Provider()
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("AWS_PROFILE"); v == "" {
		if v := os.Getenv("AWS_ACCESS_KEY_ID"); v == "" {
			t.Fatal("AWS_ACCESS_KEY_ID must be set for acceptance tests")
		}
		if v := os.Getenv("AWS_SECRET_ACCESS_KEY"); v == "" {
			t.Fatal("AWS_SECRET_ACCESS_KEY must be set for acceptance tests")
		}
	}

	region := testAccGetRegion()
	log.Printf("[INFO] Test: Using %s as test region", region)
	os.Setenv("AWS_DEFAULT_REGION", region)

	err := testAccProvider.Configure(terraform.NewResourceConfig(nil))
	if err != nil {
		t.Fatal(err)
	}
}

func testAccGetRegion() string {
	v := os.Getenv("AWS_DEFAULT_REGION")
	if v == "" {
		return "us-west-2"
	}
	return v
}

func testAccEC2ClassicPreCheck(t *testing.T) {
	client := testAccProvider.Meta().(*AWSClient)
	platforms := client.supportedplatforms
	region := client.region
	if !hasEc2Classic(platforms) {
		t.Skipf("This test can only run in EC2 Classic, platforms available in %s: %q",
			region, platforms)
	}
}

func testAccAwsRegionProvider(region string, providers *[]*schema.Provider) *schema.Provider {
	if region == "" || providers == nil {
		return nil
	}
	log.Printf("[DEBUG] Checking providers for AWS region: %s", region)
	for _, provider := range *providers {
		// Ignore if Meta is empty, this can happen for validation providers
		if provider == nil || provider.Meta() == nil {
			log.Printf("[DEBUG] Skipping empty provider")
			continue
		}

		// Ignore if Meta is not AWSClient, this will happen for other providers
		client, ok := provider.Meta().(*AWSClient)
		if !ok {
			log.Printf("[DEBUG] Skipping non-AWS provider")
			continue
		}

		clientRegion := client.region
		log.Printf("[DEBUG] Checking AWS provider region %q against %q", clientRegion, region)
		if clientRegion == region {
			log.Printf("[DEBUG] Found AWS provider with region: %s", region)
			return provider
		}
	}
	return nil
}
