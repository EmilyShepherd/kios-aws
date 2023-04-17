
package main

import (
  metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const ApiVersion = "kios.redcoat.dev/v1alpha1"
const Kind = "AWSMetadataInformation"

type ApiServer struct {
  Name string `json:"name"`
  CA string `json:"b64ClusterCA"`
  Endpoint string `json:"endpoint"`
}

type Node struct {
  Taints []string `json:"taints"`
}

type MetadataInformation struct {
  metav1.TypeMeta `json:",inline"`

  ApiServer ApiServer `json:"apiServer"`
  Node Node `json:"node"`
}
