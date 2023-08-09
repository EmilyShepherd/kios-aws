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

`v1.25.0-alpha5` is available as a prebuilt AMI in the following
regions:

| Region         | AMI ID                |
| -------------- | --------------------- |
| ap-south-1     | ami-091385bc01c832f8a |
| eu-north-1     | ami-0b6ee538b14d1eb9f |
| eu-west-3      | ami-0761c49849b23e748 |
| eu-west-2      | ami-0904f04285477df47 |
| eu-west-1      | ami-04fe4d24565904862 |
| ap-northeast-3 | ami-02150d2484d0f6647 |
| ap-northeast-2 | ami-00ba994c5cb7ec4c8 |
| ap-northeast-1 | ami-0d976f199e5a9e537 |
| ca-central-1   | ami-0be8171cb5c479eb3 |
| sa-east-1      | ami-0c63ed6c1aabfc90f |
| ap-southeast-1 | ami-0af4c59c69ed2ceb1 |
| ap-southeast-2 | ami-069a6f91bb08699b2 |
| eu-central-1   | ami-000b96a1a3b9f67ef |
| us-east-1      | ami-0fc40df5277b96f3a |
| us-east-2      | ami-09a52f3edd7258d16 |
| us-west-1      | ami-0fa269838296e4745 |
| us-west-2      | ami-026aa572b82ded35d |

