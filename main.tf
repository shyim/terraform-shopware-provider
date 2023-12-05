provider "shopware" {
  url           = "http://localhost:8000"
  client_id     = "SWIADMXJMKPYSWRNCGZXTVFZSW"
  client_secret = "NjdteXQzUWZWN3RUcVAxaFVtWXN5Y3Fkbm9EMUpwVEE1dEhjWjM"
}

resource "shopware_delivery_time" "test" {
  name    = "Terraform Delivery Time"
  unit    = "days"
  minimum = 1
  maximum = 5
}

resource "shopware_rule" "my_rule" {
  name     = "Terraform Rule"
  priority = 1
  type     = ["shipping"]

  conditions {
    type  = "customerAffiliateCode"
    value = {
      operator = "="
      affiliateCode    = "123"
    }
  }
}

resource "shopware_shipping_method" "bla" {
  name                 = "fooo"
  technical_name       = "fooo"
  active               = true
  delivery_time_id     = shopware_delivery_time.test.id
  availability_rule_id = shopware_rule.my_rule.id
}