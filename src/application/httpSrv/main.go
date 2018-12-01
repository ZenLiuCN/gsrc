package main

import (
	"encoding/json"
	"fmt"
	"logger"
	"net/http"
	"os"
	"path"
	"service"
	"strconv"
	"time"
)

var (
	log           *logger.Logger
	shutdownWait  time.Duration
	listenAddr    string
	staticPath    string
	logFile       string
	logRotateSize int
)

func init() {
	var e error
	shutdownWait, e = time.ParseDuration(os.Getenv("SHUTDOWN_DELAY"))
	if e != nil {
		shutdownWait = 5 * time.Second
	}
	var port int
	port, e = strconv.Atoi(os.Getenv("LISTEN_PORT"))
	if e != nil || port > 65535 {
		listenAddr = `:80`
	} else {
		listenAddr = `:` + strconv.Itoa(port)
	}
	logRotateSize, e = strconv.Atoi(os.Getenv("LOG_ROTATE_SIZE"))
	if e != nil {
		logRotateSize = int(1024 * logger.KB)
	}
	logFile = os.Getenv("LOG_PATH")
	if len(logFile) == 0 {
		logFile = `./logs`
	}
	logFile = path.Join(logFile, `server_`+os.Getenv("HOSTNAME")+`.log`)
	staticPath = os.Getenv("STATIC_PATH")
	if len(staticPath) == 0 {
		staticPath = `./web`
	}
	//region timezone for docker/scratch
	var loc *time.Location
	loc, e = time.LoadLocation(os.Getenv("TIME_ZONE"))
	if e != nil {
		loc, e = time.LoadLocation("Asia/Shanghai")
		if e != nil {
			panic(e)
		}
	}
	//endregion
	time.Local = loc
	w, _ := logger.NewRotateWriter(logFile, uint64(logRotateSize), time.Second*10)
	logger.Init(logger.INFO, w, os.Stdout)
	log = logger.GetLogger()
	service.SetLogger(log)

}
func main() {
	service.NewService(`api`, `/health`, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		env,_:=json.Marshal(os.Environ())
		_, _ = w.Write([]byte(fmt.Sprintf(`{"containerId":"%s","timeZone":"%s","logPath":"%s","staticPath":"%s","serverAddress":"%s","shutdownTime":"%s","logRotateSize":"%d byte","osEnvironment":%s}`,
			os.Getenv("HOSTNAME"),
			os.Getenv("TIME_ZONE"),
			logFile,
			staticPath,
			listenAddr,
			shutdownWait,
			logRotateSize,
			string(env),
		)))
	})
	service.NewResourceServiceDir(`/`, `/`, staticPath,true,3)
	service.ListenAndServe(listenAddr)
	service.WaitOsShutdown(shutdownWait)
}
