package server

import (
	"avito-test-assigment/internal/config"
	"avito-test-assigment/internal/payload"
	"avito-test-assigment/internal/utils/jwt_utils"
	"encoding/base64"
	"encoding/json"
	"golang.org/x/crypto/bcrypt"
	"io"
	"net/http"
)

type AuthServer struct {
	authCfg config.Auth
}

func NewAuthServer(authCfg config.Auth) *AuthServer {
	return &AuthServer{authCfg: authCfg}
}

func (s *AuthServer) Login(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		var errResponse payload.ErrorResponse
		errResponse.Error = err.Error()
		resp, _ := json.Marshal(&errResponse)
		w.Write(resp)
		return
	}

	var unm payload.LoginRequest
	err = json.Unmarshal(body, &unm)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		var errResponse payload.ErrorResponse
		errResponse.Error = err.Error()
		resp, _ := json.Marshal(&errResponse)
		w.Write(resp)
		return
	}

	secretKey, err := base64.StdEncoding.DecodeString(s.authCfg.SecretKey)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		var errResponse payload.ErrorResponse
		errResponse.Error = err.Error()
		resp, _ := json.Marshal(&errResponse)
		w.Write(resp)
		return
	}

	var response payload.LoginResponse
	if unm.Login == s.authCfg.UserCreds.Login {
		err = bcrypt.CompareHashAndPassword([]byte(s.authCfg.UserCreds.Password), []byte(unm.Password))
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		response.Token, err = jwt_utils.CreateToken("user", string(secretKey))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			var errResponse payload.ErrorResponse
			errResponse.Error = err.Error()
			resp, _ := json.Marshal(&errResponse)
			w.Write(resp)
			return
		}

	} else if unm.Login == s.authCfg.AdminCreds.Login {
		err = bcrypt.CompareHashAndPassword([]byte(s.authCfg.AdminCreds.Password), []byte(unm.Password))
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		response.Token, err = jwt_utils.CreateToken("admin", string(secretKey))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			var errResponse payload.ErrorResponse
			errResponse.Error = err.Error()
			resp, _ := json.Marshal(&errResponse)
			w.Write(resp)
			return
		}
	} else {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	respBytes, _ := json.Marshal(&response)
	w.WriteHeader(http.StatusOK)
	w.Write(respBytes)
}
