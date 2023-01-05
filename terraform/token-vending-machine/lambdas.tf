locals {
  service_name = replace(local.deployment_name, "-", "_")

  lambdas = {
    token-vending-machine = {
      description = "Token vending machine for ${local.deployment_name}"
    }

    log-upload = {
      description = "Log upload for ${local.deployment_name}"
    }
  }

  lambda_exists = {
    lambda_binary_exists = {
      for key, _ in local.lambdas : key => fileexists("${path.module}/lambdas/bin/${key}")
    }
  }
}

resource "null_resource" "lambda_build" {
  for_each = local.lambdas

  triggers = {
    binary_exists = local.lambda_exists.lambda_binary_exists[each.key]
    test = "foo"

    main = join("", [
      for file in fileset("${path.module}/../../lambdas/${each.key}", "*.go") :
      filebase64("${path.module}/../../lambdas/${each.key}/${file}")
    ])
  }

  provisioner "local-exec" {
    command = <<EOT
        mkdir -p ${path.module}/../../dist/lambdas/bin
        mkdir -p ${path.module}/../../dist/lambdas/archive
        GOOS=linux GOARCH=amd64 go build -ldflags '-s -w' -o ${path.module}/../../dist/lambdas/bin/${each.key} ${path.module}/../../lambdas/${each.key}/.
    EOT
  }

}

data "archive_file" "this" {
  depends_on = [null_resource.lambda_build]
  for_each   = local.lambdas

  type        = "zip"
  source_file = "${path.module}/../../dist/lambdas/bin/${each.key}"
  output_path = "${path.module}/../../dist/lambdas/archive/${each.key}.zip"
}


resource "aws_lambda_function" "this" {
  for_each = local.lambdas

  filename         = "${path.module}/../../dist/lambdas/archive/${each.key}.zip"
  function_name    = "${each.key}_${local.service_name}"
  description      = each.value.description
  role             = aws_iam_role.lambda_exec.arn
  handler          = each.key
  publish          = false
  source_code_hash = data.archive_file.this[each.key].output_base64sha256
  runtime          = "go1.x"
  timeout          = "10"
  // merge local.tags with the lambda key and service_name
  tags             = merge(local.tags, {
    "lambda"  = each.key
    "service" = local.service_name
  })
}