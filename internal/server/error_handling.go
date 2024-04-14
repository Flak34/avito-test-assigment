package server

import (
	"avito-test-assigment/internal/payload"
	"avito-test-assigment/internal/repository"
	"avito-test-assigment/internal/service"
	"encoding/json"
	"errors"
	"net/http"
)

type handlerFunc func(w http.ResponseWriter, r *http.Request) error

func Handle(f handlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			if errors.Is(err, repository.ErrObjectNotFound) {
				w.WriteHeader(http.StatusNotFound)
			} else if errors.Is(err, service.ErrIncorrectData) {
				w.WriteHeader(http.StatusBadRequest)
				w.Header().Set("Content-Type", "application/json")
				resp := payload.ErrorResponse{Error: err.Error()}
				mResp, _ := json.Marshal(&resp)
				w.Write(mResp)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
				w.Header().Set("Content-Type", "application/json")
				resp := payload.ErrorResponse{Error: err.Error()}
				mResp, _ := json.Marshal(&resp)
				w.Write(mResp)
			}
		}
	}
}
