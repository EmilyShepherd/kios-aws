
package main

import (
  "fmt"
  "os"
  "io"
)

const BinaryDstDir = "/host/usr/libexec/kubernetes/kubelet-plugins/credential-provider/exec/"

func copyBinaries(binaries []string) error {
  if err := os.MkdirAll(BinaryDstDir, 0755); err != nil {
    return fmt.Errorf("Could not create binary directory: %s", err)
  }

  for _, bin := range binaries {
    src, err := os.Open("/bin/" + bin)
    if err != nil {
      return fmt.Errorf("Could not open binary: %s", err)
    }
    defer src.Close()

    dst, _ := os.Create(BinaryDstDir + bin)
    if err != nil {
      return fmt.Errorf("Could not create binary copy: %s", err)
    }
    defer dst.Close()

    if _, err := io.Copy(dst, src); err != nil {
      return fmt.Errorf("Could not copy binary: %s", err)
    }
    if err := dst.Chmod(0755); err != nil {
      return fmt.Errorf("Could not update permissions of binary: %s", err)
    }
  }

  return nil
}

func main() {
  copyBinaries([]string{"aws-iam-authenticator", "ecr-credential-provider"})

  imds, err := NewImdsSession(30)
  if err != nil {
    fmt.Printf("Could not create IMDS Session: %s\n", err)
    os.Exit(1)
  }

  config, err := imds.GetUserData()
  if err != nil {
    fmt.Printf("Could not load User Data: %s\n", err)
    os.Exit(1)
  }

  if err = saveClusterCA(config.ApiServer.CA); err != nil {
    fmt.Printf("Could not save cluster CA: %s\n", err)
    os.Exit(1)
  }

  if err = saveKubeConfig(config, imds); err != nil {
    fmt.Printf("Could not save kubeconfig: %s", err)
    os.Exit(1)
  }

  if err = saveKubeletConfiguration(config, imds); err != nil {
    fmt.Printf("Could not save kubelet configuration", err);
    os.Exit(1)
  }

  if err = saveCredentialProviderConfig(); err != nil {
    fmt.Printf("Could not save credential provider configuration: %s", err);
    os.Exit(1)
  }
}
