# yacle

[![Build Status](https://travis-ci.org/otiai10/yacle.svg?branch=master)](https://travis-ci.org/otiai10/yacle)
[![](https://img.shields.io/badge/dynamic/json.svg?label=CWL%20Conformance&url=https%3A%2F%2Fraw.githubusercontent.com%2Fotiai10%2Fyacle%2Fmaster%2F.conformance.json&query=pass&colorB=95c31e&suffix=%20cases)](https://github.com/common-workflow-language/common-workflow-language)

Yet Another [CWL](https://github.com/common-workflow-language/common-workflow-language) Engine

# Installation

```sh
go get -u -v github.com/otiai10/yacle
```

# Try it

```sh
yacle run 1st-tool.cwl echo-job.yml
```

# for cwl conformance test

## Just execute

```sh
git submodule update --init
./cwl/run_test.sh RUNNER=yacle -n1
```

## Development

Update the RUNNER: yacle,

```sh
go install .
./cwl/run_test.sh RUNNER=yacle -n1
```

## Get confermance coverage

```sh
./testcases
```

## If you get something wrong

```
go get ./...
```

