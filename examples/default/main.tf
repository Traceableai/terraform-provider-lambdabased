terraform {
  required_providers {
    lambdabased = {
      source  = "thetradedesk/lambdabased"
    }
  }
}

provider "lambdabased" {
  region = "us-east-1"
  # account = "123456789012"  # Optional: AWS account ID to assume role in
  # assume_role_name = "MyRoleName"  # Required when account is specified
}

locals {
  credentials = {
    creds = "temporary-creds" # For instance, this can a token acquired with aws_eks_cluster_auth
  }

  parameters = {
    param = "parameter-value"
  }

  destroy_parameters = {
    param = "parameter-destroy-value"
  }
}

resource "lambdabased_resource" "test" {
    function_name = "test-function"
    triggers = {
      param = sha512(jsonencode(local.parameters)) # drop sha512 if you want to store this in cleartext in the tf state
    }
    input = jsonencode(merge(local.credentials, local.parameters))
    conceal_input = true
    conceal_result = true
    finalizer {
        function_name = "test-function" # or another function if needed
        input = jsonencode(merge(local.credentials, local.destroy_parameters))
    }
}
