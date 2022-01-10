package api

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	js "github.com/mmanjoura/template/serializer/json"
	appUser "github.com/mmanjoura/template/user"
	"github.com/pkg/errors"
)

type UserHandler interface {
	FindUserByID(http.ResponseWriter, *http.Request)
	FindUsers(http.ResponseWriter, *http.Request)
	CreateUser(http.ResponseWriter, *http.Request)
	UpdateUser(http.ResponseWriter, *http.Request)
	DeleteUser(http.ResponseWriter, *http.Request)
}

// type handler struct {
// 	userService appUser.UserService
// }

// func NewHandler(userService appUser.UserService) UserHandler {
// 	return &handler{userService: userService}
// }

func setupResponse(w http.ResponseWriter, contentType string, body []byte, statusCode int) {
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(statusCode)
	_, err := w.Write(body)
	if err != nil {
		log.Println(err)
	}
}

func (s *Server) userSerializer(contentType string) appUser.UserSerializer {

	return &js.User{}
}
func (s *Server) userUpdateserializer(contentType string) appUser.UserUpdateSerializer {

	return &js.UserUpdate{}
}

func (s *Server) FindUserByID(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	vars := mux.Vars(r)
	userId, err := strconv.Atoi(vars["id"])
	if err != nil {
		log.Println(err)
	}
	u, err := s.AuthService.FindAuthByID(r.Context(), userId)
	if err != nil {
		log.Println(err)
	}
	responseBody, err := json.Marshal(u)

	if err != nil {
		if errors.Cause(err) == appUser.ErrUserNotFound {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	setupResponse(w, contentType, responseBody, http.StatusOK)
}

func (s *Server) FindUsers(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")

	// this filter is coming from ui
	filter := appUser.UserFilter{}
	filter.Limit = 3
	users, _, err := s.UserService.FindUsers(r.Context(), filter)

	if err != nil {
		if errors.Cause(err) == appUser.ErrUserNotFound {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	responseBody, err := json.Marshal(users)

	if err != nil {
		return
	}

	setupResponse(w, contentType, responseBody, http.StatusOK)

}

func (s *Server) CreateUser(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	requestBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	u, err := s.userSerializer(contentType).DecodeUser(requestBody)

	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	err = s.UserService.CreateUser(r.Context(), u)
	if err != nil {
		if errors.Cause(err) == appUser.ErrUserInvalid {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	responseBody, err := json.Marshal(u)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	setupResponse(w, contentType, responseBody, http.StatusCreated)
}

func (s *Server) UpdateUser(w http.ResponseWriter, r *http.Request) {

	contentType := r.Header.Get("Content-Type")
	requestBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	vars := mux.Vars(r)
	userId, err := strconv.Atoi(vars["id"])
	if err != nil {
		log.Println(err)
	}

	u, err := s.userSerializer(contentType).DecodeUser(requestBody)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	upd := appUser.UserUpdate{}
	upd.Email = &u.Email
	upd.Name = &u.Name

	_, err = s.UserService.UpdateUser(r.Context(), userId, upd)
	if err != nil {
		if errors.Cause(err) == appUser.ErrUserInvalid {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	responseBody, err := s.userUpdateserializer(contentType).EncodeUserUpdate(&upd)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	setupResponse(w, contentType, responseBody, http.StatusCreated)
}

func (s *Server) DeleteUser(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	vars := mux.Vars(r)
	userId, err := strconv.Atoi(vars["id"])
	if err != nil {
		log.Println(err)
	}
	err = s.UserService.DeleteUser(r.Context(), userId)

	if err != nil {
		if errors.Cause(err) == appUser.ErrUserNotFound {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	setupResponse(w, contentType, nil, http.StatusOK)
}
