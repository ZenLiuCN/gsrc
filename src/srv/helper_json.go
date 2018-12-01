package srv

import (
	"encoding/json"
	"gabs"
	"net/http"
	"time"
)

func WriteJson(w http.ResponseWriter, status int, obj interface{}) {
	d, e := json.Marshal(obj)
	if e != nil {
		WriteJsonError(w, http.StatusInternalServerError, e, "")
		return
	}
	w.Header().Set(HEAD_Content_Type, MIME_JSON+CHARSET_UTF8)
	_, _ = w.Write(d)
	return
}
func WriteJsonError(w http.ResponseWriter, status int, er error, humanMessage string) {
	w.WriteHeader(status)
	w.Header().Set(HEAD_Content_Type, MIME_JSON+CHARSET_UTF8)
	d, _ := json.Marshal(map[string]interface{}{
		`timestamp`: time.Now().Unix(),
		`error`:     er.Error(),
		`message`:   humanMessage,
	})
	_, _ = w.Write(d)
}
func WriteGabs(w http.ResponseWriter, mime string, status int, obj *gabs.Container) {
	w.Header().Set(HEAD_Content_Type, mime+CHARSET_UTF8)
	w.WriteHeader(status)
	_, _ = w.Write(obj.Bytes())
	return
}
func WriteGabsJson(w http.ResponseWriter, obj *gabs.Container) {
	w.Header().Set(HEAD_Content_Type, MIME_JSON+CHARSET_UTF8)
	w.WriteHeader(200)
	_, _ = w.Write(obj.Bytes())
	return
}
