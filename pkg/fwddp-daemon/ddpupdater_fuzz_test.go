package daemon

import (
	"net/http"
	"os/exec"
	"regexp"
	"strings"
	"testing"

	"github.com/go-logr/logr"
	ethernetv1 "github.com/intel-collab/applications.orchestration.operators.intel-ethernet-operator/apis/ethernet/v1"
	dh "github.com/intel-collab/applications.orchestration.operators.intel-ethernet-operator/pkg/drainhelper"
	"k8s.io/apimachinery/pkg/types"
)

func FuzzPrepareDDP(f *testing.F) {

	artifactsFolder = "./testdir/intel-ethernet-operator/"
	getInventory = func(_ logr.Logger) ([]ethernetv1.Device, error) {
		return data.Inventory, nil
	}
	downloadFile = func(path, url, checksum string, client *http.Client) error {
		return nil
	}
	untarFile = func(srcPath string, dstPath string, log logr.Logger) error {
		return nil
	}
	unpackDDPZipArchive = func(srcPath string, dstPath string, log logr.Logger) error {
		return nil
	}
	nvmupdateExec = func(cmd *exec.Cmd, log logr.Logger) error {
		return nil
	}
	getDDPsFromHost = func(_ logr.Logger) []string {
		return []string{discoveredDDPPath}
	}
	createDdpPaths = func(_ string) (string, string) {
		return ddpDir, ddpDir
	}

	findDdp = func(targetPath string) (string, error) {
		return "/tmp/fuzz-file", nil
	}

	pci := "0000:00:00.1"

	f.Add(pci)

	f.Fuzz(func(t *testing.T, p1 string) {
		match, err := regexp.MatchString(`^[a-fA-F0-9]{4}:[a-fA-F0-9]{2}:[01][a-fA-F0-9]\.[0-7]$`, p1)

		if err != nil {
			return
		}
		if match {
			devNodeConfig := ethernetv1.DeviceNodeConfig{
				PCIAddress: p1,
				DeviceConfig: ethernetv1.DeviceConfig{
					DDPURL: "http://testfwurl.zip",
				},
			}

			reconciler := &NodeConfigReconciler{
				Client:      nil,
				log:         logr.Discard(),
				drainHelper: &dh.DrainHelper{},
				nodeNameRef: types.NamespacedName{},
				ddpUpdater: &ddpUpdater{
					log:        logr.Discard(),
					httpClient: http.DefaultClient,
				},
				fwUpdater: nil,
			}

			_, err := reconciler.ddpUpdater.prepareDDP(devNodeConfig, []string{"http://testfwurl.zip", "http://testfwurl-1.zip"})
			if err != nil {
				if devNodeConfig.PCIAddress != "" {
					if strings.Contains(err.Error(), "invalid argument") {
						return
					} else if strings.Contains(err.Error(), "file name too long") {
						return
					} else {
						t.Errorf("\n\nError\n-----\nInput: %+v\nMsg: %+v\n", devNodeConfig, err)
					}
				} else {
					t.Errorf("\n\nError\n-----\nInput: %+v\nMsg: %+v\n", devNodeConfig, err)
				}
			}
		}
	})
}
