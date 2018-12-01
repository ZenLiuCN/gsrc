package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFun(t *testing.T) {
	arch := `i686`
	mark := `/mingw32`
	root := `d:\Dev\env\pkg\`
	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.Contains(path, arch) && strings.Contains(info.Name(), `.pc`) {
			t.Log(parentN(path, 3), path, info.Name())
			prefix := parentN(path, 3)
			f, e := ioutil.ReadFile(path)
			if e != nil {
				panic(e)
			}
			e = ioutil.WriteFile(filepath.Join(root, info.Name()), []byte(strings.Replace(string(f), mark, strings.Replace(prefix, "\\", "/", -1), -1)), os.ModePerm)
			if e != nil {
				panic(e)
			}
		}
		return nil
	})
}

