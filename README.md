# Useful Go HTTP middlewares [![Documentation](https://godoc.org/github.com/vincentserpoul/gohttpmw?status.svg)](http://godoc.org/github.com/<username>/<library>) [![Go Report Card](https://goreportcard.com/badge/github.com/vincentserpoul/gohttpmw)](https://goreportcard.com/report/github.com/vincentserpoul/gohttpmw) [![Coverage Status](https://coveralls.io/repos/github/vincentserpoul/gohttpmw/badge.svg?branch=master)](https://coveralls.io/github/vincentserpoul/gohttpmw?branch=master) [![CircleCI](https://circleci.com/gh/vincentserpoul/gohttpmw.svg?style=svg)](https://circleci.com/gh/vincentserpoul/gohttpmw) [![Maintainability](https://api.codeclimate.com/v1/badges/344c7922467ddf1066bf/maintainability)](https://codeclimate.com/github/vincentserpoul/gohttpmw/maintainability)

## Content type

Simply set the content-type in the header.

## RBAC

depends on

- github.com/ory/ladon

Allows to check if a user is allowed to access the URL, according to his role.
You need to specify the function that will get the role name from the context.

## Request ID

depends on

- github.com/segmentio/ksuid

Set a request id in the context as well as the header.
We use ksuid as it's a sortable uinque identifier that is only 20-bytes wide.

## Security

Set the basic headers necessary for a better security.
These are not sufficient, you will need to add more according to the context of your application.

## Logger

Add logs to the request

_If you use other middlewares, make sure they don't change the reference to the request_
