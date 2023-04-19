package main

import (
	"fmt"
	"io"
	"os"
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

// Assuming this is running in the kubelet's bootstrap run, it is
// acceptable to simply write the desired hostname to the host's
// /etc/hostname file. Init will pick this ip and auto set the hostname
// before restarting the kubelet.
func setHostnameFile(imds *ImdsSession) error {
	// In AWS, the hostname should not be configurable. This is because we
	// use the EC2 role to authenticate the node with the kubernetes-api.
	// The aws-iam-authenticator setup will normally force nodes to auth
	// as their private DNS hostname.
	hostname, err := imds.GetMetadata("meta-data/hostname")
	if err != nil {
		return fmt.Errorf("Could not determine the AWS-provided hostname: %s", err)
	}

	if err := os.WriteFile("/host/etc/hostname", hostname, 0644); err != nil {
		return fmt.Errorf("Could not write hostname file: %s", err)
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

	if err = setHostnameFile(imds); err != nil {
		fmt.Printf("%s", err)
		os.Exit(1)
	}

	if err = saveKubeConfig(config, imds); err != nil {
		fmt.Printf("Could not save kubeconfig: %s", err)
		os.Exit(1)
	}

	if err = saveCredentialProviderConfig(); err != nil {
		fmt.Printf("Could not save credential provider configuration: %s", err)
		os.Exit(1)
	}

	// During the bootstrap run, the kubelet may attempt to generate its
	// own serving certificate. This is useless as a) it is self signed
	// and b) it is likely to be using the wrong IP address, or missing
	// the host name.
	// We have serverTLSBootstrap turned on, so kubelet will request its
	// certificate from the api-server - we just need to ensure that there
	// isn't an unexpired certificate present so that this process is
	// triggered.
	os.Remove("/host/var/lib/kubelet/pki/kubelet.crt")

	// Saving the config file should always be the last operation, as its
	// appearance is what triggers init to restart the kubelet.
	if err = saveKubeletConfiguration(config, imds); err != nil {
		fmt.Printf("Could not save kubelet configuration", err)
		os.Exit(1)
	}
}
