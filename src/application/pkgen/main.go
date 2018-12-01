package main

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var arch, mark string
var root string

func init() {
	p, _ := filepath.Abs(os.Args[0])
	root = filepath.Dir(p)
}
func main() {
	if len(os.Args)==2{
		arch=os.Args[1]
	}
	log.Printf("root %s arch %s",root,arch)
	arch = strings.ToLower(arch)
	switch arch {
	case `i686`:
		mark = "/mingw32"
	case `x86_64`:
		mark = "/mingw64"
	default:
		arch = `i686`
		mark = "/mingw32"
	}
	log.Printf("root %s arch %s",root,arch)
	os.RemoveAll(filepath.Join(root,"*.pc"))
	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.Contains(path, arch) && strings.Contains(info.Name(), `.pc`) {
			prefix := parentN(path, 3)
			f, e := ioutil.ReadFile(path)
			if e != nil {
				panic(e)
			}
			e = ioutil.WriteFile(filepath.Join(root, info.Name()), []byte(strings.Replace(string(f), mark, strings.Replace(prefix, "\\", "/", -1), -1)), os.ModePerm)
			if e != nil {
				panic(e)
			}
			log.Printf("gen pkg-config %s of prefix %s ", info.Name(),prefix)
		}
		return nil
	})
}

func parentN(path string, n int) (r string) {
	r = path
	for i := 0; i < n; i++ {
		r = filepath.Dir(r)
	}
	return r
}
