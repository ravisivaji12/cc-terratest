
---

# 🧪 `cc-terratest` — Azure Terratest Framework

This repository provides a **modular and reusable Terratest framework** built for validating Azure infrastructure using configuration files. It's designed to eliminate hardcoded test logic by using JSON-driven inputs and can be consumed as a Go module.

---

## 📁 Project Structure

```
cc-terratest/
├── terraform/                      # Infra-as-Code (optional if using external TF)
│   └── resource_group/
├── terratest/                      # Root of Go module for testing
│   ├── go.mod                      # Go module initialized here
│   ├── configs/                    # Input test configs (JSON)
│   ├── common/                     # Shared test helpers
│   │   └── load.go                 # Loads JSON config
│   └── modules/
│       └── resource_group/
│           └── cc_rg_test.go       # Resource Group validation logic
```

---

## 📦 Publishing as Go Package

The Go module is defined at `terratest/go.mod`, so the import path for other teams will be:

```go
github.com/<your-org>/cc-terratest/terratest/modules/resource_group
```

Example:

```go
import rgtest "github.com/ravisivaji12/cc-terratest/terratest/modules/resource_group"
```

---

## 🔧 Setup in Your Repository

### ✅ Step 1: Add as Dependency

In your consuming repo, run:

```bash
go get github.com/ravisivaji12/cc-terratest/terratest/modules/resource_group@latest
```

Or specify the version hash:

```bash
go get github.com/ravisivaji12/cc-terratest/terratest/modules/resource_group@v0.0.0-20250728100000-<commit-sha>
```

> 💡 Make sure your repo is **public** or you have proper **GitHub auth** for private access.

---

### ✅ Step 2: Create Your Own Test File

```go
package test

import (
    "testing"

    rgtest "github.com/ravisivaji12/cc-terratest/terratest/modules/resource_group"
)

func TestAzureRgValidation(t *testing.T) {
    rgtest.TestGenericCcRgValidation(t)
}
```

---

### ✅ Step 3: Add Test Config JSON

Place the following config in your repo (e.g., `terratest/configs/cc_rg_test_config.json`):

```json
{
  "expected_principal_id": "5d976b48-4340-4c05-ab09-e80ef3751794",
  "expected_roles": ["Contributor"],
  "resource_groups": {
    "cc-prod-rg": {
      "location": "canadacentral",
      "tags": {
        "env": "prod",
        "team": "cloud"
      },
      "lock": {
        "level": "CanNotDelete"
      }
    },
    "cc-dev-rg": {
      "location": "canadaeast",
      "tags": {
        "env": "dev"
      },
      "lock": null
    }
  }
}
```

---

## ✅ What This Framework Validates

For every Resource Group declared in the config, it will:

| Check                         | Description                                          |
| ----------------------------- | ---------------------------------------------------- |
| 📍 Location                   | Validates actual RG location matches expected        |
| 🏷️ Tags                      | Asserts all expected tags exist                      |
| 🔐 Locks                      | Validates expected lock level (`CanNotDelete`, etc.) |
| 🔑 RBAC Assignment            | Ensures current principal has correct role           |
| ✅ Contributor Role            | Ensures principal is **Contributor** specifically    |
| 📋 At Least One Role Assigned | Sanity check that RG has any RBAC assignments        |

---

## 🔑 Auth Setup (for Azure SDK)

This test uses `azidentity.NewDefaultAzureCredential()`:

* Works with:

  * Azure CLI login (`az login`)
  * Managed Identity (e.g., Azure DevOps agent)
  * Environment-based SP credentials

If using service principal:

```bash
export AZURE_CLIENT_ID=<app-id>
export AZURE_CLIENT_SECRET=<secret>
export AZURE_TENANT_ID=<tenant>
```

---

## ▶️ Running the Tests

From the root of your repo:

```bash
go test ./terratest/modules/resource_group -v
```

Make sure:

