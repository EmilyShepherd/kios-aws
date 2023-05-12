#!/bin/bash
#
# Generate EKS Images is responsible for loading the list of images that
# should be preprimed onto images. NB: These are _not_ static pods, but
# are instead images for containers that are highly likely to be running
# on the node:
#   - Default VPC CNI Addon images for the k8s version
#   - Default Kube Proxy Addon images for the k8s version
#
# NB: The logic here was derived from the AWS EKS image build script:
# https://github.com/awslabs/amazon-eks-ami/blob/master/scripts/install-worker.sh

AWS_REGION=eu-central-1
K8S_VERSION=${K8S_VERSION:-1.25}
ECR_URL=602401143452.dkr.ecr.${AWS_REGION}.amazonaws.com

# Gets the version marked as default for this version of k8s
get_default_version() {
  aws eks describe-addon-versions \
    --kubernetes-version=${K8S_VERSION} \
    --addon-name $1 | jq -r "
      .addons[].addonVersions[] |
      select(.compatibilities[].defaultVersion == true) |
      .addonVersion
  "
}

IMAGES=()

VPC_VERSION=$(get_default_version vpc-cni)
IMAGES+=(
  ${ECR_URL}/amazon-k8s-cni:${VPC_VERSION}
  ${ECR_URL}/amazon-k8s-cni-init:${VPC_VERSION}
)

PROXY_VERSION=$(get_default_version kube-proxy)
PROXY_MINIMAL_VERSION=$(echo -n ${PROXY_VERSION} | sed 's/-/-minimal-/')
IMAGES+=(
  ${ECR_URL}/eks/kube-proxy:${PROXY_VERSION}
  ${ECR_URL}/eks/kube-proxy:${PROXY_MINIMAL_VERSION}
)

regions=$(aws ec2 describe-regions --output text | cut -f4)

for img in "${IMAGES[@]}"
do
  printf "$img"
  for region in ${regions}
  do
    printf "\t$(echo -n $img | sed "s/${AWS_REGION}/${region}/")"
  done
  printf "\n"
done > extra_images

