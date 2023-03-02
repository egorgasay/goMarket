package handler

import (
	"errors"
	"github.com/go-chi/chi"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"gomarket/config"
	"gomarket/internal/cookies"
	"gomarket/internal/storage"
	servicemocks "gomarket/internal/usecase/mocks"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandler_Registration(t *testing.T) {
	type mockBehavior func(r *servicemocks.MockIUseCase)
	url := "http://localhost:8080/api/user/register"

	tests := []struct {
		name               string
		mockBehavior       mockBehavior
		expectedStatusCode int
		body               string
	}{
		{
			name: "Ok",
			mockBehavior: func(r *servicemocks.MockIUseCase) {
				r.EXPECT().CreateUser("admin", "admin").
					Return(nil).AnyTimes()
			},
			body:               `{"login": "admin", "password": "admin"}`,
			expectedStatusCode: 200,
		},
		{
			name: "Bad Request",
			mockBehavior: func(r *servicemocks.MockIUseCase) {
				r.EXPECT().CreateUser("admin", "admin").
					Return(nil).AnyTimes()
			},
			body:               ``,
			expectedStatusCode: 400,
		},
		{
			name: "Internal Server Error",
			mockBehavior: func(r *servicemocks.MockIUseCase) {
				r.EXPECT().CreateUser("admin", "admin").
					Return(errors.New("DB error")).AnyTimes()
			},
			body:               `{"login": "admin", "password": "admin"}`,
			expectedStatusCode: 500,
		},
		{
			name: "User Already Exists",
			mockBehavior: func(r *servicemocks.MockIUseCase) {
				r.EXPECT().CreateUser("admin", "admin").
					Return(storage.ErrUsernameConflict).AnyTimes()
			},
			body:               `{"login": "admin", "password": "admin"}`,
			expectedStatusCode: 409,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()
			logic := servicemocks.NewMockIUseCase(c)
			test.mockBehavior(logic)
			cfg := config.New()
			h := NewHandler(cfg, logic)

			r := httptest.NewRequest(http.MethodPost, url, strings.NewReader(test.body))
			w := httptest.NewRecorder()
			router := chi.NewRouter()

			router.Group(h.PublicRoutes)
			router.Group(h.PrivateRoutes)
			router.ServeHTTP(w, r)

			// Assert
			assert.Equal(t, test.expectedStatusCode, w.Code)
		})
	}
}

func TestHandler_Login(t *testing.T) {
	type mockBehavior func(r *servicemocks.MockIUseCase)
	url := "http://localhost:8080/api/user/login"
	tests := []struct {
		name               string
		mockBehavior       mockBehavior
		expectedStatusCode int
		body               string
	}{
		{
			name: "Ok",
			mockBehavior: func(r *servicemocks.MockIUseCase) {
				r.EXPECT().CheckPassword("admin", "admin").
					Return(nil).AnyTimes()
			},
			body:               `{"login": "admin", "password": "admin"}`,
			expectedStatusCode: 200,
		},
		{
			name: "Bad Request",
			mockBehavior: func(r *servicemocks.MockIUseCase) {
				r.EXPECT().CheckPassword("admin", "admin").
					Return(nil).AnyTimes()
			},
			body:               ``,
			expectedStatusCode: 400,
		},
		{
			name: "Internal Server Error",
			mockBehavior: func(r *servicemocks.MockIUseCase) {
				r.EXPECT().CheckPassword("admin", "admin").
					Return(errors.New("DB error")).AnyTimes()
			},
			body:               `{"login": "admin", "password": "admin"}`,
			expectedStatusCode: 500,
		},
		{
			name: "Unauthorized",
			mockBehavior: func(r *servicemocks.MockIUseCase) {
				r.EXPECT().CheckPassword("admin", "admin").
					Return(storage.ErrWrongPassword).AnyTimes()
			},
			body:               `{"login": "admin", "password": "admin"}`,
			expectedStatusCode: 401,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()
			logic := servicemocks.NewMockIUseCase(c)
			test.mockBehavior(logic)
			cfg := config.New()
			h := NewHandler(cfg, logic)

			r := httptest.NewRequest(http.MethodPost, url, strings.NewReader(test.body))
			w := httptest.NewRecorder()
			router := chi.NewRouter()

			router.Group(h.PublicRoutes)
			router.Group(h.PrivateRoutes)
			router.ServeHTTP(w, r)

			// Assert
			assert.Equal(t, test.expectedStatusCode, w.Code)
		})
	}
}

