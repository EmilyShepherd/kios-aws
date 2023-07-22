package main

import (
	"fmt"
	"os"
)

// This file is read during kiOS' boot and is used to set node labels
const CrioConfigurationPath = "/etc/crio/crio.conf"

var settingWritten = false

func writeTableHeading(f *os.File, name string) error {
	_, err := f.WriteString(fmt.Sprintf("[crio.%s]\n", name))
	return err
}

// Helper function to write a label to the node-labels file
func writeSetting(f *os.File, key string, value string) error {
	_, err := f.WriteString(fmt.Sprintf("%s = \"%s\"\n", key, value))
	settingWritten = true
	return err
}

// Generates the node-labels file, which is used by kiOS to set the
// labels which kubelet should register itself with
func saveCrioConfiguration(config ContainerRuntimeConfiguration) (bool, error) {
	f, err := os.OpenFile("/host"+CrioConfigurationPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return false, err
	}

	if config.ImageVolumes != "" {
		writeTableHeading(f, "image")
		writeSetting(f, "image_volumes", config.ImageVolumes)
	}

	return settingWritten, f.Close()
}
