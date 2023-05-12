
.PHONY: all
all: bin/aws-bootstrap bin/ecr-credential-provider bin/aws-iam-authenticator

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


EFI=../core/.build/bootpart/EFI/Boot/Bootx64.efi
efi_size=$(shell ls -s $(EFI) | cut -f1 -d' ')
boot_blocks=$(shell expr $(efi_size) + 70)
size.boot.img=$(shell expr $(boot_blocks) '*' 2)
.boot.img: $(EFI)
	mkfs.vfat -C $@ -f1 $(boot_blocks)
	mcopy -si $@ ../core/.build/bootpart/EFI ::

.data.img:
	fakeroot mkfs.ext4 -d ../core/.build/datapart $@ 600M

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

S3_BUCKET := kios.redcoat.dev
S3_KEY := kios.img
import-snapshot: kios.img
	aws s3 cp $< s3://$(S3_BUCKET)/$(S3_KEY)
	aws ec2 import-snapshot --disk-container "UserBucket={S3Bucket=$(S3_BUCKET),S3Key=$(S3_KEY)}" | jq -r .ImportTaskId > $@

snapshot := $(shell aws ec2 describe-import-snapshot-tasks --import-task-ids $(shell cat import-snapshot) | jq -r '[.ImportSnapshotTasks[0].SnapshotTaskDetail | to_entries[].value | select(type == "string")] | @tsv')
AMI_NAME ?= kios

ifeq ($(word 3,$(snapshot)),completed)
register-image:
	aws ec2 register-image \
		--name $(AMI_NAME) \
		--architecture x86_64 \
		--boot-mode uefi \
		--ena-support \
		--root-device-name /dev/xvda \
		--block-device-mappings \
			DeviceName=/dev/xvda,Ebs={SnapshotId=$(word 2,$(snapshot))}
endif

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
