package handler

import (
	"encoding/json"
	"errors"
	"github.com/go-chi/httplog"
	"gomarket/internal/loyalty/config"
	"gomarket/internal/loyalty/cookies"
	"gomarket/internal/loyalty/schema"
	"gomarket/internal/loyalty/storage"
	"gomarket/internal/loyalty/usecase"
	"gomarket/pkg/bettererror"
	"io"
	"net/http"
)

type Handler struct {
	conf  *config.Config
	logic usecase.IUseCase
}

func NewHandler(cfg *config.Config, logic usecase.IUseCase) *Handler {
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
		w.Write(bettererror.New(err).SetAppLayer(bettererror.Handler).JSON())
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
		if err != nil {
			if err == storage.ErrUsernameConflict {
				w.WriteHeader(http.StatusConflict)
				w.Write(bettererror.New(err).SetAppLayer(bettererror.Logic).JSON())
				return
			}
			oplog := httplog.LogEntry(r.Context())
			oplog.Error().Msg(err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(bettererror.New(err).SetAppLayer(bettererror.Storage).JSON())
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
		if err == storage.ErrWrongPassword {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write(bettererror.New(err).SetAppLayer(bettererror.Logic).JSON())
			return
		} else if err != nil {
			oplog := httplog.LogEntry(r.Context())
			oplog.Error().Msg(err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(bettererror.New(err).SetAppLayer(bettererror.Storage).JSON())
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
			w.Write(bettererror.New(err).SetAppLayer(bettererror.Handler).JSON())
			return
		}

		id, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(bettererror.New(err).SetAppLayer(bettererror.Handler).JSON())
			return
		}

		err = h.logic.CheckID(h.conf.AccrualSystemAddress, cookie, string(id))
		if errors.Is(err, storage.ErrCreatedByThisUser) {
			w.WriteHeader(http.StatusOK)
			w.Write(bettererror.New(err).SetAppLayer(bettererror.Logic).JSON())
			return
		}

		if errors.Is(err, storage.ErrCreatedByAnotherUser) {
			w.WriteHeader(http.StatusConflict)
			w.Write(bettererror.New(err).SetAppLayer(bettererror.Logic).JSON())
			return
		}

		if errors.Is(err, storage.ErrBadID) {
			w.WriteHeader(http.StatusUnprocessableEntity)
			w.Write(bettererror.New(err).SetAppLayer(bettererror.Logic).JSON())
			return
		}

		if err != nil {
			oplog := httplog.LogEntry(r.Context())
			oplog.Error().Msg(err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(bettererror.New(err).SetAppLayer(bettererror.Storage).JSON())
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
			oplog := httplog.LogEntry(r.Context())
			oplog.Error().Msg(err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(bettererror.New(err).SetAppLayer(bettererror.Handler).JSON())
			return
		}

		orders, err := h.logic.GetOrders(cookie)
		if errors.Is(err, storage.ErrNoResult) {
			w.WriteHeader(http.StatusNoContent)
			w.Write(bettererror.New(err).SetAppLayer(bettererror.Logic).JSON())
			return
		}

		if err != nil {
			oplog := httplog.LogEntry(r.Context())
			oplog.Error().Msg(err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(bettererror.New(err).SetAppLayer(bettererror.Storage).JSON())
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(orders)
	}
}

func (h Handler) GetBalance() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		cookie, err := cookies.Get(r)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(bettererror.New(err).SetAppLayer(bettererror.Handler).JSON())
			return
		}

		balance, err := h.logic.GetBalance(cookie)
		if err != nil {
			oplog := httplog.LogEntry(r.Context())
			oplog.Error().Msg(err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(bettererror.New(err).SetAppLayer(bettererror.Storage).JSON())
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(balance)
	}
}

func (h Handler) PostWithdraw() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		cookie, err := cookies.Get(r)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(bettererror.New(err).SetAppLayer(bettererror.Handler).JSON())
			return
		}

		var withdrawn schema.WithdrawnRequest
		err = BindJSON(w, r, &withdrawn)
		if err != nil {
			return
		}

		err = h.logic.DrawBonuses(cookie, withdrawn.Sum, withdrawn.Order)
		if errors.Is(err, storage.ErrNotEnoughMoney) {
			w.WriteHeader(http.StatusPaymentRequired)
			w.Write(bettererror.New(err).SetAppLayer(bettererror.Logic).JSON())
			return
		}

		if errors.Is(err, storage.ErrBadID) {
			w.WriteHeader(http.StatusUnprocessableEntity)
			w.Write(bettererror.New(err).SetAppLayer(bettererror.Logic).JSON())
			return
		}

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			oplog := httplog.LogEntry(r.Context())
			oplog.Error().Msg(err.Error())
			w.Write(bettererror.New(err).SetAppLayer(bettererror.Storage).JSON())
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func (h Handler) GetWithdrawals() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		cookie, err := cookies.Get(r)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(bettererror.New(err).SetAppLayer(bettererror.Handler).JSON())
			return
		}

		withdrawals, err := h.logic.GetWithdrawals(cookie)
		if errors.Is(err, storage.ErrNoWithdrawals) {
			w.WriteHeader(http.StatusNoContent)
			w.Write(bettererror.New(err).SetAppLayer(bettererror.Logic).JSON())
			return
		}

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			oplog := httplog.LogEntry(r.Context())
			oplog.Error().Msg(err.Error())
			w.Write(bettererror.New(err).SetAppLayer(bettererror.Storage).JSON())
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(withdrawals)
	}
}
