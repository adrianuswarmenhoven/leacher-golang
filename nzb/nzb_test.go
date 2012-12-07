package nzb

import (
	"fmt"
	"os"
	"testing"
)

func TestParse(t *testing.T) {
	input := `<?xml version="1.0" encoding="iso-8859" ?>
<!DOCTYPE nzb PUBLIC "-//newzBin//DTD NZB 1.0//EN" "http://www.newzbin.com/DTD/nzb/nzb-1.0.dtd">
<nzb xmlns="http://www.newzbin.com/DTD/2003/nzb">

<file poster="hihi &lt;hihi@kere.ws&gt;" date="1344347536" subject="&lt;kere.ws&gt; - MP3 - 1344347688 - High_On_Fire-The_Art_Of_Self_Defense-(Remastered)-2012-FNT - [21/21] - &quot;11-high_on_fire-master_of_fists_(demo).mp3&quot; yEnc (1/27)">
<groups><group>alt.binaries.cores</group><group>alt.binaries.mom</group></groups>
<segments>
<segment bytes="786553" number="3">1344347694.85858.3@eu.news.astraweb.com</segment>
<segment bytes="786323" number="10">1344347694.90408.10@eu.news.astraweb.com</segment>
</segments>
</file>
</nzb>
`
	result := Nzb{}
	if err := Parse([]byte(input), &result); err != nil {
		fmt.Fprintf(os.Stderr, "Failed parsing nzb %s", err)
		t.FailNow()
	}
	if len(result.Files) != 1 {
		t.Error("failed files")
		t.FailNow()
	}
	if len(result.Files[0].Segments) != 2 {
		t.Error("failed segments on file")
	}
	if result.Files[0].Segments[0].Bytes != 786553 {
		t.Error("failed Files[0].Segment[0].Bytes")
	}
	if result.Files[0].Segments[1].Bytes != 786323 {
		t.Error("failed Files[0].Segment[1].Bytes")
	}
	if result.Files[0].Groups[0] != "alt.binaries.cores" {
		t.Error("failed Files[0].Groups[0]")
	}
	if result.Files[0].Groups[1] != "alt.binaries.mom" {
		t.Error("failed Files[0].Groups[1]")
	}
	if result.Files[0].Filename != "11-high_on_fire-master_of_fists_(demo).mp3" {
		t.Error("failed Files[0].Filename")
	}
}
