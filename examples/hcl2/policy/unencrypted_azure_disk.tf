resource "azurerm_managed_disk" "sample" {
  encryption_settings {
    enabled = false
  }
}
