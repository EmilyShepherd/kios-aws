
PROJECT_URL=https://github.com/EmilyShepherd/kiOS
VERSION=v1.25.0-alpha2
RELEASE_URL=$(PROJECT_URL)/releases/download/$(VERSION)
MODE=local

.PHONY: all
all: bin/aws-bootstrap bin/ecr-credential-provider bin/aws-iam-authenticator

bin:
	mkdir bin

bin/aws-bootstrap: bin cmd/bootstrap/*.go
	CGO_ENABLED=0 go build -trimpath -o $@ -ldflags="-w -s" ./cmd/bootstrap/

bin/ecr-credential-provider: bin
	$(MAKE) -C ecr-credential-provider ecr-credential-provider
	cp ecr-credential-provider/ecr-credential-provider $@

bin/aws-iam-authenticator: bin
	$(MAKE) -C kubelet-credential-provider bin GOTAGS="no_add no_server no_init no_verify"
	cp kubelet-credential-provider/_output/bin/aws-iam-authenticator $@

ifeq ($(MODE),local)
EFI=../core/.build/bootpart/EFI/Boot/Bootx64.efi

datapart/modules:
	cp -r ../core/.build/datapart/modules $@

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

efi_size=$(shell ls -s $(EFI) | cut -f1 -d' ')
boot_blocks=$(shell expr $(efi_size) + 70)
size.boot.img=$(shell expr $(boot_blocks) '*' 2)
.boot.img: $(EFI)
	mkfs.vfat -C $@ -f1 $(boot_blocks)
	mcopy -si $@ $(subst Boot/Bootx64.efi,,$(EFI)) ::

.data.img: $(OCI) datapart/modules
	fakeroot mkfs.ext4 -d datapart $@ 600M

size.start.img := 2048
size.end.img := 33
.start.img .end.img:
	dd if=/dev/zero of=$@ bs=512 count=$(size$(@))

kios.img: .start.img .boot.img .data.img .end.img
	cat $^ > $@
	{ \
		echo g; \
		\
		echo n; \
		echo 1; \
		echo $(size.start.img); \
		echo +$(shell expr $(size.boot.img) - 1); \
		echo t; \
		echo 1; \
		\
		echo n; \
		echo 2; \
		echo $(shell expr $(size.boot.img) + $(size.start.img)); \
		echo -0; \
		\
		echo w; \
	} | fdisk $@

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
