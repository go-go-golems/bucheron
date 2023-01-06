resource "aws_api_gateway_rest_api" "api" {
  name = "api"
  tags = local.tags
}

resource "aws_api_gateway_account" "name" {
  cloudwatch_role_arn = "${aws_iam_role.cloudwatch.arn}"
}

resource "aws_iam_role" "cloudwatch" {
  name = "apigateway_cloudwatch_role"
  assume_role_policy = <<EOF
{
 "Version": "2012-10-17",
 "Statement": [
   {
     "Sid": "",
     "Effect": "Allow",
     "Principal": {
     "Service": "apigateway.amazonaws.com"
    },
    "Action": "sts:AssumeRole"
  }
 ]
}
  EOF
}

resource "aws_iam_policy_attachment" "api_gateway_logs" {
  name = "api_gateway_logs"
  roles = ["${aws_iam_role.cloudwatch.id}"]
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonAPIGatewayPushToCloudWatchLogs"
}

resource "aws_api_gateway_method_settings" "name" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  stage_name  = aws_api_gateway_deployment.v1.stage_name
  method_path = "*/*"

  settings {
    logging_level = "INFO"
  }
}

resource "aws_lambda_permission" "token_vending_machine" {
  statement_id  = "AllowAPIGatewayInvoke"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.this["token-vending-machine"].arn
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_api_gateway_rest_api.api.execution_arn}/*/*"
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

  type                    = "AWS_PROXY"
  integration_http_method = "POST"
  uri                     = aws_lambda_function.this["token-vending-machine"].invoke_arn
}
#
resource "aws_api_gateway_method_response" "token_response_200" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  resource_id = aws_api_gateway_resource.token.id
  http_method = aws_api_gateway_method.token.http_method
  status_code = "200"

  response_models = {
    "application/json" = "Empty"
  }
}

resource "aws_api_gateway_deployment" "v1" {
  depends_on = [
    aws_api_gateway_integration.token_vending_machine,
  ]

  rest_api_id = aws_api_gateway_rest_api.api.id
  stage_name  = "v1"
}

