resource "aws_s3_bucket" "logs" {
  bucket = var.s3_bucket_name
}
