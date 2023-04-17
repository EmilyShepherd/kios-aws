
package main

import (
  "fmt"
  "os"
  "encoding/base64"

  "sigs.k8s.io/yaml"
  kubelet "k8s.io/kubelet/config/v1beta1"
  kubeconfig "k8s.io/client-go/tools/clientcmd/api/v1"
)

// Saves the given cluster CA to file after first base 64 decoding it
func saveClusterCA(ca string) error {
  clusterCA, err := base64.StdEncoding.DecodeString(ca)
  if err != nil {
    return fmt.Errorf("Could not decode CA certificate: %s", err)
  }

  if err = os.WriteFile("/tmp/ca.cert", clusterCA, 0644); err != nil {
    return fmt.Errorf("Could not write CA certificate to disk: %s", err)
  }

  return nil
}

// Generates a KubeConfig file for Kubelet, marshals it to YAML, and
// saves it
func saveKubeConfig(config *MetadataInformation, imds *ImdsSession) error {
  region, _ := imds.GetString("meta-data/placement/region")

  kubeConfig, err := yaml.Marshal(&kubeconfig.Config {
    Kind: "Config",
    APIVersion: "v1",
    Clusters: []kubeconfig.NamedCluster{kubeconfig.NamedCluster{
      Name: "default",
      Cluster: kubeconfig.Cluster{
        Server: config.ApiServer.Endpoint,
        CertificateAuthority: "/tmp/ca.cert",
      },
    }},
    AuthInfos: []kubeconfig.NamedAuthInfo{kubeconfig.NamedAuthInfo{
      Name: "default",
      AuthInfo: kubeconfig.AuthInfo{
        Exec: &kubeconfig.ExecConfig{
          Command: "/usr/libexec/kubernetes/kubelet-plugins/credential-provider/exec/aws-iam-authenticator",
          Args: []string{
            "token",
            "-i",
            config.ApiServer.Name,
            "--region",
            region,
          },
        },
      },
    }},
    Contexts: []kubeconfig.NamedContext{kubeconfig.NamedContext{
      Name: "default",
      Context: kubeconfig.Context{
        Cluster: "default",
        AuthInfo: "default",
      },
    }},
    CurrentContext: "default",
  })
  if err != nil {
    return fmt.Errorf("Could not marshal KubeConfig YAML: %s", err)
  }

  if err = os.WriteFile("/tmp/kubeconfig", kubeConfig, 0644); err != nil {
    return fmt.Errorf("Could not write Kubeconfig to disk: %s", err)
  }

  return nil
}

// Loads the template kubeconfig file from disk, adds the relavent
// settings to it, before remarshalling it as YAML and saving it back to
// disk
func saveKubeletConfiguration(config *MetadataInformation, imds *ImdsSession) error {
  az, _ := imds.GetString("meta-data/placement/availability-zone")
  instanceId, _ := imds.GetString("meta-data/instance-id")

  kubeletConfig := kubelet.KubeletConfiguration{}

  kubeletConfig.ProviderID = "aws:///" + az + "/" + instanceId

  kubelet, _ := yaml.Marshal(&kubeletConfig)
  os.WriteFile("/tmp/kubelet.yaml", kubelet, 0644)

  return nil
}
