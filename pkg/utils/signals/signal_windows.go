// +build !linux

package signals

import (
	"os"
)

var shutdownSignals = []os.Signal{os.Interrupt}
