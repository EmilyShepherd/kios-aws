
package main

import (
  "fmt"
  "os"
  "io"
  "encoding/base64"
  "time"

  "sigs.k8s.io/yaml"
  kubelet "k8s.io/kubelet/config/v1beta1"
  kubeconfig "k8s.io/client-go/tools/clientcmd/api/v1"
  metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const ClusterCADir = "/etc/kubernetes/pki"
const ClusterCAPath = ClusterCADir + "/ca.crt"

// The path where the kubelet expects its various configurs to exist.
// These are hard coded into kios' init, so cannot be changed here.
const KubeletKubeconfigPath = "/etc/kubernetes/kubelet.conf"
const KubeletConfigurationPath = "/var/lib/kubelet/config.yaml"
const CredentialProviderConfigPath = "/etc/kubernetes/credential-providers.yaml"


// Saves the given cluster CA to file after first base 64 decoding it
func saveClusterCA(ca string) error {
  if err := os.MkdirAll("/host" + ClusterCADir, 0755); err != nil {
    return fmt.Errorf("Could not create Cluster CA Directory: %s", err)
  }

  clusterCA, err := base64.StdEncoding.DecodeString(ca)
  if err != nil {
    return fmt.Errorf("Could not decode CA certificate: %s", err)
  }

  if err = os.WriteFile("/host" + ClusterCAPath, clusterCA, 0644); err != nil {
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
        CertificateAuthority: ClusterCAPath,
      },
    }},
    AuthInfos: []kubeconfig.NamedAuthInfo{kubeconfig.NamedAuthInfo{
      Name: "default",
      AuthInfo: kubeconfig.AuthInfo{
        Exec: &kubeconfig.ExecConfig{
          Command: "/usr/libexec/kubernetes/kubelet-plugins/credential-provider/exec/aws-iam-authenticator",
          APIVersion: "client.authentication.k8s.io/v1beta1",
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

  if err = os.WriteFile("/host" + KubeletKubeconfigPath, kubeConfig, 0644); err != nil {
    return fmt.Errorf("Could not write Kubeconfig to disk: %s", err)
  }

  return nil
}

// Reads the given template file from disk and unmarshals it as YAML
func yamlFromFile(filename string, obj interface{}) error {
  file, err := os.Open("/etc/templates/" + filename)
  if err != nil {
    return fmt.Errorf("Could not open template file %s: %s", filename, err)
  }

  data, err := io.ReadAll(file)
  if err != nil {
    return fmt.Errorf("Could not read file %s: %s", filename, err)
  }

  if err := yaml.Unmarshal(data, obj); err != nil {
    return fmt.Errorf("Could not parse YAML from file %s: %s", filename, err)
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
  if err := yamlFromFile("config.yaml", &kubeletConfig); err != nil {
    return err
  }

  kubeletConfig.ProviderID = "aws:///" + az + "/" + instanceId

  kubelet, _ := yaml.Marshal(&kubeletConfig)
  os.WriteFile("/host" + KubeletConfigurationPath, kubelet, 0644)

  return nil
}

// Creates the credential provider configuration file for image
// credentials
func saveCredentialProviderConfig() error {
  config := kubelet.CredentialProviderConfig{}

  if err := yamlFromFile("credential-providers.yaml", &config); err != nil {
    return err
  }

  config.Providers = append(config.Providers, kubelet.CredentialProvider{
    Name: "ecr-credential-provider",
    MatchImages: []string{
      "*.dkr.ecr.*.amazonaws.com",
      "*.dkr.ecr.*.amazonaws.cn",
      "*.dkr.ecr-fips.*.amazonaws.com",
      "*.dkr.ecr.us-iso-east-1.c2s.ic.gov",
      "*.dkr.ecr.us-isob-east-1.sc2s.sgov.gov",
    },
    DefaultCacheDuration: &metav1.Duration{
      Duration: 12 * time.Hour,
    },
    APIVersion: "credentialprovider.kubelet.k8s.io/v1alpha1",
    Args: []string{"get-credentials"},
  })

  providerConfig, _ := yaml.Marshal(&config)
  os.WriteFile("/host" + CredentialProviderConfigPath, providerConfig, 0644)

  return nil
}