func TestHandler_PostOrders(t *testing.T) {
	type mockBehavior func(r *servicemocks.MockIUseCase)
	url := "http://localhost:8080/api/user/orders"
	tests := []struct {
		name               string
		mockBehavior       mockBehavior
		expectedStatusCode int
		body               string
		dontNeedCookie     bool
	}{
		{
			name: "Ok",
			mockBehavior: func(r *servicemocks.MockIUseCase) {
				r.EXPECT().CheckID("", "8d5f8aeeb64e3ce20b537d04c486407eaf489646617cfcf493e76f5b794fa080-61646d696e", "12345678903").
					Return(nil).AnyTimes()
			},
			body:               `12345678903`,
			expectedStatusCode: 202,
		},
		{
			name: "Already created by this user",
			mockBehavior: func(r *servicemocks.MockIUseCase) {
				r.EXPECT().CheckID("", "8d5f8aeeb64e3ce20b537d04c486407eaf489646617cfcf493e76f5b794fa080-61646d696e", "12345678903").
					Return(storage.ErrCreatedByThisUser).AnyTimes()
			},
			body:               `12345678903`,
			expectedStatusCode: 200,
		},
		{
			name: "Already created by another user",
			mockBehavior: func(r *servicemocks.MockIUseCase) {
				r.EXPECT().CheckID("", "8d5f8aeeb64e3ce20b537d04c486407eaf489646617cfcf493e76f5b794fa080-61646d696e", "12345678903").
					Return(storage.ErrCreatedByAnotherUser).AnyTimes()
			},
			body:               `12345678903`,
			expectedStatusCode: 409,
		},
		{
			name: "Bad format",
			mockBehavior: func(r *servicemocks.MockIUseCase) {
				r.EXPECT().CheckID("", "8d5f8aeeb64e3ce20b537d04c486407eaf489646617cfcf493e76f5b794fa080-61646d696e", "12345678902").
					Return(storage.ErrBadID).AnyTimes()
			},
			body:               `12345678902`,
			expectedStatusCode: 422,
		},
		{
			name: "Internal Server Error",
			mockBehavior: func(r *servicemocks.MockIUseCase) {
				r.EXPECT().CheckID("", "8d5f8aeeb64e3ce20b537d04c486407eaf489646617cfcf493e76f5b794fa080-61646d696e", "12345678903").
					Return(errors.New("DB Error")).AnyTimes()
			},
			body:               `12345678903`,
			expectedStatusCode: 500,
		},
		{
			name: "Unauthorized",
			mockBehavior: func(r *servicemocks.MockIUseCase) {
				r.EXPECT().CheckID("", "", "").
					Return(errors.New("")).AnyTimes()
			},
			body:               `12345678903`,
			expectedStatusCode: 401,
			dontNeedCookie:     true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()
			logic := servicemocks.NewMockIUseCase(c)
			test.mockBehavior(logic)
			cfg := config.New()
			h := NewHandler(cfg, logic)

			r := httptest.NewRequest(http.MethodPost, url, strings.NewReader(test.body))
			if !test.dontNeedCookie {
				cookie := cookies.NewCookie("admin")
				r.AddCookie(cookie)
			}
			w := httptest.NewRecorder()
			router := chi.NewRouter()

			router.Group(h.PublicRoutes)
			router.Group(h.PrivateRoutes)
			router.ServeHTTP(w, r)

			// Assert
			assert.Equal(t, test.expectedStatusCode, w.Code)
		})
	}
}