* `cc_rg_test_config.json` is available relative to the test file.
* You are authenticated with Azure.

---

## 🚀 Versioning and Tags

Use Git tags to version your module for consumption:

```bash
git tag v0.1.0
git push origin v0.1.0
```

Other teams can then `go get github.com/<you>/cc-terratest/terratest/modules/resource_group@v0.1.0`.

---

## 🧪 Sample Output (Test Logs)

```bash
PS C:\Official\Ravi.Sivaji\cc-terratest\terratest\modules\resource_group> go test -v
=== RUN   TestGenericCcRgValidation
=== RUN   TestGenericCcRgValidation/Validate_cc-prod-rg
=== RUN   TestGenericCcRgValidation/Validate_cc-dev-rg
=== RUN   TestGenericCcRgValidation/Check_RBAC

--- Checking RBAC for Resource Group: cc-prod-rg ---
Expected Principal ID: 5d976b48-****
Expected Role Names: ["/subscriptions/abd34832-******/providers/Microsoft.Authorization/roleDefinitions/b24988ac-42a0-ab88-******"]
Found assignment - PrincipalID: 5d976b48-4340-4c05-ab09-e80ef3751794 | RoleDefID: /subscriptions/abd34832-****/providers/Microsoft.Authorization/roleDefinitions/b24988ac-****
✅ Match found: PrincipalID 5d976b48-4340-4c05-ab09-e80ef3751794 has role /subscriptions/abd34832-****/providers/Microsoft.Authorization/roleDefinitions/b24988ac-****
Found assignment - PrincipalID: 8a500209-**** | RoleDefID: /subscriptions/abd34832-****/providers/Microsoft.Authorization/roleDefinitions/b24988ac-****

--- Checking RBAC for Resource Group: cc-dev-rg ---
Expected Principal ID: 5d976b48-****
Expected Role Names: ["/subscriptions/abd34832-******/providers/Microsoft.Authorization/roleDefinitions/b24988ac-******"]
Found assignment - PrincipalID: 5d976b48-4340-4c05-ab09-e80ef3751794 | RoleDefID: /subscriptions/abd34832-****/providers/Microsoft.Authorization/roleDefinitions/b24988ac-****
✅ Match found: PrincipalID 5d976b48-4340-4c05-ab09-e80ef3751794 has role /subscriptions/abd34832-****/providers/Microsoft.Authorization/roleDefinitions/b24988ac-****
Found assignment - PrincipalID: 8a500209-**** | RoleDefID: /subscriptions/abd34832-****/providers/Microsoft.Authorization/roleDefinitions/b24988ac-6180-42a0-ab88-20f7382dd24c

=== RUN   TestGenericCcRgValidation/RBAC_HasAtLeastOneAssignment
=== RUN   TestGenericCcRgValidation/RBACContainsCurrentPrincipal
=== RUN   TestGenericCcRgValidation/CurrentPrincipalMustBeContributor
--- PASS: TestGenericCcRgValidation (7.58s)
    --- PASS: TestGenericCcRgValidation/Validate_cc-prod-rg (3.61s)
    --- PASS: TestGenericCcRgValidation/Validate_cc-dev-rg (0.58s)
    --- PASS: TestGenericCcRgValidation/Check_RBAC (1.76s)
    --- PASS: TestGenericCcRgValidation/RBAC_HasAtLeastOneAssignment (0.68s)
    --- PASS: TestGenericCcRgValidation/RBACContainsCurrentPrincipal (0.45s)
    --- PASS: TestGenericCcRgValidation/CurrentPrincipalMustBeContributor (0.49s)
PASS
ok      github.com/ravisivaji12/cc-terratest/terratest/modules/resource_group   11.649s
```

---

## 💡 Tips

* Use `go mod tidy` after `go get`
* Commit `go.sum` and `go.mod`
* Structure test logic around expected input from config only
* Avoid hardcoding any TF outputs or state parsing
