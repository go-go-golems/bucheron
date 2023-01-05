output "token_url" {
  value = "${aws_api_gateway_deployment.v1.invoke_url}/token"
}

output "upload_url" {
  value = "${aws_api_gateway_deployment.v1.invoke_url}/upload"
}
