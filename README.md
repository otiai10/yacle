# yacle

[![Build Status](https://travis-ci.org/otiai10/yacle.svg?branch=master)](https://travis-ci.org/otiai10/yacle)

Yet Another [CWL](https://github.com/common-workflow-language/common-workflow-language) Engine

# installation

```sh
go get -u github.com/otiai10/yacle
```

# try it

```sh
yacle run 1st-tool.cwl echo-job.yml
```

# for cwl conformance test

## Just execute

```sh
cd cwl
./run_test.sh RUNNER=yacle -n1
```

## Install newest yacle and test with it

move to yacle directory and

```sh
go install . ;  cwl/run_test.sh RUNNER=yacle -n1
```
