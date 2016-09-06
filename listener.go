package cmd

import (
	"errors"
	"net"
	"os"
)

type fileFunc interface {
	File() (f *os.File, err error)
}

func listenerFiles(listeners []net.Listener) ([]*os.File, error) {
	files := make([]*os.File, 0, len(listeners))
	for _, l := range listeners {
		fd, ok := l.(fileFunc)
		if !ok {
			return nil, errors.New("no File() method for " + l.Addr().String())
		}
		f, err := fd.File()
		if err != nil {
			return nil, err
		}
		files = append(files, f)
	}
	return files, nil
}