func TestHandler_GetUserOrders(t *testing.T) {
	type mockBehavior func(r *servicemocks.MockIUseCase)
	url := "http://localhost:8080/api/user/orders"
	tests := []struct {
		name               string
		mockBehavior       mockBehavior
		expectedStatusCode int
		isEmpty            bool
	}{
		{
			name: "Ok",
			mockBehavior: func(r *servicemocks.MockIUseCase) {
				r.EXPECT().GetOrders("8d5f8aeeb64e3ce20b537d04c486407eaf489646617cfcf493e76f5b794fa080-61646d696e").
					Return([]byte(""), nil).AnyTimes()
			},
			expectedStatusCode: 200,
		},
		{
			name: "Empty",
			mockBehavior: func(r *servicemocks.MockIUseCase) {
				r.EXPECT().GetOrders("6a3fa3a06a653f65e901e58dc0882a11aa6ae29bf1bbf8e6c2754e2551b50bb0-61646d696e32").
					Return([]byte(""), storage.ErrNoResult).AnyTimes()
			},
			isEmpty:            true,
			expectedStatusCode: 204,
		},
		{
			name: "Internal Server Error",
			mockBehavior: func(r *servicemocks.MockIUseCase) {
				r.EXPECT().GetOrders("8d5f8aeeb64e3ce20b537d04c486407eaf489646617cfcf493e76f5b794fa080-61646d696e").
					Return([]byte(""), errors.New("DB Error")).AnyTimes()
			},
			expectedStatusCode: 500,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()
			logic := servicemocks.NewMockIUseCase(c)
			test.mockBehavior(logic)
			cfg := config.New()
			h := NewHandler(cfg, logic)

			r := httptest.NewRequest(http.MethodGet, url, strings.NewReader(""))
			if test.isEmpty {
				cookie := cookies.NewCookie("admin2")
				r.AddCookie(cookie)
			} else {
				cookie := cookies.NewCookie("admin")
				r.AddCookie(cookie)
			}
			w := httptest.NewRecorder()
			router := chi.NewRouter()

			router.Group(h.PublicRoutes)
			router.Group(h.PrivateRoutes)
			router.ServeHTTP(w, r)

			// Assert
			assert.Equal(t, test.expectedStatusCode, w.Code)
		})
	}
}

func TestHandler_GetBalance(t *testing.T) {
	type mockBehavior func(r *servicemocks.MockIUseCase)
	url := "http://localhost:8080/api/user/balance"
	tests := []struct {
		name               string
		mockBehavior       mockBehavior
		expectedStatusCode int
		body               string
		dontNeedCookie     bool
	}{
		{
			name: "Ok",
			mockBehavior: func(r *servicemocks.MockIUseCase) {
				r.EXPECT().GetBalance("8d5f8aeeb64e3ce20b537d04c486407eaf489646617cfcf493e76f5b794fa080-61646d696e").
					Return([]byte(""), nil).AnyTimes()
			},
			expectedStatusCode: 200,
		},
		{
			name: "Err with db",
			mockBehavior: func(r *servicemocks.MockIUseCase) {
				r.EXPECT().GetBalance("8d5f8aeeb64e3ce20b537d04c486407eaf489646617cfcf493e76f5b794fa080-61646d696e").
					Return([]byte(""), errors.New("err with DB")).AnyTimes()
			},
			expectedStatusCode: 500,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()
			logic := servicemocks.NewMockIUseCase(c)
			test.mockBehavior(logic)
			cfg := config.New()
			h := NewHandler(cfg, logic)

			r := httptest.NewRequest(http.MethodGet, url, strings.NewReader(test.body))
			if !test.dontNeedCookie {
				cookie := cookies.NewCookie("admin")
				r.AddCookie(cookie)
			}
			w := httptest.NewRecorder()
			router := chi.NewRouter()

			router.Group(h.PublicRoutes)
			router.Group(h.PrivateRoutes)
			router.ServeHTTP(w, r)

			// Assert
			assert.Equal(t, test.expectedStatusCode, w.Code)
		})
	}
}

