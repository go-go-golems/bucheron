#############################################
## API GATEWAY - Sets up & configure api gw
#############################################
#
#resource "aws_api_gateway_rest_api" "weather_gw" {
#  name        = "weather-api"
#  description = "created by terraform"
#}
#
#resource "aws_api_gateway_resource" "proxy" {
#  rest_api_id = "${aws_api_gateway_rest_api.weather_gw.id}"
#  parent_id   = "${aws_api_gateway_rest_api.weather_gw.root_resource_id}"
#  path_part   = "hello"
#}
#
#resource "aws_api_gateway_method" "options_method" {
#  rest_api_id   = "${aws_api_gateway_rest_api.weather_gw.id}"
#  resource_id   = "${aws_api_gateway_resource.proxy.id}"
#  http_method   = "OPTIONS"
#  authorization = "NONE"
#}
#
#resource "aws_api_gateway_method_response" "options_200" {
#  rest_api_id = "${aws_api_gateway_rest_api.weather_gw.id}"
#  resource_id = "${aws_api_gateway_resource.proxy.id}"
#  http_method = "${aws_api_gateway_method.options_method.http_method}"
#  status_code = "200"
#
#  response_models {
#    "application/json" = "Empty"
#  }
#
#  response_parameters {
#    "method.response.header.Access-Control-Allow-Headers" = true
#    "method.response.header.Access-Control-Allow-Methods" = true
#    "method.response.header.Access-Control-Allow-Origin"  = true
#  }
#
#  depends_on = ["aws_api_gateway_method.options_method"]
#}
#
#resource "aws_api_gateway_integration" "options_integration" {
#  rest_api_id = "${aws_api_gateway_rest_api.weather_gw.id}"
#  resource_id = "${aws_api_gateway_resource.proxy.id}"
#  http_method = "${aws_api_gateway_method.options_method.http_method}"
#  type        = "MOCK"
#
#  request_templates {
#    "application/json" = "{ \"statusCode\": 200 }"
#  }
#
#  depends_on = ["aws_api_gateway_method.options_method"]
#}
#
#resource "aws_api_gateway_integration_response" "options_integration_response" {
#  rest_api_id = "${aws_api_gateway_rest_api.weather_gw.id}"
#  resource_id = "${aws_api_gateway_resource.proxy.id}"
#  http_method = "${aws_api_gateway_method.options_method.http_method}"
#  status_code = "${aws_api_gateway_method_response.options_200.status_code}"
#
#  response_parameters = {
#    "method.response.header.Access-Control-Allow-Headers" = "'Content-Type,X-Amz-Date,Authorization,X-Api-Key,X-Amz-Security-Token'"
#    "method.response.header.Access-Control-Allow-Methods" = "'DELETE,GET,HEAD,OPTIONS,PATCH,POST,PUT'"
#    "method.response.header.Access-Control-Allow-Origin"  = "'*'"
#  }
#
#  depends_on = ["aws_api_gateway_method_response.options_200"]
#}
#
#resource "aws_api_gateway_method" "proxy" {
#  rest_api_id   = "${aws_api_gateway_rest_api.weather_gw.id}"
#  resource_id   = "${aws_api_gateway_resource.proxy.id}"
#  http_method   = "ANY"
#  authorization = "NONE"
#}
#
#resource "aws_api_gateway_method_response" "200" {
#  rest_api_id = "${aws_api_gateway_rest_api.weather_gw.id}"
#  resource_id = "${aws_api_gateway_resource.proxy.id}"
#  http_method = "${aws_api_gateway_method.proxy.http_method}"
#  status_code = "200"
#
#  response_models = {
#    "application/json" = "Empty"
#  }
#
#  response_parameters = {
#    "method.response.header.Access-Control-Allow-Origin" = true
#  }
#
#  depends_on = ["aws_api_gateway_method.proxy"]
#}
#
#resource "aws_api_gateway_integration" "lambda" {
#  rest_api_id = "${aws_api_gateway_rest_api.weather_gw.id}"
#  resource_id = "${aws_api_gateway_method.proxy.resource_id}"
#  http_method = "${aws_api_gateway_method.proxy.http_method}"
#
#  integration_http_method = "POST"
#  type                    = "AWS_PROXY"
#  uri                     = "${aws_lambda_function.weather_api.invoke_arn}"
#  depends_on              = ["aws_api_gateway_method.proxy", "aws_lambda_function.weather_api"]
#}
#
#resource "aws_api_gateway_deployment" "gw_deploy" {
#  depends_on = [
#    "aws_api_gateway_integration.lambda",
#  ]
#
#  rest_api_id = "${aws_api_gateway_rest_api.weather_gw.id}"
#  stage_name  = "stage"
#}