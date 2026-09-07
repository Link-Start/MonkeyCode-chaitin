package apikey

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (s *Service) RegisterAdmin(router chi.Router) {
	router.Get("/api-keys", func(w http.ResponseWriter, r *http.Request) {
		keys, err := s.AdminList(r.Context(), r.URL.Query().Get("user_id"))
		if err != nil {
			keyError(w, http.StatusInternalServerError, "读取调用密钥失败")
			return
		}
		keyJSON(w, http.StatusOK, map[string]any{"api_keys": keys})
	})
	router.Delete("/api-keys/{keyID}", func(w http.ResponseWriter, r *http.Request) {
		if err := s.AdminRevoke(r.Context(), chi.URLParam(r, "keyID")); err != nil {
			status := http.StatusInternalServerError
			if errors.Is(err, ErrNotFound) {
				status = http.StatusNotFound
			}
			keyError(w, status, "调用密钥不存在")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
}
