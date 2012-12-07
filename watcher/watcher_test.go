package watcher

import (
	"io/ioutil"
	"os"
	"path"
	"testing"
	"time"
)

type FakeFile struct {
	s string
}

func (s *FakeFile) Name() string       { return s.s }
func (s *FakeFile) Size() int64        { return 0 }
func (s *FakeFile) Mode() os.FileMode  { return 0 }
func (s *FakeFile) ModTime() time.Time { return time.Now() }
func (s *FakeFile) IsDir() bool        { return false }
func (s *FakeFile) Sys() interface{}   { return nil }

func TestDifference(t *testing.T) {
	previous := []os.FileInfo{
		&FakeFile{"c"},
		&FakeFile{"e"},
	}
	current := []os.FileInfo{
		&FakeFile{"a"},
		&FakeFile{"b"},
		&FakeFile{"e"},
	}
	diff := difference(previous, current)
	if len(diff) != 2 {
		t.Error("expected 2 items")
	}
}

func TestFilterFiles(t *testing.T) {
	files := []os.FileInfo{
		&FakeFile{"foo.nzb"},
		&FakeFile{"blah.txt"},
	}
	filtered := filterNzbs(files)
	if filtered[0].Name() != "foo.nzb" {
		t.Error("expected foo.nzb")
	}
	if len(filtered) != 1 {
		t.Errorf("expected only foo.nzb, got %v", filtered)
	}
}

func TestStartWatching(t *testing.T) {
	dir, _ := ioutil.TempDir("", "temp_queue")
	defer os.Remove(dir)

	newFiles := StartWatching(dir)
	defer close(newFiles)

	fileA := "a.nzb"
	os.Create(path.Join(dir, fileA))
	select {
	case a := <-newFiles:
		if len(a) != 1 {
			t.Fatal("expected a.nzb only")
		}
		if a[0].Name() != fileA {
			t.Errorf("expected %s, got %s", fileA, a[0].Name())
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for a.nzb")
	}

	fileB := "b.nzb"
	os.Create(path.Join(dir, fileB))
	select {
	case b := <-newFiles:
		if len(b) != 1 {
			t.Fatal("expected b.nzb only")
		}
		if b[0].Name() != fileB {
			t.Errorf("expected %s, got %s", fileB, b[0].Name())
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for b.nzb")
	}
}
