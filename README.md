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

`v1.25.0-alpha2` is available as a prebuilt AMI in the following
regions:

| Region         | AMI ID                |
| -------------- | --------------------- |
| eu-central-1	 | ami-057c6620738fa1c5e |
| ap-south-1	 | ami-07699146c8968c7ba |
| eu-north-1     | ami-0c6358704cfd5f1bf |
| eu-west-3	     | ami-0b593ebe5bbf4821e |
| eu-west-2	     | ami-0ded97d5b8fc142b4 |
| eu-west-1	     | ami-06fe2c6763905e945 |
| ap-northeast-3 | ami-0c95e36ef4ef3fc9c |
| ap-northeast-2 | ami-0f3a6936713590534 |
| ap-northeast-1 | ami-0f66205bd15df4716 |
| ca-central-1	 | ami-0e7c38fd620186da5 |
| sa-east-1	     | ami-0d4af857fe6fd75fe |
| ap-southeast-1 | ami-06ffe8484cc5d1d70 |
| ap-southeast-2 | ami-09f3a80d02408a2d4 |
| us-east-1	     | ami-084acfb5fc5a6b4e0 |
| us-east-2	     | ami-04748fae9727a71a4 |
| us-west-1	     | ami-0ff36342dd5e120c3 |
| us-west-2	     | ami-043b9a1854e6d6af7 |