func TestHandler_PostWithdraw(t *testing.T) {
	type mockBehavior func(r *servicemocks.MockIUseCase)
	url := "http://localhost:8080/api/user/balance/withdraw"
	tests := []struct {
		name               string
		mockBehavior       mockBehavior
		body               string
		expectedStatusCode int
	}{
		{
			name: "Ok",
			mockBehavior: func(r *servicemocks.MockIUseCase) {
				r.EXPECT().DrawBonuses("8d5f8aeeb64e3ce20b537d04c486407eaf489646617cfcf493e76f5b794fa080-61646d696e", 751.0, "2377225624").
					Return(nil).AnyTimes()
			},
			body:               "{\"order\": \"2377225624\",\"sum\": 751} ",
			expectedStatusCode: 200,
		},
		{
			name: "Not Enough Funds",
			mockBehavior: func(r *servicemocks.MockIUseCase) {
				r.EXPECT().DrawBonuses("8d5f8aeeb64e3ce20b537d04c486407eaf489646617cfcf493e76f5b794fa080-61646d696e", 751.0, "2377225624").
					Return(storage.ErrNotEnoughMoney).AnyTimes()
			},
			body:               "{\"order\": \"2377225624\",\"sum\": 751} ",
			expectedStatusCode: 402,
		},
		{
			name: "Wrong ID",
			mockBehavior: func(r *servicemocks.MockIUseCase) {
				r.EXPECT().DrawBonuses("8d5f8aeeb64e3ce20b537d04c486407eaf489646617cfcf493e76f5b794fa080-61646d696e", 751.0, "2377225624").
					Return(storage.ErrBadID).AnyTimes()
			},
			body:               "{\"order\": \"2377225624\",\"sum\": 751} ",
			expectedStatusCode: 422,
		},
		{
			name: "Internal Server Error",
			mockBehavior: func(r *servicemocks.MockIUseCase) {
				r.EXPECT().DrawBonuses("8d5f8aeeb64e3ce20b537d04c486407eaf489646617cfcf493e76f5b794fa080-61646d696e", 751.0, "2377225624").
					Return(errors.New("DB Error")).AnyTimes()
			},
			body:               "{\"order\": \"2377225624\",\"sum\": 751} ",
			expectedStatusCode: 500,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()
			logic := servicemocks.NewMockIUseCase(c)
			test.mockBehavior(logic)
			cfg := config.New()
			h := NewHandler(cfg, logic)

			r := httptest.NewRequest(http.MethodPost, url, strings.NewReader(test.body))
			cookie := cookies.NewCookie("admin")
			r.AddCookie(cookie)
			w := httptest.NewRecorder()
			router := chi.NewRouter()

			router.Group(h.PublicRoutes)
			router.Group(h.PrivateRoutes)
			router.ServeHTTP(w, r)

			// Assert
			log.Println(w.Body)
			assert.Equal(t, test.expectedStatusCode, w.Code)
		})
	}
}

func TestHandler_GetWithdrawals(t *testing.T) {
	type mockBehavior func(r *servicemocks.MockIUseCase)
	url := "http://localhost:8080/api/user/withdrawals"
	tests := []struct {
		name               string
		mockBehavior       mockBehavior
		expectedStatusCode int
		body               string
		dontNeedCookie     bool
	}{
		{
			name: "Ok",
			mockBehavior: func(r *servicemocks.MockIUseCase) {
				r.EXPECT().GetWithdrawals("8d5f8aeeb64e3ce20b537d04c486407eaf489646617cfcf493e76f5b794fa080-61646d696e").
					Return([]byte(""), nil).AnyTimes()
			},
			expectedStatusCode: 200,
		},
		{
			name: "No content",
			mockBehavior: func(r *servicemocks.MockIUseCase) {
				r.EXPECT().GetWithdrawals("8d5f8aeeb64e3ce20b537d04c486407eaf489646617cfcf493e76f5b794fa080-61646d696e").
					Return([]byte(""), storage.ErrNoWithdrawals).AnyTimes()
			},
			expectedStatusCode: 204,
		},
		{
			name: "Err with db",
			mockBehavior: func(r *servicemocks.MockIUseCase) {
				r.EXPECT().GetWithdrawals("8d5f8aeeb64e3ce20b537d04c486407eaf489646617cfcf493e76f5b794fa080-61646d696e").
					Return([]byte(""), errors.New("err with DB")).AnyTimes()
			},
			expectedStatusCode: 500,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()
			logic := servicemocks.NewMockIUseCase(c)
			test.mockBehavior(logic)
			cfg := config.New()
			h := NewHandler(cfg, logic)

			r := httptest.NewRequest(http.MethodGet, url, strings.NewReader(test.body))
			if !test.dontNeedCookie {
				cookie := cookies.NewCookie("admin")
				r.AddCookie(cookie)
			}
			w := httptest.NewRecorder()
			router := chi.NewRouter()

			router.Group(h.PublicRoutes)
			router.Group(h.PrivateRoutes)
			router.ServeHTTP(w, r)

			// Assert
			assert.Equal(t, test.expectedStatusCode, w.Code)
		})
	}
}
