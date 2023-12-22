
terraform {
  required_providers {
    pingdom = {
      source = "registry.terraform.io/karlderkaefer/pingdom"
    }
  }
}

provider "pingdom" {}


data "pingdom_transaction_checks" "example" {
}
