# tftag

[![Release](https://github.com/bschaatsbergen/tftag/actions/workflows/goreleaser.yaml/badge.svg)](https://github.com/bschaatsbergen/tftag/actions/workflows/goreleaser.yaml) ![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/bschaatsbergen/tftag) ![GitHub commits since latest release (by SemVer)](https://img.shields.io/github/commits-since/bschaatsbergen/tftag/latest) [![Go Reference](https://pkg.go.dev/badge/github.com/bschaatsbergen/tftag.svg)](https://pkg.go.dev/github.com/bschaatsbergen/tftag)

A DRY approach to tagging Terraform resources

## Brew

To install tftag using brew, simply do the below.

```sh
brew tap bschaatsbergen/tftag
brew install tftag
```

## Binaries

You can download the [latest binary](https://github.com/bschaatsbergen/tftag/releases/latest) for Linux, MacOS, and Windows.

## Examples

Using `tftag` is very simple.

### Create a `.tftag.hcl`

```hcl
tftag "all" {
  tags = {
    Pine = "Apple",
  }
}

tftag "developers" {
  tags = {
    Straw = "Berry",
  }
}
```

### Optionally apply a filter on a resource

```hcl
resource "aws_s3_bucket" "users" {
  bucket = "users-bucket"
}

resource "aws_sns_topic" "user_updates" {
  #tftag:developers
  name = "user-updates-topic"
}
```

### Run tftag

```console
$ tftag
INFO[0000] Tagged `aws_sns_topic.user_updates` in main.tf
```

### Your resources are updated

```hcl
resource "aws_s3_bucket" "users" {
  bucket = "users-bucket"
  tags = {
    Pine  = "Apple"
    Straw = "Berry"
  }
}

resource "aws_sns_topic" "user_updates" {
  #tftag:developers
  name = "user-updates-topic"
  tags = {
    Straw = "Berry"
  }
}
```

## Contributing

Contributions are highly appreciated and always welcome.
Have a look through existing [Issues](https://github.com/bschaatsbergen/tftag/issues) and [Pull Requests](https://github.com/bschaatsbergen/tftag/pulls) that you could help with.
