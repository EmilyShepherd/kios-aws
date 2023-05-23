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

`v1.25.0-alpha3` is available as a prebuilt AMI in the following
regions:

| Region         | AMI ID                |
| -------------- | --------------------- |
| eu-central-1   | ami-0c95db368938356d9 |
| ap-south-1     | ami-0ae610b9a099ab326 |
| eu-north-1     | ami-086a2f07e3a613f23 |
| eu-west-3      | ami-0352fb306b7b1bdae |
| eu-west-2      | ami-0d9e90f925f6efcb8 |
| eu-west-1      | ami-07152e399f3c2ecf4 |
| ap-northeast-3 | ami-01d71e7639818e91d |
| ap-northeast-2 | ami-04c032f008a7cd895 |
| ap-northeast-1 | ami-0d9e800ec72a16877 |
| ca-central-1   | ami-0d0c758fc5c2d20ca |
| sa-east-1      | ami-039875f19eec7e550 |
| ap-southeast-1 | ami-098960be091fb2a2e |
| ap-southeast-2 | ami-0122b6b53b63f50d5 |
| us-east-1      | ami-0b02881901dcd7366 |
| us-east-2      | ami-0ddc8d3c49870d0a3 |
| us-west-1      | ami-0096d677a951fcbf3 |
| us-west-2      | ami-05cdecc9b08f7df28 |
