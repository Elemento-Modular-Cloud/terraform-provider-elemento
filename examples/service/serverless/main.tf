terraform {
  required_providers {
    elemento = {
      source = "hashicorp/elemento"
    }
  }
}

provider "elemento" {
  target   = "aws-dev-private"
  username = "user@elemento.cloud"
  password = "oassword"
}

resource "elemento_serverless" "my_serverless" {
  name   = "my-serverless"
  region = "eu-south-1"
  runtime = "python3.13"
  handler = "main.handler"
  memory = 512
  timeout = 30
  is_public = true
  environment = {
    STAGE = "test"
  }
  code_zip_b64 = "UEsDBBQAAAAIAHJRFV2KSnlhqQAAAAgBAAAHABwAbWFpbi5weVVUCQADWCSIalgkiGp1eAsAAQQAAAAABAAAAABljrEOwjAMRPd+xZGJShFCjF1Z2PkCNzEkUokhdQsV4t8pAQkhvFi+s9/Z8wGBku94ySMntXCSlG9aNxXmMsaUXswGPjotGxRTTEdkvgzcKzwpYXliDeItzqTBIjB5zr1FK36yYHWrurA+Cf+0+Z4KiVoZFBoYMY3iSKOkn3cy65AT7mUoRq+kQ78Vz6bBZr22X+uVP4tmx10nOGQ5Ye+o4ytNizfuUT0BUEsBAh4DFAAAAAgAclEVXYpKeWGpAAAACAEAAAcAGAAAAAAAAQAAAKSBAAAAAG1haW4ucHlVVAUAA1gkiGp1eAsAAQQAAAAABAAAAABQSwUGAAAAAAEAAQBNAAAA6gAAAAAA"
}

# output "payment_url" {
#   description = "Payment URL for the created resource"
#   value       = elemento_objectstorage.my_bucket.payment_url
# }
