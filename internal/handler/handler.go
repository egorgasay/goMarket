package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"gomarket/config"
	"gomarket/internal/cookies"
	"gomarket/internal/schema"
	"gomarket/internal/storage/service"
	"gomarket/internal/usecase"
	"io"
	"net/http"
)

type Handler struct {
	conf  *config.Config
	logic usecase.UseCase
}

func NewHandler(cfg *config.Config, logic usecase.UseCase) *Handler {
	if cfg == nil {
		panic("конфиг равен nil")
	}

	return &Handler{conf: cfg, logic: logic}
}

func BindJSON(w http.ResponseWriter, r *http.Request, obj any) error {
	decoder := json.NewDecoder(r.Body)

	err := decoder.Decode(obj)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf(`{"error": "%s"}`, err)))
		return err
	}

	return nil
}

func (h Handler) PostRegister() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var cred schema.AuthRequestJSON
		err := BindJSON(w, r, &cred)
		if err != nil {
			return
		}

		err = h.logic.CreateUser(cred.Login, cred.Password)
		if err == service.ErrUsernameConflict {
			w.WriteHeader(http.StatusConflict)
			w.Write([]byte(fmt.Sprintf(`{"error": "%s"}`, err)))
			return
		} else if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf(`{"error": "%s"}`, err)))
			return
		}

		cookies.Set(w, cred.Login)
		w.WriteHeader(http.StatusOK)
	}
}

func (h Handler) PostLogin() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var cred schema.AuthRequestJSON
		err := BindJSON(w, r, &cred)
		if err != nil {
			return
		}

		err = h.logic.CheckPassword(cred.Login, cred.Password)
		if err == service.ErrWrongPassword {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(fmt.Sprintf(`{"error": "%s"}`, err)))
			return
		} else if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf(`{"error": "%s"}`, err)))
			return
		}

		cookies.Set(w, cred.Login)
		w.WriteHeader(http.StatusOK)
	}
}

func (h Handler) PostOrders() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := cookies.Get(r)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf(`{"error": "%s"}`, err)))
			return
		}

		id, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf(`{"error": "%s"}`, err)))
			return
		}

		err = h.logic.CheckID(h.conf.AccrualSystemAddress, cookie, string(id))
		if errors.Is(err, service.ErrCreatedByThisUser) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(fmt.Sprintf(`{"msg": "%s"}`, err)))
			return
		}

		if errors.Is(err, service.ErrCreatedByAnotherUser) {
			w.WriteHeader(http.StatusConflict)
			w.Write([]byte(fmt.Sprintf(`{"error": "%s"}`, err)))
			return
		}

		if errors.Is(err, service.ErrBadID) {
			w.WriteHeader(http.StatusUnprocessableEntity)
			w.Write([]byte(fmt.Sprintf(`{"error": "%s"}`, err)))
			return
		}

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf(`{"error": "%s"}`, err)))
			return
		}

		w.WriteHeader(http.StatusAccepted)
	}
}

func (h Handler) GetUserOrders() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		cookie, err := cookies.Get(r)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf(`{"error": "%s"}`, err)))
			return
		}

		orders, err := h.logic.GetOrders(cookie)
		if errors.Is(err, service.ErrNoResult) {
			w.WriteHeader(http.StatusNoContent)
			w.Write([]byte(fmt.Sprintf(`{"msg": "%s"}`, err)))
			return
		}

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf(`{"error": "%s"}`, err)))
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(orders)
	}
}
func (h Handler) GetOrder() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}

func (h Handler) GetBalance() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		cookie, err := cookies.Get(r)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf(`{"error": "%s"}`, err)))
			return
		}

		balance, err := h.logic.GetBalance(cookie)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf(`{"error": "%s"}`, err)))
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(balance)
	}
}

func (h Handler) PostWithdraw() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}

func (h Handler) GetWithdrawals() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}
