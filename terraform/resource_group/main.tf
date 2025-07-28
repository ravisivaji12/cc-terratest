terraform {
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = ">= 3.0"
    }
  }
}

provider "azurerm" {
  features {}
  subscription_id = "abd34832-7708-43f9-a480-e3b7a87b41d7"
}

variable "resource_groups" {
  type = map(object({
    location = string
    tags     = map(string)
    lock = optional(object({
      level = string
      notes = optional(string)
    }))
  }))
}

resource "azurerm_resource_group" "rg" {
  for_each = var.resource_groups
  name     = each.key
  location = each.value.location
  tags     = each.value.tags
}

resource "azurerm_management_lock" "rg_lock" {
  for_each = {
    for k, v in var.resource_groups : k => v.lock if v.lock != null
  }

  name       = "${each.key}-lock"
  scope      = azurerm_resource_group.rg[each.key].id
  lock_level = each.value.level
  notes      = each.value.notes != null ? each.value.notes : ""
}
