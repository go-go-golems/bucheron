resource "aws_api_gateway_rest_api" "api" {
  name = "api"
  tags = local.tags
}

resource "aws_api_gateway_resource" "token" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  parent_id   = aws_api_gateway_rest_api.api.root_resource_id
  path_part   = "token"
}

resource "aws_api_gateway_method" "token" {
  rest_api_id   = aws_api_gateway_rest_api.api.id
  resource_id   = aws_api_gateway_resource.token.id
  http_method   = "POST"
  authorization = "NONE"
}

resource "aws_api_gateway_integration" "token_vending_machine" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  resource_id = aws_api_gateway_resource.token.id
  http_method = aws_api_gateway_method.token.http_method

  type                    = "AWS"
  integration_http_method = "POST"
  uri                     = aws_lambda_function.this["token-vending-machine"].invoke_arn
}

resource "aws_api_gateway_resource" "upload" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  parent_id   = aws_api_gateway_rest_api.api.root_resource_id
  path_part   = "upload"
}

resource "aws_api_gateway_method" "upload" {
  rest_api_id   = aws_api_gateway_rest_api.api.id
  resource_id   = aws_api_gateway_resource.upload.id
  http_method   = "POST"
  authorization = "NONE"
}

resource "aws_api_gateway_integration" "log_upload" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  resource_id = aws_api_gateway_resource.upload.id
  http_method = aws_api_gateway_method.upload.http_method

  type                    = "AWS"
  integration_http_method = "POST"
  uri                     = aws_lambda_function.this["log-upload"].invoke_arn
}

resource "aws_api_gateway_deployment" "v1" {
  depends_on = [
    aws_api_gateway_integration.token_vending_machine,
    aws_api_gateway_integration.log_upload,
  ]

  rest_api_id = aws_api_gateway_rest_api.api.id
  stage_name  = "v1"
}

resource "aws_lambda_permission" "token_vending_machine" {
  statement_id  = "AllowAPIGatewayInvoke"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.this["token-vending-machine"].arn
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_api_gateway_rest_api.api.execution_arn}/*/*"
}

resource "aws_lambda_permission" "log_upload" {
  statement_id  = "AllowAPIGatewayInvoke"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.this["log-upload"].arn
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_api_gateway_rest_api.api.execution_arn}/*/*"
}

resource "aws_api_gateway_method_response" "token_response_200" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  resource_id = aws_api_gateway_resource.token.id
  http_method = aws_api_gateway_method.token.http_method
  status_code = "200"
}

resource "aws_api_gateway_method_response" "upload_response_200" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  resource_id = aws_api_gateway_resource.upload.id
  http_method = aws_api_gateway_method.upload.http_method
  status_code = "200"
}

resource "aws_api_gateway_integration_response" "token_integration_response_200" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  resource_id = aws_api_gateway_resource.token.id
  http_method = aws_api_gateway_method.token.http_method
  status_code = aws_api_gateway_method_response.token_response_200.status_code
}

resource "aws_api_gateway_integration_response" "upload_integration_response_200" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  resource_id = aws_api_gateway_resource.upload.id
  http_method = aws_api_gateway_method.upload.http_method
  status_code = aws_api_gateway_method_response.upload_response_200.status_code
}


#resource "aws_api_gateway_domain_name" "token_vending_machine" {
#  domain_name = "token_vending_machine.com"
#
#  certificate_arn = "arn:aws:acm:us-east-1:123456789012:certificate/12345678-1234-1234-1234-123456789012"
#}
#
#resource "aws_api_gateway_base_path_mapping" "token_vending_machine" {
#  domain_name = aws_api_gateway_domain_name.token_vending_machine.domain_name
#  stage_name  = aws_api_gateway_deployment.token_vending_machine.stage_name
#  rest_api_id = aws_api_gateway_rest_api.token_vending_machine.id
#  base_path   = "token_vending_machine"
#}
#
#resource "aws_acm_certificate" "token_vending_machine" {
#  domain_name       = "token_vending_machine.com"
#  validation_method = "DNS"
#}
#
#resource "aws_cloudfront_distribution" "token_vending_machine" {
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
#    origin_path = "/${aws_api_gateway_rest_api.token_vending_machine.id}"
#    custom_header {
#      name  = "Host"
#      value = "execute-api.${local.region}.amazonaws.com"
#    }
#  }
#
#  viewer_certificate {
#    acm_certificate_arn = aws_acm_certificate.token_vending_machine.arn
#    ssl_support_method  = "sni-only"
#  }
#}

