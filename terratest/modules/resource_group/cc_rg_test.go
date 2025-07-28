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

const subscriptionID = "abd34832-7708-43f9-a480-e3b7a87b41d7" // Change this to your actual sub ID

func TestGenericCcRgValidation(t *testing.T) {
	// Load configuration file
	configPath := filepath.Join("..", "..", "configs", "cc_rg_test_config.json")
	cfg, err := common.LoadRgTestConfig(configPath)
	require.NoError(t, err)

	// Create Azure SDK clients
	cred, _ := azidentity.NewDefaultAzureCredential(nil)
	rgClient, _ := armresources.NewResourceGroupsClient(subscriptionID, cred, nil)
	lockClient, _ := armlocks.NewManagementLocksClient(subscriptionID, cred, nil)
	roleClient, _ := armauthorization.NewRoleAssignmentsClient(subscriptionID, cred, nil)

	// Loop through each RG defined in config
	for rgName, expected := range cfg.ResourceGroups {
		t.Run("Validate_"+rgName, func(t *testing.T) {
			// Get RG info
			rgResp, err := rgClient.Get(context.Background(), rgName, nil)
			require.NoError(t, err)

			rg := rgResp.ResourceGroup
			assert.Equal(t, expected.Location, *rg.Location)
			assert.Regexp(t, regexp.MustCompile(`^/subscriptions/.+/resourceGroups/.+`), *rg.ID)

			for k, v := range expected.Tags {
				actual, ok := rg.Tags[k]
				require.True(t, ok)
				assert.Equal(t, v, *actual)
			}

			// Validate locks
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

	// Test 1: Principal must be assigned one of expected roles
	t.Run("Check_RBAC", func(t *testing.T) {
		for rgName := range cfg.ResourceGroups {
			scope := fmt.Sprintf("/subscriptions/%s/resourceGroups/%s", subscriptionID, rgName)
			pager := roleClient.NewListForScopePager(scope, nil)

			found := false

			fmt.Printf("\n--- Checking RBAC for Resource Group: %s ---\n", rgName)
			fmt.Printf("Expected Principal ID: %s\n", cfg.ExpectedPrincipalID)
			fmt.Printf("Expected Role Names: %v\n", cfg.ExpectedRoles)

			for pager.More() {
				page, _ := pager.NextPage(context.Background())
				for _, role := range page.Value {
					actualPID := *role.Properties.PrincipalID
					roleDefID := *role.Properties.RoleDefinitionID
					fmt.Printf("Found assignment - PrincipalID: %s | RoleDefID: %s\n", actualPID, roleDefID)

					if actualPID == cfg.ExpectedPrincipalID {
						for _, roleName := range cfg.ExpectedRoles {
							if strings.Contains(strings.ToLower(roleDefID), strings.ToLower(roleName)) {
								fmt.Printf("✅ Match found: PrincipalID %s has role %s\n", actualPID, roleName)
								found = true
								break
							}
						}
					}
				}
			}
			assert.True(t, found, "Expected principal not assigned in the "+rgName)
		}
	})

	// Test 2: Ensure at least one assignment exists in each RG
	t.Run("RBAC_HasAtLeastOneAssignment", func(t *testing.T) {
		for rgName := range cfg.ResourceGroups {
			scope := fmt.Sprintf("/subscriptions/%s/resourceGroups/%s", subscriptionID, rgName)
			pager := roleClient.NewListForScopePager(scope, nil)

			count := 0
			for pager.More() {
				page, _ := pager.NextPage(context.Background())
				count += len(page.Value)
			}
			assert.Greater(t, count, 0, fmt.Sprintf("No RBAC assignments found for RG %s", rgName))
		}
	})

	// Test 3: Ensure current principal is assigned to all RGs
	t.Run("RBACContainsCurrentPrincipal", func(t *testing.T) {
		for rgName := range cfg.ResourceGroups {
			scope := fmt.Sprintf("/subscriptions/%s/resourceGroups/%s", subscriptionID, rgName)
			pager := roleClient.NewListForScopePager(scope, nil)

			found := false
			for pager.More() {
				page, _ := pager.NextPage(context.Background())
				for _, role := range page.Value {
					if *role.Properties.PrincipalID == cfg.ExpectedPrincipalID {
						found = true
						break
					}
				}
			}
			assert.True(t, found, fmt.Sprintf("Current principal %s not assigned to RG %s", cfg.ExpectedPrincipalID, rgName))
		}
	})

	// Test 4: Ensure principal has Contributor role (based on fixed ID)
	t.Run("CurrentPrincipalMustBeContributor", func(t *testing.T) {
		const contributorID = "b24988ac-6180-42a0-ab88-20f7382dd24c"

		for rgName := range cfg.ResourceGroups {
			scope := fmt.Sprintf("/subscriptions/%s/resourceGroups/%s", subscriptionID, rgName)
			pager := roleClient.NewListForScopePager(scope, nil)

			found := false
			for pager.More() {
				page, _ := pager.NextPage(context.Background())
				for _, role := range page.Value {
					pid := *role.Properties.PrincipalID
					rid := strings.ToLower(*role.Properties.RoleDefinitionID)
					if pid == cfg.ExpectedPrincipalID && strings.HasSuffix(rid, contributorID) {
						found = true
						break
					}
				}
			}
			assert.True(t, found, fmt.Sprintf("Principal %s must have Contributor role in RG %s", cfg.ExpectedPrincipalID, rgName))
		}
	})
}
