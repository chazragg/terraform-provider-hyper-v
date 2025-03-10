terraform {
  required_providers {
    hyperv = {
      source = "registry.terraform.io/chazragg/hyperv"
    }
  }
}

provider "hyperv" {
  host     = "192.168.0.12"
  username = "test"
  password = "password"
  port = 5985
}

resource "hyperv_vm" "name" {
  name = "testVM"
  generation = 2
  memory_startup = 536870912
  path = "D:\\vhd"
  switch_name = "Default Switch"
}