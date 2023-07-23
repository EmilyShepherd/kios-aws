package main

import (
	"fmt"
	"os"

	v1 "k8s.io/api/core/v1"
)

// This file is read during kiOS' boot and is used to set node labels
const NodeLabelsPath = "/etc/kubernetes/node-labels"

// Helper function to write a label to the node-labels file
func writeLabel(f *os.File, key string, value string) error {
	_, err := f.WriteString(fmt.Sprintf("%s: %s\n", key, value))
	return err
}

// Generates the node-labels file, which is used by kiOS to set the
// labels which kubelet should register itself with
func saveNodeLabels(config *MetadataInformation, imds *ImdsSession) error {
	f, err := os.OpenFile(NodeLabelsPath, os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		return err
	}

	instanceType, _ := imds.GetString("meta-data/instance-type")
	zone, _ := imds.GetString("meta-data/placement/availability-zone")
	region, _ := imds.GetString("meta-data/placement/region")

	// These labels are traditionally set by kubelet when the
	// --cloud-provider aws flag is passed. However kubelet on kiOS is
	// providerless, so it is our job to set these labels.
	writeLabel(f, v1.LabelInstanceTypeStable, instanceType)
	writeLabel(f, v1.LabelTopologyZone, zone)
	writeLabel(f, v1.LabelTopologyRegion, region)

	for key, value := range config.Node.Labels {
		writeLabel(f, key, value)
	}

	return f.Close()
}
