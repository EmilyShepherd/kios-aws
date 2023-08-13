package main

import (
	"github.com/EmilyShepherd/kios-go-sdk/pkg/bootstrap"
)

var Bootstrap = bootstrap.Bootstrap{
	Binaries: []string{"aws-iam-authenticator", "ecr-credential-provider"},
	Provider: &Provider{},
}

func main() {
	Bootstrap.Run()
}
