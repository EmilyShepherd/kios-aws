package main

import (
	"encoding/base64"
	"fmt"
	"time"

	"github.com/EmilyShepherd/kios-go-sdk/pkg/bootstrap"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubeconfig "k8s.io/client-go/tools/clientcmd/api/v1"
	kubelet "k8s.io/kubelet/config/v1beta1"
)

// This providers information about the desired state of the node to the
// bootstrap SDK
type Provider struct {
	config *MetadataInformation
	imds   *ImdsSession
}

func (p *Provider) Init() error {
	imds, err := NewImdsSession(30)
	if err != nil {
		fmt.Errorf("Could not create IMDS Session: %s\n", err)
	}
	p.imds = imds

	config, err := imds.GetUserData()
	if err != nil {
		return fmt.Errorf("Could not load User Data: %s\n", err)
	}
	p.config = config

	return nil
}

// Returns the base64 decoded cluster CA certificate from the user data
func (p *Provider) GetClusterCA() bootstrap.Cert {
	clusterCA, err := base64.StdEncoding.DecodeString(p.config.ApiServer.CA)
	if err != nil {
		// fmt.Errorf("Could not decode CA certificate: %s", err)
	}

	return bootstrap.Cert{
		Cert: clusterCA,
	}
}

// Returns the default settings for the ECR Credential Provider
func (p *Provider) GetCredentialProviders() []kubelet.CredentialProvider {
	return []kubelet.CredentialProvider{kubelet.CredentialProvider{
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
		Args:       []string{"get-credentials"},
	}}
}

func (p *Provider) GetHostname() string {
	// In AWS, the hostname should not be configurable. This is because we
	// use the EC2 role to authenticate the node with the kubernetes-api.
	// The aws-iam-authenticator setup will normally force nodes to auth
	// as their private DNS hostname.
	hostname, err := p.imds.GetString("meta-data/hostname")
	if err != nil {
		//return fmt.Errorf("Could not determine the AWS-provided hostname: %s", err)
	}

	return hostname
}

func (p *Provider) GetNodeLabels() map[string]string {
	labels := p.config.Node.Labels

	if len(labels) == 0 {
		labels = make(map[string]string)
	}

	instanceType, _ := p.imds.GetString("meta-data/instance-type")
	zone, _ := p.imds.GetString("meta-data/placement/availability-zone")
	region, _ := p.imds.GetString("meta-data/placement/region")

	labels[v1.LabelInstanceTypeStable] = instanceType
	labels[v1.LabelTopologyZone] = zone
	labels[v1.LabelTopologyRegion] = region

	return labels
}

func (p *Provider) GetClusterEndpoint() string {
	return p.config.ApiServer.Endpoint
}

func (p *Provider) GetClusterAuthInfo() kubeconfig.AuthInfo {
	region, _ := p.imds.GetString("meta-data/placement/region")
	return kubeconfig.AuthInfo{
		Exec: &kubeconfig.ExecConfig{
			Command:    "/usr/libexec/kubernetes/kubelet-plugins/credential-provider/exec/aws-iam-authenticator",
			APIVersion: "client.authentication.k8s.io/v1beta1",
			Args: []string{
				"token",
				"-i",
				p.config.ApiServer.Name,
				"--region",
				region,
			},
		},
	}
}

func (p *Provider) GetContainerRuntimeConfiguration() bootstrap.ContainerRuntimeConfiguration {
	return p.config.Node.ContainerRuntime
}
