package main

import (
	"context"
	"dsock"
	"github.com/docker/docker/api/types"
	"logger"
	"net/http"
	"srv"
	"time"
)

var (
	D *dsock.Docker
	log *logger.Logger
)

func init() {
	logger.Init(logger.DEBUG)
	log=logger.GetLogger()
	srv.SetLogger(log)
	var e error
	D, e = dsock.NewDockerClient("1.38")
	if e != nil {
		panic(e)
	}
}
func main() {
	srv.NewService("dockerApi", "/api/docker", func(w http.ResponseWriter, r *http.Request) {
		c, e := D.ContainerList(context.Background(), types.ContainerListOptions{})
		if e != nil {
			srv.WriteJsonError(w, http.StatusInternalServerError, e, "连接Docker服务失败")
		}
		srv.WriteJson(w, 200, c)
	})
	srv.ListenAndServe(":80")
	srv.WaitOsShutdown(5 * time.Second)
}
