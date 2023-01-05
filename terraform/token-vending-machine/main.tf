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


#// -- chatgpt crap
#
#provider "aws" {
#  alias  = "us-west-2"
#  region = "us-west-2"
#}
#
#resource "aws_lambda_function" "example" {
#  filename         = "function.zip"
#  function_name    = "example"
#  role             = "arn:aws:iam::123456789012:role/lambda_basic_execution"
#  handler          = "example"
#  runtime          = "go1.x"
#  source_code_hash = filebase64sha256("function.zip")
#
#  environment {
#    variables = {
#      foo = "bar"
#    }
#  }
#}
#
#resource "aws_iam_policy" "api_gateway_execution_policy" {
#  name        = "api_gateway_execution_policy"
#  path        = "/"
#  description = "Policy for API Gateway execution role"
#
#  policy = <<EOF
#{
#  "Version": "2012-10-17",
#  "Statement": [
#    {
#      "Effect": "Allow",
#      "Action": [
#        "logs:CreateLogGroup",
#        "logs:CreateLogStream",
#        "logs:PutLogEvents"
#      ],
#      "Resource": "arn:aws:logs:*:*:*"
#    },
#    {
#      "Effect": "Allow",
#      "Action": [
#        "lambda:InvokeFunction"
#      ],
#      "Resource": "*"
#    }
#  ]
#}
#EOF
#}
#
#resource "aws_iam_role" "api_gateway_execution_role" {
#  name = "api_gateway_execution_role"
#
#  assume_role_policy = <<EOF
#{
#  "Version": "2012-10-17",
#  "Statement": [
#    {
#      "Effect": "Allow",
#      "Principal": {
#        "Service": "apigateway.amazonaws.com"
#      },
#      "Action": "sts:AssumeRole"
#    }
#  ]
#}
#EOF
#}
#
#resource "aws_iam_role_policy_attachment" "api_gateway_execution_policy_attachment" {
#  role       = aws_iam_role.api_gateway_execution_role.name
#  policy_arn = aws_iam_policy.api_gateway_execution_policy.arn
#}
#
#
#resource "aws_api_gateway_rest_api" "example" {
#  name = "example"
#}
#
#resource "aws_api_gateway_resource" "example" {
#  rest_api_id = aws_api_gateway_rest_api.example.id
#  parent_id   = aws_api_gateway_rest_api.example.root_resource_id
#  path_part   = "example"
#}
#resource "aws_api_gateway_method" "example" {
#  rest_api_id   = aws_api_gateway_rest_api.example.id
#  resource_id   = aws_api_gateway_resource.example.id
#  http_method   = "POST"
#  authorization = "NONE"
#}
#resource "aws_api_gateway_integration" "example" {
#  rest_api_id = aws_api_gateway_rest_api.example.id
#  resource_id = aws_api_gateway_resource.example.id
#  http_method = aws_api_gateway_method.example.http_method
#
#  type                    = "AWS_PROXY"
#  uri                     = "arn:aws:apigateway:${local.region}:lambda:path/2015-03-31/functions/${aws_lambda_function.example.arn}/invocations"
#  integration_http_method = "POST"
#}
#
#resource "aws_api_gateway_deployment" "example" {
#  rest_api_id = aws_api_gateway_rest_api.example.id
#  stage_name  = "prod"
#}
#
#resource "aws_api_gateway_domain_name" "example" {
#  domain_name = "example.com"
#
#  certificate_arn = "arn:aws:acm:us-east-1:123456789012:certificate/12345678-1234-1234-1234-123456789012"
#}
#resource "aws_api_gateway_basepath_mapping" "example" {
#  domain_name = aws_api_gateway_domain_name.example.domain_name
#  stage_name  = aws_api_gateway_deployment.example.stage_name
#  rest_api_id = aws_api_gateway_rest_api.example.id
#  base_path   = "example"
#}
#
#resource "aws_acm_certificate" "example" {
#  domain_name       = "example.com"
#  validation_method = "DNS"
#}
#resource "aws_cloudfront_distribution" "example" {
#  enabled         = true
#  is_ipv6_enabled = true
#
#  default_cache_behavior {
#    target_origin_id       = "api_gateway"
#    viewer_protocol_policy = "redirect-to-https"
#    allowed_methods        = ["GET", "HEAD"]
#    cached_methods         = ["GET", "HEAD"]
#    min_ttl                = 0
#    default_ttl            = 3600
#    max_ttl                = 86400
#  }
#
#  custom_error_response {
#    error_caching_min_ttl = 600
#    error_code            = 404
#    response_code         = 404
#    response_page_path    = "/error-pages/404.html"
#  }
#
#  custom_error_response {
#    error_caching_min_ttl = 600
#    error_code            = 500
#    response_code         = 500
#    response_page_path    = "/error-pages/500.html"
#  }
#
#  origins {
#    type        = "custom"
#    domain_name = "execute-api.${local.region}.amazonaws.com"
#    origin_path = "/${aws_api_gateway_rest_api.example.id}"
#    custom_header {
#      name  = "Host"
#      value = "execute-api.${local.region}.amazonaws.com"
#    }
#  }
#
#  viewer_certificate {
#    acm_certificate_arn = aws_acm_certificate.example.arn
#    ssl_support_method  = "sni-only"
#  }
#}
#
