resource "aws_iam_role" "lambda_exec" {
  name = "lambda_exec"
  tags = local.tags

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF

}

# Attach role to Managed Policy
variable "iam_policy_arn" {
  description = "IAM Policy to be attached to role"
  type        = list(string)

  default = [
    "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
  ]
}

resource "aws_iam_policy_attachment" "role_attach" {
  name       = "policy-lambda-exec"
  roles      = [aws_iam_role.lambda_exec.id]
  count      = length(var.iam_policy_arn)
  policy_arn = element(var.iam_policy_arn, count.index)
}

resource "aws_iam_policy" "api_gateway_execution_policy" {
  name        = "api_gateway_execution_policy"
  path        = "/"
  description = "Policy for API Gateway execution role"

  tags = local.tags

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "logs:CreateLogGroup",
        "logs:CreateLogStream",
        "logs:PutLogEvents"
      ],
      "Resource": "arn:aws:logs:*:*:*"
    },
    {
      "Effect": "Allow",
      "Action": [
        "lambda:InvokeFunction"
      ],
      "Resource": "*"
    }
  ]
}
EOF
}

resource "aws_iam_role" "api_gateway_execution_role" {
  name = "api_gateway_execution_role"
  tags = local.tags

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
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

resource "aws_iam_role_policy_attachment" "api_gateway_execution_policy_attachment" {
  role       = aws_iam_role.api_gateway_execution_role.name
  policy_arn = aws_iam_policy.api_gateway_execution_policy.arn
}


resource "aws_iam_role" "token_vending_machine" {
  name               = "token_vending_machine"
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_policy" "sts_policy" {
  name        = "sts_policy"
  description = "Policy to grant STS permissions to the Lambda function"
  policy      = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": ["s3:PutObject"],
      "Effect": "Allow",
      "Resource": "${aws_s3_bucket.logs.arn}/*"
    },
    {
      "Action": [
        "sts:GetCallerIdentity",
        "sts:GetSessionToken"
      ],
      "Effect": "Allow",
      "Resource": "*"
    }
  ]
}
EOF
}

resource "aws_iam_policy_attachment" "attach_sts_policy" {
  name       = "attach_sts_policy"
  policy_arn = aws_iam_policy.sts_policy.arn
  roles      = [aws_iam_role.token_vending_machine.name]
}


resource "aws_iam_policy_attachment" "attach_basic_execution_role" {
  name       = "attach_basic_execution_role"
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
  roles      = [aws_iam_role.token_vending_machine.name, aws_iam_role.lambda_exec.name]
}

// create IAM user to use for getsessiontoken
resource "aws_iam_user" "token_vending_machine_user" {
  name = "token_vending_machine_user"
}

// attach the sts_policy to the user
resource "aws_iam_user_policy_attachment" "token_vending_machine_user_policy" {
  user       = aws_iam_user.token_vending_machine_user.name
  policy_arn = aws_iam_policy.sts_policy.arn
}

// create access keys for the user
resource "aws_iam_access_key" "token_vending_machine_user_access_key" {
  user = aws_iam_user.token_vending_machine_user.name
}

