package main

import (
	"github.com/EmilyShepherd/kios-aws/pkg/awsbootstrap"
	"github.com/EmilyShepherd/kios-go-sdk/pkg/bootstrap"
)

var Bootstrap = bootstrap.Bootstrap{
	Binaries: []string{"aws-iam-authenticator", "ecr-credential-provider"},
	Provider: &awsbootstrap.Provider{},
}

func main() {
	Bootstrap.Run()
}
