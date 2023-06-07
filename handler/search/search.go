package search

import (
	"fmt"
	"net/http"

	"github.com/FayeZheng0/ask_pubmed/core"
	"github.com/FayeZheng0/ask_pubmed/handler/render"
	"github.com/fox-one/pkg/httputil/param"
	"github.com/go-chi/chi"
)

func Post(service core.SearchResultService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Query     string  `json:"query" valid:"minstringlength(1),required"`
			Amount    int     `json:"amount,omitempty"`
			Threshold float32 `json:"threshold,omitempty"`
		}

		if err := param.Binding(r, &body); err != nil {
			render.Error(w, http.StatusBadRequest, err)
			return
		}

		ctx := r.Context()
		err := service.SetSearchParams(ctx, body.Query, r.RemoteAddr, body.Amount, body.Threshold)
		if err != nil {
			render.Error(w, http.StatusInternalServerError, fmt.Errorf(http.StatusText(http.StatusInternalServerError)))
			return
		}
		searchRes, err := service.GetSearchResults(ctx)
		if err != nil {
			render.Error(w, http.StatusInternalServerError, fmt.Errorf(http.StatusText(http.StatusInternalServerError)))
			return
		}
		render.JSON(w, searchRes)
	}
}

func Get(service core.SearchResultService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := chi.URLParam(r, "query")
		ctx := r.Context()

		err := service.SetSearchParams(ctx, query, r.RemoteAddr, 0, 0)
		if err != nil {
			render.Error(w, http.StatusInternalServerError, fmt.Errorf(http.StatusText(http.StatusInternalServerError)))
			return
		}
		searchRes, err := service.GetSearchResults(ctx)
		if err != nil {
			render.Error(w, http.StatusInternalServerError, fmt.Errorf(http.StatusText(http.StatusInternalServerError)))
			return
		}
		render.JSON(w, render.H{
			"resq": searchRes,
		})
	}
}
