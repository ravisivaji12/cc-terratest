package resource_group

import (
	"context"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/ravisivaji12/cc-terratest/terratest/common"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/authorization/armauthorization"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armlocks"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

const subscriptionID = "<SUBSCRIPTION_ID>"

func TestGenericCcRgValidation(t *testing.T) {
	configPath := filepath.Join("..", "..", "configs", "cc_rg_test_config.json")
	cfg, err := common.LoadRgTestConfig(configPath)
	require.NoError(t, err)

	cred, _ := azidentity.NewDefaultAzureCredential(nil)
	rgClient, _ := armresources.NewResourceGroupsClient(subscriptionID, cred, nil)
	lockClient, _ := armlocks.NewManagementLocksClient(subscriptionID, cred, nil)
	roleClient, _ := armauthorization.NewRoleAssignmentsClient(subscriptionID, cred, nil)

	for rgName, expected := range cfg.ResourceGroups {
		t.Run("Validate_"+rgName, func(t *testing.T) {
			rgResp, err := rgClient.Get(context.Background(), rgName, nil)
			require.NoError(t, err)

			rg := *rgResp.ResourceGroup
			assert.Equal(t, expected.Location, *rg.Location)
			assert.Regexp(t, regexp.MustCompile(`^/subscriptions/.+/resourceGroups/.+`), *rg.ID)

			for k, v := range expected.Tags {
				actual, ok := rg.Tags[k]
				require.True(t, ok)
				assert.Equal(t, v, *actual)
			}

			lockPager := lockClient.NewListAtResourceGroupLevelPager(rgName, nil)
			foundLock := ""
			for lockPager.More() {
				page, _ := lockPager.NextPage(context.Background())
				for _, l := range page.Value {
					foundLock = string(*l.Properties.Level)
				}
			}

			if expected.Lock == nil {
				assert.True(t, foundLock == "")
			} else {
				assert.Equal(t, expected.Lock.Level, foundLock)
			}
		})
	}

	t.Run("Check_RBAC", func(t *testing.T) {
		for rgName := range cfg.ResourceGroups {
			scope := fmt.Sprintf("/subscriptions/%s/resourceGroups/%s", subscriptionID, rgName)
			pager := roleClient.NewListForScopePager(scope, nil)

			found := false
			for pager.More() {
				page, _ := pager.NextPage(context.Background())
				for _, role := range page.Value {
					if *role.Properties.PrincipalID == cfg.ExpectedPrincipalID {
						for _, roleName := range cfg.ExpectedRoles {
							if strings.Contains(*role.Properties.RoleDefinitionID, strings.ToLower(roleName)) {
								found = true
								break
							}
						}
					}
				}
			}
			assert.True(t, found, "Expected principal not assigned in "+rgName)
		}
	})
}
