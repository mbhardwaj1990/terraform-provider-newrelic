//go:build integration
// +build integration

package newrelic

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccNewRelicGcpLinkAccount_Basic(t *testing.T) {
	rName := acctest.RandString(5)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccNewRelicGcpLinkAccountDestroy,
		Steps: []resource.TestStep{
			//Test: Create
			{
				Config: testAccNewRelicGcpLinkAccountConfig(rName),
			},
			//Test: Update
			//TODO
			{
				Config: testAccNewRelicGcpLinkAccountConfigUpdated(rName),
			},
		},
	})
}

func testAccNewRelicGcpLinkAccountDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*ProviderConfig).NewClient
	for _, r := range s.RootModule().Resources {
		if r.Type != "newrelic_gcp_link_account" {
			continue
		}
		resourceId, err := strconv.Atoi(r.Primary.ID)
		if err != nil {
			fmt.Errorf("unable to convert string to int")
		}
		_, err = client.Cloud.GetLinkedAccount(testAccountID, resourceId)
		if err != nil {
			return err
		}
	}
	return nil
}

func testAccNewRelicGcpLinkAccountConfig(name string) string {
	return fmt.Sprintf(`
	resource "newrelic_gcp_link_account" "gcp_account"{
		name = "%[1]s"
		project_id = ""
	}
	`, name)
}

func testAccNewRelicGcpLinkAccountConfigUpdated(name string) string {
	return fmt.Sprintf(`
	resource "newrelic_gcp_link_account" "gcp_account"{
		name = "%[1]s"
		project_id = ""
	}
	`, name)
}
