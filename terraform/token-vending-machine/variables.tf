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