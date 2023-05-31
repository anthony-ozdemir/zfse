# Development Guide

## Requirements for Contributions

To ensure consistency and maintainability of the `ZFSE` project, all contributions must adhere to the following
requirements:

* Maximum line-length is `120` characters. Readability is far more important that dogmatic line-length. Thus, feel free
  to use all available space.
* Run the `golangci-lint` tool to check your code formatting before submitting your changes. Submissions with linting
  errors will not be accepted, unless a valid reason is provided.
* For significant changes or updates, please open a topic at discussions first to receive approval from the maintainers
  before you invest considerable effort into your proposed changes.

## Setup

ZFSE utilizes [golangci-lint](https://golangci-lint.run/) as the linter solution. Refer to installation
instruction provided by them [here](https://golangci-lint.run/usage/install/).

Maximum line-lengths should be set at `120` characters.
This can be achieved in `JetBrains GoLand` IDE via:

* `Settings > Editor > Code Style > Go > Wrapping and Braces`
* `Hardwrap at`: `120`
* `Visual Guides`: `120`
* `Function call arguments`: `Wrap if long`
* `Composite literals`: `Wrap if long`
* `Function parameters`: `Wrap if long`
* `Function return parameters`: `Wrap if long`