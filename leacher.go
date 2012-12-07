package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"io/ioutil"
	"path"
	"path/filepath"
	"github.com/gar3thjon3s/leacher/watcher"
	"github.com/gar3thjon3s/leacher/nzb"
)

var (
	nntpHost       = flag.String("host", "", "nntp hostname")
	nntpPort       = flag.Int("port", 119, "nntp port")
	nntpUser       = flag.String("user", "", "nntp user")
	nntpPass       = flag.String("pass", "", "nntp password")
	maxConnections = flag.Int("maxcon", 10, "maximum number of concurrent connections")
	homeDir        = flag.String("home", path.Join(os.Getenv("HOME"), ".leacher-go"), "leacher home directory")
)

func Usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS]\n", filepath.Base(os.Args[0]))
	fmt.Fprintln(os.Stderr, "\nOptions:")
	flag.PrintDefaults()
}

func ensureDirs(dirs ...string) error {
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0777); err != nil {
			return err
		}
		log.Printf("ensuring %s exists", d)
	}
	return nil
}

func main() {
	flag.Usage = Usage
	flag.Parse()

	if *nntpHost == "" {
		fmt.Fprintln(os.Stderr, "host is required")
		flag.Usage()
		os.Exit(1)
	}

	queueDir := path.Join(*homeDir, "queue")
	completeDir := path.Join(*homeDir, "complete")
	tmpDir := path.Join(*homeDir, "tmp")
	if err := ensureDirs(queueDir, completeDir, tmpDir); err != nil {
		fmt.Printf("could not create dirs: %s\n", err.Error())
		os.Exit(1)
	}

	nzbChannel := watcher.StartWatching(queueDir)
	defer close(nzbChannel)

	for {
		nzbs := <-nzbChannel
		for _, f := range nzbs {
			log.Printf("got new nzb file: %s", f)
			bytes, err := ioutil.ReadFile(path.Join(queueDir, f))
			if err != nil {
				log.Printf("could not read nzb file '%s': %s", f, err.Error())
				continue
			}
			nzbFile := nzb.Nzb{}
			if err := nzb.Parse(bytes, &nzbFile); err != nil {
				log.Printf("could not parse nzb file '%s': %s", f, err.Error())
			}
			log.Printf("parsed: %v", nzbFile)
		}
	}

	// watch the queue directory, parse file and put nzb onto an nzb queue channel
	// loop reading from nzb queue, putting files of nzb onto a file channel
	// have n workers reading from file channel (n = no. of concurrent downloads)

	// how to signal workers to finish up and shutdown cleanly?

	os.Exit(0)
}
