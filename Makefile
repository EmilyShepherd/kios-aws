
bin:
	mkdir bin

bin/aws-bootstrap: bin cmd/bootstrap/*.go
	CGO_ENABLED=0 go build -o $@ -ldflags="-w -s" ./cmd/bootstrap/

bin/ecr-credential-provider: bin
	$(MAKE) -C ecr-credential-provider ecr-credential-provider
	cp ecr-credential-provider/ecr-credential-provider $@

bin/aws-iam-authenticator: bin
	$(MAKE) -C kubelet-credential-provider bin
	cp kubelet-credential-provider/_output/bin/aws-iam-authenticator $@
