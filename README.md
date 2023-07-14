# kiOS AWS Build

## Running kiOS

kiOS on AWS connects to the EKS cluster API Server in the same way that
"vanilla" EKS nodes do: nodes should have an instance profile, with
their role trusted in the clusters `aws-auth` ConfigMap:

```yaml
data:
  mapRoles: |
  - groups:
    - system:nodes
    rolearn: "arn:aws:iam::YOUR-ACCOUNT-ID:role/NODE-ROLE-NAME"
    username: "system:node:{{EC2PrivateDNSName}}"
```

The Workers do not require the `AmazonEKSWorkerNodePolicy` policy,
although if you are planning on pulling AWS ECR Images, nodes should
have permission to pull them:

```
arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly
```

Unlike EKS nodes, the instance metadata user data is _not_ a script,
instead it is a YAML configuration file, in the following format:

```yaml
apiVersion: kios.redcoat.dev/v1alpha1
kind: MetadataInformation
apiServer:
  name: EKS-CLUSTER-NAME
  endpoint: EKS-CLUSTER-URL
  b64ClusterCA: BASE64-EKS-CLUSTER-CA-CERTIFICATE
```

## AMI IDs

`v1.25.0-alpha4` is available as a prebuilt AMI in the following
regions:

| Region         | AMI ID                |
| -------------- | --------------------- |
| ap-south-1     | ami-0264f4da7e1b01430 |
| eu-north-1     | ami-0d9f1142e241b9954 |
| eu-west-3      | ami-036cae17c95e46565 |
| eu-west-2      | ami-08d58dde54547beda |
| eu-west-1      | ami-010eca3b0e0ace5d4 |
| ap-northeast-3 | ami-0622566835ba6d9e1 |
| ap-northeast-2 | ami-0c123f8f3a0ab5521 |
| ap-northeast-1 | ami-09035a8cb6b101c1d |
| ca-central-1   | ami-0ce61282059ab4127 |
| sa-east-1      | ami-0a5b1a00f64771f78 |
| ap-southeast-1 | ami-072c4201e81866fa5 |
| ap-southeast-2 | ami-0e183b0333d592948 |
| eu-central-1   | ami-0a605557f381f1143 |
| us-east-1      | ami-0a25072a47ea8ef55 |
| us-east-2      | ami-001c7d4b312c28419 |
| us-west-1      | ami-07aa2878992730381 |
| us-west-2      | ami-0227cd3b28d55dae8 |
