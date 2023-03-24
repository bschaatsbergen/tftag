resource "aws_s3_bucket" "users" {
  bucket = "users-bucket"
  tags = {
    Pine = "Apple"
  }
}
