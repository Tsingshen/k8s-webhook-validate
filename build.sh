#!/bin/sh

go mod tidy
GOOS=linux go build