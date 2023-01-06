locals {
  service_name = replace(local.deployment_name, "-", "_")

  lambdas = {
    token-vending-machine = {
      description = "Token vending machine for ${local.deployment_name}"
      role        = aws_iam_role.token_vending_machine
      environment = {
        TOKEN_ACCESS_KEY_ID     = aws_iam_access_key.token_vending_machine_user_access_key.id
        TOKEN_SECRET_ACCESS_KEY = aws_iam_access_key.token_vending_machine_user_access_key.secret
      }
    }

    log-upload = {
      description = "Log upload for ${local.deployment_name}"
      role        = aws_iam_role.lambda_exec
    }
  }

  source_dir = "${path.module}/../"

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
    test          = "foo"

    main = join("", [
      for file in fileset("${local.source_dir}/lambdas/${each.key}", "*.go") :
      filebase64("${local.source_dir}/lambdas/${each.key}/${file}")
    ])
  }

  provisioner "local-exec" {
    command = <<EOT
        mkdir -p ${local.source_dir}/dist/lambdas/bin
        mkdir -p ${local.source_dir}/dist/lambdas/archive
        GOOS=linux GOARCH=amd64 go build -ldflags '-s -w' -o ${local.source_dir}/dist/lambdas/bin/${each.key} ${local.source_dir}/lambdas/${each.key}/.
    EOT
  }

}

data "archive_file" "this" {
  depends_on = [null_resource.lambda_build]
  for_each   = local.lambdas

  type        = "zip"
  source_file = "${local.source_dir}/dist/lambdas/bin/${each.key}"
  output_path = "${local.source_dir}/dist/lambdas/archive/${each.key}.zip"
}


resource "aws_lambda_function" "this" {
  for_each = local.lambdas

  filename         = "${local.source_dir}/dist/lambdas/archive/${each.key}.zip"
  function_name    = "${each.key}_${local.service_name}"
  description      = each.value.description
  role             = each.value.role.arn
  handler          = each.key
  publish          = false
  source_code_hash = data.archive_file.this[each.key].output_base64sha256
  runtime          = "go1.x"
  timeout          = "10"
  // if environment is defined then set it
  environment {
    variables = try(each.value.environment, {
      test: "FOOBAR"
    })
  }
  // merge local.tags with the lambda key and service_name
  tags = merge(local.tags, {
    "lambda"  = each.key
    "service" = local.service_name
  })
}

locals {
  log_upload_lambda = aws_lambda_function.this["log-upload"]
  token_lambda      = aws_lambda_function.this["token-vending-machine"]
}

