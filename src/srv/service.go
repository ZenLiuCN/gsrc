package srv

import (
	"context"
	"encoding/json"
	"github.com/NYTimes/gziphandler"
	mux2 "github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"logger"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

var (
	srv    = make(map[string]Service)
	Mux    *mux2.Router
	Server *http.Server
)

func init() {
	Mux = mux2.NewRouter()
}

type Service interface {
	Name() string
	ServeHTTP(http.ResponseWriter, *http.Request)
	Disable()
	Enable()
	Status() bool
	Register(*mux2.Router)
}

func RegisterService(service Service) {
	service.Register(Mux)
	srv[service.Name()] = service
}
func DisableService(name string) bool {
	if s, ok := srv[name]; ok {
		s.Disable()
		return true
	}
	return false
}
func EnableService(name string) bool {
	if s, ok := srv[name]; ok {
		s.Enable()
		return true
	}
	return false
}
func ServiceStatus(name string) int {
	if s, ok := srv[name]; ok {
		if s.Status() {
			return 1
		}
		return 0
	}
	return -1
}
func ServicesName() (keys []string) {
	keys = make([]string, 0, len(srv))
	for k := range srv {
		keys = append(keys, k)
	}
	return
}
func ServicesStatus() (ser map[string]bool) {
	ser = make(map[string]bool)
	for key, value := range srv {
		ser[key] = value.Status()
	}
	return
}

type ServiceFunc struct {
	name    string
	path    string
	enable  bool
	Handler http.HandlerFunc
}

func (s ServiceFunc) Name() string {
	return s.name
}
func (s ServiceFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !s.enable {
		w.WriteHeader(404)
		return
	}
	s.Handler(w, r)
}
func (s *ServiceFunc) Disable() {
	s.enable = false
}
func (s *ServiceFunc) Enable() {
	s.enable = true
}
func (s ServiceFunc) Status() bool {
	return s.enable
}
func (s ServiceFunc) Register(r *mux2.Router) {
	r.Path(s.path).Handler(&s)
}

func NewService(name string, path string, handlerFunc http.HandlerFunc) *ServiceFunc {
	s := new(ServiceFunc)
	s.name = name
	s.path = path
	s.enable = true
	s.Handler = handlerFunc
	RegisterService(s)
	return s
}
func NewServiceHandler(name string, path string, handlerFunc http.Handler) {
	Mux.Path(path).Handler(handlerFunc)
}
func NewResourceService(path string, handlerFunc http.Handler) {
	Mux.PathPrefix(path).Handler(handlerFunc)
}
func NewResourceServiceDir(path, prefix, rootDir string, gzip bool, level int) {
	if gzip {
		Mux.Use(gziphandler.MustNewGzipLevelHandler(level))
	}
	Mux.PathPrefix(path).Handler(http.StripPrefix(prefix, http.FileServer(http.Dir(rootDir))))
}
func NewWebsocketService(name string, path string, handler WsHandlerFunction, readBufferSize, writeBufferSize int) *ServiceFunc {
	if readBufferSize == 0 {
		readBufferSize = 1024
	}
	if writeBufferSize == 0 {
		writeBufferSize = 1024
	}
	s := new(ServiceFunc)
	s.name = name
	s.path = path
	s.enable = true
	s.Handler = func(w http.ResponseWriter, r *http.Request) {
		log.Infof("Get Ws Connection ")
		var upgrader = websocket.Upgrader{
			ReadBufferSize:  readBufferSize,
			WriteBufferSize: writeBufferSize,
			CheckOrigin:     func(r *http.Request) bool { return true },
		}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			_ = handler(nil, err, -1, nil)
			return
		}
		c := NewWsConnJson(conn)
		_ = handler(c, nil, 0, nil)
		for {
			messageType, r, err := conn.NextReader()
			if err != nil {
				log.Errorf(`read data error %v`, err)
				if e := handler(c, err, messageType, r); e != nil {
					return
				} else {
					return
				}
			}
			if e := handler(c, err, messageType, r); e != nil {
				return
			} else {
				continue
			}
		}
	}
	RegisterService(s)
	return s
}
func NewMiddleWare(fun mux2.MiddlewareFunc) {
	Mux.Use(fun)
}

func ListenAndServe(addr string) {
	Mux.Use(loggerMiddleWare)
	Server = &http.Server{
		Addr:         addr,
		Handler:      Mux,
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
	}
	go func() {
		if er := Server.ListenAndServe(); er != nil &&er!=http.ErrServerClosed{
			panic(er)
		}
	}()
}
func Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()
	e := Server.Shutdown(ctx)
	Server = nil
	return e
}
func WaitOsShutdown(timeout time.Duration) {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	log.Infoln("Shutdown with timeout: %v", timeout)
	if err := Server.Shutdown(ctx); err != nil {
		logger.Errorf("Error: %v", err)
	} else {
		logger.Infoln("Server stopped")
	}
}
func IsRunning() bool {
	if Server == nil {
		return false
	}
	return true
}

func loggerMiddleWare(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Infof("incoming request %+v", jsonPretty(dumpRequest(r)) )
		next.ServeHTTP(w, r)
	})
}

func HasService(name string) bool {
	for k := range srv {
		if k == name {
			return true
		}
	}
	return false
}
func HasEnableService(name string) bool {
	for k, v := range srv {
		if k == name && v.Status() {
			return true
		}
	}
	return false
}

func dumpHeader(r *http.Request) (h []string) {
	for k, v := range r.Header {
		h = append(h, k + `:` + strings.Join(v,","))
	}
	return
}
func dumpRequest(r *http.Request) (q map[string]interface{}) {
	q=make(map[string]interface{})
	q[`url`]=r.RequestURI
	q[`header`]=dumpHeader(r)
	q[`remote`]=r.RemoteAddr
	if x:=r.Header.Get(`X-Forwarded-For`);len(x)!=0{
		q[`remote`]=x
	}
	return
}

func jsonPretty(obj interface{})string{
	d,_:=json.MarshalIndent(obj,""," ")
	return string(d)
}