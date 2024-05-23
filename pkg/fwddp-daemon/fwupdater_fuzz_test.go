package daemon

import (
	"fmt"
	"testing"

	"github.com/go-logr/logr"
	dh "github.com/intel-collab/applications.orchestration.operators.intel-ethernet-operator/pkg/drainhelper"
	"k8s.io/apimachinery/pkg/types"
)

func FuzzUpdateFirmware(f *testing.F) {

	f.Add("0000:00:00.1", "http://testfwurl", "-if ioctl")
	f.Fuzz(func(t *testing.T, p1, p2, p3 string) {

		reconciler := &NodeConfigReconciler{
			Client:      nil,
			log:         logr.Discard(),
			drainHelper: &dh.DrainHelper{},
			nodeNameRef: types.NamespacedName{},
			ddpUpdater:  nil,
			fwUpdater:   &fwUpdater{},
		}

		_, err := reconciler.fwUpdater.updateFirmware(p1, p2, p3)
		if err != nil {
			fmt.Printf("updateFirmware() failed : %s, %s, %s", p1, p2, p3)
		} else {
			return
		}
	})
}
