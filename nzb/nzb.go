package nzb

import (
	"encoding/xml"
	"log"
	"regexp"
)

type NzbSegment struct {
	Bytes     int    `xml:"bytes,attr"`
	Number    int    `xml:"number,attr"`
	MessageId string `xml:",innerxml"`
}

type NzbFile struct {
	Groups   []string `xml:"groups>group"`
	Poster   string   `xml:"poster,attr"`
	Date     string   `xml:"date,attr"`
	Subject  string   `xml:"subject,attr"`
	Filename string
	Segments []NzbSegment `xml:"segments>segment"`
}

type Nzb struct {
	Files []NzbFile `xml:"file"`
}

var (
	filenameRegexp    = regexp.MustCompile("\"(.*?)\"")
	isoEncodingRegexp = regexp.MustCompile(" *encoding=\".*?\"")
)

func Parse(data []byte, nzb *Nzb) error {
	// doesn't like anything but utf-8, just ignore
	data = isoEncodingRegexp.ReplaceAll(data, []byte(""))
	if err := xml.Unmarshal(data, nzb); err != nil {
		return err
	}
	for i := range nzb.Files {
		if matches := filenameRegexp.FindStringSubmatch(nzb.Files[i].Subject); matches != nil {
			log.Printf("found filename in subject '%s'", matches[1])
			nzb.Files[i].Filename = matches[1]
		}
	}
	return nil
}
