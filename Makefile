
PROJECT_URL=https://github.com/EmilyShepherd/kiOS
VERSION=v1.25.0-alpha5
RELEASE_URL=$(PROJECT_URL)/releases/download/$(VERSION)
MODE=local

.PHONY: all
all: bin/aws-bootstrap bin/ecr-credential-provider bin/aws-iam-authenticator

bin:
	mkdir bin

bin/aws-bootstrap: bin pkg/*/*.go
	CGO_ENABLED=0 go build -trimpath -o $@ -ldflags="-w -s" .

bin/ecr-credential-provider: bin
	$(MAKE) -C ecr-credential-provider ecr-credential-provider
	cp ecr-credential-provider/ecr-credential-provider $@

bin/aws-iam-authenticator: bin
	$(MAKE) -C kubelet-credential-provider bin GOTAGS="no_add no_server no_init no_verify"
	cp kubelet-credential-provider/_output/bin/aws-iam-authenticator $@

ifeq ($(MODE),local)
EFI=../core/.build/bootpart/EFI/Boot/Bootx64.efi

datapart/modules:
	cp -r ../core/.build/root/lib/modules $@

else
EFI=bootpart/EFI/Boot/Bootx64.efi
$(EFI):
	mkdir -p $(dir $(EFI))
	curl -Lo $@ $(RELEASE_URL)/kios-x86_64.efi

datapart/modules:
	mkdir -p $@
	curl -L $(RELEASE_URL)/kios-modules-x86_64.tar.gz | gunzip -c | tar -xC $@
endif

extra_images:
	./hack/generate-eks-images.sh

OCI=datapart/data/oci/overlay-layers/layers.json
$(OCI): $(wildcard datapart/meta/etc/kubernetes/manifests/*.yaml) extra_images
	./hack/prime-containers.sh

kios.img: $(OCI)
	DATAPART_PATH=./datapart BOOTPART_PATH=../core/.build/bootpart ../core/scripts/create-image

register-image: kios.img
	./hack/import-image $< $(VERSION)

.PHONY: clean
clean:
	rm -rf .*.img kios.img bin import-snapshot

LOOP := /dev/loop0
mount: kios.img
	sudo losetup -P $(LOOP) $<
	mkdir -p .bootpart .datapart
	sudo mount $(LOOP)p1 .bootpart
	sudo mount $(LOOP)p2 .datapart

umount:
	sudo losetup -d $(LOOP)
	sudo umount .bootpart .datapart
	rmdir .bootpart .datapart
