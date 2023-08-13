package main

import (
	"fmt"
	"strings"

	"k8s.io/klog/v2"
	kubelet "k8s.io/kubelet/config/v1beta1"
	"sigs.k8s.io/yaml"
)

// Loads the template kubeconfig file from disk, adds the relavent
// settings to it, before remarshalling it as YAML and saving it back to
// disk
func (p *Provider) GetKubeletConfiguration(kubeletConfig kubelet.KubeletConfiguration) kubelet.KubeletConfiguration {
	if err := yaml.Unmarshal([]byte(p.config.Node.KubeletConfiguration), &kubeletConfig); err != nil {
		fmt.Printf("WARNING: Bad YAML in KubeletConfiguration. Ignoring. %s", err)
	}

	kubeletConfig.ServerTLSBootstrap = true
	kubeletConfig.RegisterWithTaints = p.config.Node.Taints

	// In the spirit on unopinionated-ness, we will accept it if a
	// ProviderID has been specified.
	// NB: If you are running with a EKS-provided cluster, the control
	// plane WILL instantly delete any nodes which do not have an
	// expected ProviderID, so override with caution!
	if kubeletConfig.ProviderID != "" {
		klog.Warning("ProviderID is manually set. Use with caution")
	} else {
		klog.Infof("ProviderID is not manually set. Creating EKS-expected providerID from metadata")

		az, _ := p.imds.GetString("meta-data/placement/availability-zone")
		instanceId, _ := p.imds.GetString("meta-data/instance-id")
		kubeletConfig.ProviderID = "aws:///" + az + "/" + instanceId
	}
	klog.Infof("Using ProviderID: %s", kubeletConfig.ProviderID)

	// If the defined KubeletConfiguration has already set the ClusterDNS
	// values, we won't make an attempt to use the EKS-default values.
	if len(kubeletConfig.ClusterDNS) != 0 {
		klog.Info("Cluster DNS is manually set")
	} else {
		klog.Info("Cluster DNS is not manually set, using EKS default")

		// EKS' default service CIDR is 10.100.0.0/16 _unless_ the VPC CIDR is
		// in the 10.0.0.0/8 - in this case, the service CIDR is 172.20.0.0/16.
		// By convention, the cluster dns service cluster IP is x.x.0.10
		ip, _ := p.imds.GetString("meta-data/local-ipv4")

		if strings.HasPrefix(ip, "10.") {
			kubeletConfig.ClusterDNS = []string{"172.20.0.10"}
		} else {
			kubeletConfig.ClusterDNS = []string{"10.100.0.10"}
		}
	}
	klog.Infof("Using Cluster DNS: %v", kubeletConfig.ClusterDNS)

	// If using AWS VPC CNI with prefix mode turned off, there is a limit
	// to the number of IP addresses (and therefore pods) each node can
	// have. The only way to represent this currently is by setting a pod
	// limit at the kubelet level.
	// If this is on, we'll look up the number of IP addresses that this
	// node can have (minus those not used by AWS VPC CNI). An offset can
	// be applied if we know pods with hostNetwork will be on the node as
	// these do not use up one of the IP addresse.
	// The default offset is 3 for:
	//   - The node pod
	//   - An assumed kube-proxy DaemonSet
	//   - An assumed aws-vpc-cni DaemonSet
	if p.config.Node.MaxPods.Set {
		instanceType, _ := p.imds.GetString("meta-data/instance-type")
		kubeletConfig.MaxPods = int32(PodLimits[instanceType] + p.config.Node.MaxPods.Offset)
	}

	return kubeletConfig
}
