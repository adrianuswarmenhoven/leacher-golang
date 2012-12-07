package watcher

import (
	"io/ioutil"
	"log"
	"regexp"
	"time"
)

var nzbRegexp = regexp.MustCompile(".*\\.nzb$")

func filterNzbs(files []string) []string {
	result := []string{}
	for _, f := range files {
		if match := nzbRegexp.FindStringIndex(f); match != nil {
			result = append(result, f)
		}
	}
	return result
}

// how come golang doesn't have sets? :|
func difference(previous []string, current []string) []string {
	result := []string{}
	for _, f := range current {
		newFile := true
		for _, f2 := range previous {
			if f == f2 {
				newFile = false
				break
			}
		}
		if newFile {
			result = append(result, f)
		}
	}
	return result
}

func StartWatching(dirname string) chan []string {
	nzbChannel := make(chan []string)
	go func() {
		previousFiles := []string{}
		for {
			fileInfos, err := ioutil.ReadDir(dirname)
			if err != nil {
				log.Printf("could not read directory for nzbs: %s", dirname)
			}
			currentFiles := make([]string, len(fileInfos))
			for i, f := range fileInfos {
				currentFiles[i] = f.Name()
			}
			newFiles := difference(previousFiles, filterNzbs(currentFiles))
			if len(newFiles) > 0 {
				log.Printf("found %d new nzbs: %s\n", len(newFiles), newFiles)
				nzbChannel <- newFiles
			}
			previousFiles = currentFiles
			time.Sleep(1 * time.Second)
		}
	}()
	return nzbChannel
}
