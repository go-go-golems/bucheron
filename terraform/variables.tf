variable "app_name" {
    default = "bucheron"
}

variable "environment" {
  default = "dev"
}

variable "tags" {
  type    = map(string)
  default = {}
}

variable "s3_bucket_name" {
  default = "wesen-ppa-control-logs"
}