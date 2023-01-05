output "token_url" {
  value = "${aws_api_gateway_deployment.v1.invoke_url}/token"
}

output "upload_url" {
  value = "${aws_api_gateway_deployment.v1.invoke_url}/upload"
}

output "token_access_key_id" {
    value = "${aws_iam_access_key.token_vending_machine_user_access_key.id}"
}

output "token_secret_access_key" {
    value = "${aws_iam_access_key.token_vending_machine_user_access_key.secret}"
  sensitive = true
}
