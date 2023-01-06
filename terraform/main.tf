provider "aws" {
  region = local.region
}

locals {
  region = "us-east-1"
  deployment_name = "${var.app_name}-${var.environment}"
  tags = merge(
    {
      "Name" = local.deployment_name
      "Environment" = var.environment
      "Application" = var.app_name
    },
    var.tags
  )
}
