package api

import (
	"net/http"
)

func Default(r *http.Request) (int, string) {
  return http.StatusNotFound, 
}
