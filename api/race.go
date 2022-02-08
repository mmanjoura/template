package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/mmanjoura/template/race"
	js "github.com/mmanjoura/template/serializer/json"
	"github.com/pkg/errors"
)

type RaceHandler interface {
	FindRacecards(http.ResponseWriter, *http.Request)
	FindResults(http.ResponseWriter, *http.Request)
	FindDetailInfo(http.ResponseWriter, *http.Request)
}

func setupResponse(w http.ResponseWriter, contentType string, body []byte, statusCode int) {
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(statusCode)
	_, err := w.Write(body)
	if err != nil {
		log.Println(err)
	}
}

func (s *Server) CardSerializer(contentType string) race.CardSerializer {

	return &js.Card{}
}

func (s *Server) ResultSerializer(contentType string) race.ResultSerializer {

	return &js.Result{}
}

func (s *Server) DetailInfoSerializer(contentType string) race.DetailInfoSerializer {

	return &js.DetailInfo{}
}

func (s *Server) FindRacecards(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")

	filter := race.RaceFilter{}
	filter.Limit = 3
	raceCards, _, err := s.RaceService.FindRacecards(r.Context(), filter)

	if err != nil {
		if errors.Cause(err) == race.ErrCardNotFound {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	responseBody, err := json.Marshal(raceCards)

	if err != nil {
		return
	}

	setupResponse(w, contentType, responseBody, http.StatusOK)

}
func (s *Server) FindResults(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")

	filter := race.RaceFilter{}
	filter.Limit = 3
	results, _, err := s.RaceService.FindResults(r.Context(), filter)

	if err != nil {
		if errors.Cause(err) == race.ErrResultNotFound {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	responseBody, err := json.Marshal(results)

	if err != nil {
		return
	}

	setupResponse(w, contentType, responseBody, http.StatusOK)

}
func (s *Server) FindDetailInfo(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")

	filter := race.RaceFilter{}
	filter.Limit = 3
	detailInfo, _, err := s.RaceService.FindDetailInfo(r.Context(), filter)

	if err != nil {
		if errors.Cause(err) == race.ErrDetailInfoNotFound {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	responseBody, err := json.Marshal(detailInfo)

	if err != nil {
		return
	}

	setupResponse(w, contentType, responseBody, http.StatusOK)

}
