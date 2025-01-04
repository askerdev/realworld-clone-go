package handler

import (
	"encoding/json"
	"io"
	"net/http"
)

func JSON(w http.ResponseWriter, val any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(val)
}

func ParseBody(req io.ReadCloser, dest any) error {
	return json.NewDecoder(req).Decode(dest)
}
