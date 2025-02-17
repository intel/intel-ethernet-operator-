// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020-2024 Intel Corporation

package daemon

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-logr/logr"
	ethernetv1 "github.com/intel-collab/applications.orchestration.operators.intel-ethernet-operator/apis/ethernet/v1"
	"github.com/intel-collab/applications.orchestration.operators.intel-ethernet-operator/pkg/utils"
)

var findDdp = findDdpProfile
var createDdpPaths = utils.CreateFullDdpPaths

type ddpUpdater struct {
	log        logr.Logger
	httpClient *http.Client
}

func (d *ddpUpdater) handleDDPUpdate(pciAddr string, ddpPath string) (bool, error) {
	log := d.log.WithName("handleDDPUpdate")
	if ddpPath == "" {
		return false, nil
	}

	err := d.updateDDP(pciAddr, ddpPath)
	if err != nil {
		log.Error(err, "Failed to update DDP", "device", pciAddr)
		return false, err
	}
	return true, nil
}

// ddpProfilePath is the path to our extracted DDP profile
func (d *ddpUpdater) updateDDP(pciAddr, ddpProfilePath string) error {
	log := d.log.WithName("updateDDP")

	devId, err := execCmd([]string{"sh", "-c", "lspci -vs " + pciAddr +
		" | awk '/Device Serial/ {print $NF}' | sed s/-//g"}, log)
	if err != nil {
		return err
	}
	devId = strings.TrimSuffix(devId, "\n")
	if devId == "" {
		return fmt.Errorf("failed to extract devId")
	}

	// create both intel/ice/ddp and updates/intel/ice/ddp
	// DDP paths for compatibility with different drivers
	intelPath, updatesPath := d.getDdpUpdatePaths()

	for _, path := range []string{intelPath, updatesPath} {
		if err := os.MkdirAll(path, 0600); err != nil {
			return err
		}

		target := filepath.Join(path, "ice-"+devId+".pkg")
		log.V(4).Info("Copying", "source", ddpProfilePath, "target", target)

		if err := utils.CopyFile(ddpProfilePath, target); err != nil {
			return err
		}
	}

	return nil
}

func (d *ddpUpdater) prepareDDP(config ethernetv1.DeviceNodeConfig, discoveredDDPs []string) (string, error) {
	log := d.log.WithName("prepareDDP")

	// filesystem path to DDP profile always take priority over URL path
	packageToLoad := ""
	if config.DeviceConfig.DiscoveredDDPPath != "" {
		for _, ddpPath := range discoveredDDPs {
			if ddpPath == config.DeviceConfig.DiscoveredDDPPath {
				log.V(4).Info("Provided path matched to discovered path")
				packageToLoad = ddpPath
				break
			}
		}
		if packageToLoad == "" {
			if config.DeviceConfig.DDPURL == "" {
				return "", fmt.Errorf("provided DDP path: '%s' not found", config.DeviceConfig.DiscoveredDDPPath)
			}
			log.V(4).Info("Provided path not found, fall back to ddpURL")
		}
	}

	if config.DeviceConfig.DDPURL == "" && packageToLoad == "" {
		log.V(4).Info("Empty DDPURL")
		return "", nil
	}

	// start preparation process
	targetPath := filepath.Join(artifactsFolder, config.PCIAddress)
	err := utils.CreateFolder(targetPath, log)
	if err != nil {
		return "", err
	}

	// extract from XZ archive located on host
	if packageToLoad != "" {
		fullPath := filepath.Join(targetPath, strings.TrimSuffix(filepath.Base(packageToLoad), ".xz"))
		log.V(4).Info("Extracting DDP from archive", "archive", packageToLoad)
		if err := unpackDDPXZArchive(packageToLoad, fullPath, log); err != nil {
			return "", err
		}
		return findDdp(targetPath)
	}

	// download from provided URL
	fullPath := filepath.Join(targetPath, filepath.Base(config.DeviceConfig.DDPURL))
	log.V(4).Info("Downloading", "url", config.DeviceConfig.DDPURL, "dstPath", fullPath)
	err = downloadFile(fullPath, config.DeviceConfig.DDPURL, config.DeviceConfig.DDPChecksum, d.httpClient)
	if err != nil {
		return "", err
	}

	log.V(4).Info("DDP file downloaded - extracting")
	// XXX so this unpacks into the same directory as the source file
	// We might add more comments here explaining the mechanics and reasoning
	if err := unpackDDPZipArchive(fullPath, targetPath, log); err != nil {
		return "", err
	}

	return findDdp(targetPath)
}

func (d *ddpUpdater) getDdpUpdatePaths() (string, string) {
	log := d.log.WithName("getDDPUpdatePath")

	const baseDir = "/lib/firmware/"
	intelPath, updatesPath := createDdpPaths(baseDir)
	log.V(4).Info("Using DDP paths", "path", intelPath, "path", updatesPath)

	return intelPath, updatesPath
}

func findDdpProfile(targetPath string) (string, error) {
	var ddpProfilesPaths []string
	walkFunction := func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(info.Name(), ".pkg") && info.Mode()&os.ModeSymlink == 0 {
			ddpProfilesPaths = append(ddpProfilesPaths, path)
		}
		return nil
	}
	err := filepath.Walk(targetPath, walkFunction)
	if err != nil {
		return "", err
	}
	if len(ddpProfilesPaths) != 1 {
		return "", fmt.Errorf("expected to find exactly 1 file ending with '.pkg', but found %v - %v", len(ddpProfilesPaths), ddpProfilesPaths)
	}
	return ddpProfilesPaths[0], err
}
