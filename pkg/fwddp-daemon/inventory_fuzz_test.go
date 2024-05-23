package daemon

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/go-logr/logr"
	ethernetv1 "github.com/intel-collab/applications.orchestration.operators.intel-ethernet-operator/apis/ethernet/v1"
)

func FuzzAddNetInfo(f *testing.F) {

	d := ethernetv1.Device{
		PCIAddress: "0000:00:00.1",
		Name:       "dummy",
		VendorID:   "0001",
		DeviceID:   "123",
	}

	b, err := json.Marshal(d)

	if err != nil {
		fmt.Printf("Error: %v", err)
	}

	f.Add(b)
	f.Fuzz(func(t *testing.T, p1 []byte) {

		var log logr.Logger
		var dev ethernetv1.Device
		err = json.Unmarshal(p1, &dev)

		if err != nil {
			fmt.Printf("Error: %v", err)
		}

		addNetInfo(log, &dev)
	})
}
