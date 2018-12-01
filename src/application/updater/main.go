package main

import (
	`archive/zip`
	"bufio"
	`context`
	"fmt"
	"gabs"
	`io`
	"io/ioutil"
	`log`
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	`strconv`
	"strings"
	"syscall"
	"time"
)

var (
	versionURL, appname, appext string
)

func main() {
	f, e := ioutil.ReadFile(filepath.Join(root(), "application"))
	if e != nil {
		messageBox(`程序损坏`, "程序组件损坏,请重新安装系统!", MB_OK|MB_ICONERROR)
		return
	}
	gg, e := gabs.ParseJSON(f)
	if e != nil {
		messageBox(`程序损坏`, "程序组件损坏,请重新安装系统!", MB_OK|MB_ICONERROR)
		return
	}
	versionURL, e = gg.GetString("update")
	if e != nil {
		messageBox(`程序损坏`, "程序组件损坏,请重新安装系统!", MB_OK|MB_ICONERROR)
		return
	}
	appname, e = gg.GetString("app")
	if e != nil {
		messageBox(`程序损坏`, "程序组件损坏,请重新安装系统!", MB_OK|MB_ICONERROR)
		return
	}
	appext, e = gg.GetString("param")
	if e != nil {
		messageBox(`程序损坏`, "程序组件损坏,请重新安装系统!", MB_OK|MB_ICONERROR)
		return
	}
	http.DefaultClient.Timeout = 5 * time.Second
	r, e := http.Get(versionURL)
	switch {
	case e == nil:
		break
	case strings.Contains(e.Error(), `net/http: request canceled`) || strings.Contains(e.Error(), `net/http: timeout`):
		messageBox(`错误`, `无法连接更新服务器.请稍后重试`, MB_OK|MB_ICONERROR|MB_SYSTEMMODAL)
		return
	default:
		fmt.Println(e)
		return
	}
	defer r.Body.Close()
	b, e := ioutil.ReadAll(r.Body)
	if e != nil {
		fmt.Println(e)
		return
	}
	g, e := gabs.ParseJSON(b)
	if e != nil {
		fmt.Println(e)
		return
	}
	v, e := g.GetInt("version")
	if e != nil {
		fmt.Println(e)
		return
	}
	uri, e := g.GetString("url")
	if e != nil {
		fmt.Println(e)
		return
	}
	desc, e := g.GetString("desc")
	if e != nil {
		fmt.Println(e)
		return
	}
	vr, e := ioutil.ReadFile(filepath.Join(root(), "version"))
	var (
		vo int
		vg *gabs.Container
	)
	if e != nil && !os.IsNotExist(e) {
		return
	} else if e != nil && os.IsNotExist(e) {
		vo = -1
		goto down
	}
	vg, e = gabs.ParseJSON(vr)
	if e != nil {
		vo = -1
		goto down
		return
	}
	vo, e = vg.GetInt("version")
	if e != nil {
		vo = -1
		goto down
		return
	}
	if vo < v {
		goto down
	} else {
		return
	}
down:
	rbx := messageBox(`程序更新`, fmt.Sprintf("程序更新:\n有新版本程序<版本v %d>,当前版本<v %d>.\n\n%s\n\n\t\t\t是否更新? \n ", v, vo, desc), MB_OKCANCEL|MB_ICONQUESTION|MB_SYSTEMMODAL|MB_SETFOREGROUND)
	if rbx != IDOK {
		return
	}
	ch := make(chan bool)
	go downLoad(uri, v)
	ShowDialog(ch)
	defer DestroyWindow(procBox)
	for {
		msg := tMSG{}
		gotMessage, err := getMessage(&msg, 0, 0, 0)
		if err != nil {
			log.Println(err)
			return
		}

		if gotMessage {
			translateMessage(&msg)
			dispatchMessage(&msg)
		} else {
			break
		}
	}
}
func root() string {
	f, e := filepath.Abs(os.Args[0])
	if e != nil {
		return "."
	}
	f = filepath.Dir(f)
	return f
}

func downLoad(uri string, vn int) {
	updateProcess(`结束运行中的程序...`, 5)
	fmt.Println(`结束运行中的程序`)
	//kill process
	cmd := exec.Command("cmd", "/c", "taskkill /F /T /IM "+appname)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	cmd.Start()
	time.Sleep(1 * time.Second)
	cmd.Start()
	updateProcess(`开始下载程序包...`, 15)

	http.DefaultClient.Timeout = 5 * time.Second
	rs, e := http.Head(uri)
	size, e := strconv.Atoi(rs.Header.Get("Content-Length"))
	if e != nil {
		panic(e)
	}
	f, e := ioutil.TempFile("", "update.")
	if e != nil {
		return
	}
	defer f.Close()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				break
			default:
				fi, err := f.Stat()
				if err != nil {
					log.Fatal(err)
				}
				sz := fi.Size()
				if sz == 0 {
					sz = 1
				}
				updateProcess(`下载程序包...`, int(float64(sz)/float64(size)*100)-15-15)
				time.Sleep(500 * time.Millisecond)
			}
		}

	}(ctx)
	rs, e = http.Get(uri)
	if e != nil {
		return
	}
	defer rs.Body.Close()
	//f, e := os.OpenFile(filepath.Join(root(), appname), os.O_CREATE|os.O_TRUNC, 0666)
	bw := bufio.NewWriter(f)
	bw.ReadFrom(rs.Body)
	bw.Flush()
	updateProcess(`开始解压文件...`, 100-15)
	 e = unZip(f.Name(), root())
	if e != nil {
		updateProcess(`解压文件失败!`, 100)
		return
	}
	os.Remove(f.Name())
	g := gabs.New()
	g.Set(vn, "version")
	g.Set(time.Now(), "date")
	e = ioutil.WriteFile(filepath.Join(root(), "version"), g.Bytes(), 0666)
	if e != nil {
		fmt.Println("write version error", e)
		return
	}
	updateProcess(`更新完成`, 100)
	fmt.Println(`完成更新`)
	PostMessage(procBtn, WM_ENABLE, 1, NullPtr)
	return
}

func unZip(src, des string) error {
	zipReader, _ := zip.OpenReader(src)
	for _, file := range zipReader.Reader.File {
		if strings.Contains(file.Name,"updater.exe"){
			continue
		}
		zippedFile, err := file.Open()
		if err != nil {
			log.Println(err)
			return err
		}
		defer zippedFile.Close()
		targetDir := des
		extractedFilePath := filepath.Join(
			targetDir,
			file.Name,
		)

		if file.FileInfo().IsDir() {
			log.Println("Directory Created:", extractedFilePath)
			os.MkdirAll(extractedFilePath, file.Mode())
		} else {
			log.Println("File extracted:", file.Name)
			outputFile, err := os.OpenFile(
				extractedFilePath,
				os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
				file.Mode(),
			)
			if err != nil {
				log.Println(err)
				return err
			}
			defer outputFile.Close()
			_, err = io.Copy(outputFile, zippedFile)
			if err != nil {
				log.Println(err)
				return err
			}
		}
	}
	return nil
}
