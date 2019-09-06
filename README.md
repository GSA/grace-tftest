# GRACE Terraform Test [![License](https://img.shields.io/badge/license-CC0-blue)](LICENSE.md) [![GoDoc](https://img.shields.io/badge/go-documentation-blue.svg)](https://godoc.org/github.com/GSA/grace-tftest/aws) [![CircleCI](https://circleci.com/gh/GSA/grace-tftest.svg?style=shield)](https://circleci.com/gh/GSA/grace-tftest) [![Go Report Card](https://goreportcard.com/badge/github.com/GSA/grace-tftest)](https://goreportcard.com/report/github.com/GSA/grace-tftest)

## Repository Contents

This repository contains supplemental AWS functions that were required by the GRACE team and not available in [terratest](https://github.com/gruntwork-io/terratest). These functions can potentially be commited to terratest in pull requests in the future.

## Usage instructions
To enable debugging, set the `TFTEST_DEBUG` environment variable to `true`

1. Install system dependencies.
    1. [Go](https://golang.org/)
    1. [Dep](https://golang.github.io/dep/docs/installation.html)
    1. [GolangCI](https://github.com/golangci/golangci-lint)
    1. [gosec](https://github.com/securego/gosec)
1. [Configure AWS](https://www.terraform.io/docs/providers/aws/#authentication) with AWS credentials locally.




## Public domain

This project is in the worldwide [public domain](LICENSE.md). As stated in [CONTRIBUTING](CONTRIBUTING.md):

> This project is in the public domain within the United States, and copyright and related rights in the work worldwide are waived through the [CC0 1.0 Universal public domain dedication](https://creativecommons.org/publicdomain/zero/1.0/).
>
> All contributions to this project will be released under the CC0 dedication. By submitting a pull request, you are agreeing to comply with this waiver of copyright interest.