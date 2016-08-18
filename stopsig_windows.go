// +build windows

package cmd

import "os"

var stopSignals = []os.Signal{os.Interrupt}
