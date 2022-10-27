package api

import (
	"github.com/group-coldwalle/coinsign/qieusdtserver/util"
	"net/http"
	"strconv"
)

const (
	OK    = 0
	Error = -1
)

func HttpOnlyOK(w http.ResponseWriter) {
	HttpJson(w, http.StatusOK, util.EncodeHttpResopne(OK, []byte{'o', 'k'}, nil))
}

func HttpOnlyError(w http.ResponseWriter) {
	HttpJson(w, http.StatusOK, util.EncodeHttpResopne(Error, nil, nil))
}

func HttpData(w http.ResponseWriter, code int, msg []byte, data map[string]interface{}) {
	HttpJson(w, http.StatusOK, util.EncodeHttpResopne(code, msg, data))
}

func HttpOK(w http.ResponseWriter, data map[string]interface{}) {
	HttpJson(w, http.StatusOK, util.EncodeHttpResopne(OK, []byte{'o', 'k'}, data))
}

func HttpError(w http.ResponseWriter, msg []byte, data map[string]interface{}) {
	HttpJson(w, http.StatusOK, util.EncodeHttpResopne(Error, msg, data))
}

func HttpJson(w http.ResponseWriter, httpStatus int, data []byte) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	w.WriteHeader(httpStatus)
	w.Write(data)
}
func HttpOKData(w http.ResponseWriter, data interface{}) {
	HttpJson(w, http.StatusOK, util.EncodeHttpResopneData(OK, []byte{'o', 'k'}, data))
}

func HttpOKDataHash(w http.ResponseWriter, data interface{}, hash string) {
	HttpJson(w, http.StatusOK, util.EncodeHttpResopneHash(OK, []byte{'o', 'k'}, data, hash))
}
