package util

import "net/http"

func Headers(req *http.Request, meta map[string]string, headers ...string) {
	for _, h := range headers {
		meta[h] = req.Header.Get(h)
	}
}

func IfSet(key, value string, meta map[string]string) {
	if value != "" {
		meta[key] = value
	}
}

func RemoteAddr(req *http.Request, key string, meta map[string]string) {
	meta[key] = req.RemoteAddr
}

func Digout(req *http.Request, key string, meta map[string]string, digger func(*http.Request) string) {
	IfSet(key, digger(req), meta)
}
