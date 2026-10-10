package httpx

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// WriteJSON — для net/http.
func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if v == nil {
		return
	}
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("write json", "err", err)
	}
}

// WriteError — для net/http.
func WriteError(w http.ResponseWriter, status int, code, msg string) {
	WriteJSON(w, status, ErrorBody{Code: code, Message: msg})
}

// GinError — для gin.
func GinError(c *gin.Context, status int, code, msg string) {
	c.AbortWithStatusJSON(status, ErrorBody{Code: code, Message: msg})
}
