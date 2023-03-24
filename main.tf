resource "aws_s3_bucket" "users" {
  bucket = "users-bucket"

  taxgs = {
    a = base64(
      "cyx"
    )
    "${var.environment}" = "Apple"
    Google               = "Rules"
    Pine                 = "Apple"
    Money                = "Bad"
  }
  tags = {
    Google = "Rules"
    Pine   = "Apple"
    Money  = "Bad"
  }
}

resource "aws_s3_bucket" "test" {
  #tftag:finance
  bucket = "test-bucket"
  tags = {
    Money = "Bad"
  }
}