package handler

import (
	"errors"
	"github.com/go-chi/chi"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"gomarket/config"
	"gomarket/internal/cookies"
	"gomarket/internal/storage"
	servicemocks "gomarket/internal/storage/mocks"
	"gomarket/internal/usecase"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandler_Registration(t *testing.T) {
	type mockBehavior func(r *servicemocks.MockIStorage)
	url := "http://localhost:8080/api/user/register"

	tests := []struct {
		name               string
		mockBehavior       mockBehavior
		expectedStatusCode int
		body               string
	}{
		{
			name: "Ok",
			mockBehavior: func(r *servicemocks.MockIStorage) {
				r.EXPECT().CreateUser("admin", "admin").
					Return(nil).AnyTimes()
			},
			body:               `{"login": "admin", "password": "admin"}`,
			expectedStatusCode: 200,
		},
		{
			name: "Bad Request",
			mockBehavior: func(r *servicemocks.MockIStorage) {
				r.EXPECT().CreateUser("admin", "admin").
					Return(nil).AnyTimes()
			},
			body:               ``,
			expectedStatusCode: 400,
		},
		{
			name: "Internal Server Error",
			mockBehavior: func(r *servicemocks.MockIStorage) {
				r.EXPECT().CreateUser("admin", "admin").
					Return(errors.New("DB error")).AnyTimes()
			},
			body:               `{"login": "admin", "password": "admin"}`,
			expectedStatusCode: 500,
		},
		{
			name: "User Already Exists",
			mockBehavior: func(r *servicemocks.MockIStorage) {
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
			repos := servicemocks.NewMockIStorage(c)
			test.mockBehavior(repos)
			logic := usecase.New(repos)
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
	type mockBehavior func(r *servicemocks.MockIStorage)
	url := "http://localhost:8080/api/user/login"
	tests := []struct {
		name               string
		mockBehavior       mockBehavior
		expectedStatusCode int
		body               string
	}{
		{
			name: "Ok",
			mockBehavior: func(r *servicemocks.MockIStorage) {
				r.EXPECT().CheckPassword("admin", "admin").
					Return(nil).AnyTimes()
			},
			body:               `{"login": "admin", "password": "admin"}`,
			expectedStatusCode: 200,
		},
		{
			name: "Bad Request",
			mockBehavior: func(r *servicemocks.MockIStorage) {
				r.EXPECT().CheckPassword("admin", "admin").
					Return(nil).AnyTimes()
			},
			body:               ``,
			expectedStatusCode: 400,
		},
		{
			name: "Internal Server Error",
			mockBehavior: func(r *servicemocks.MockIStorage) {
				r.EXPECT().CheckPassword("admin", "admin").
					Return(errors.New("DB error")).AnyTimes()
			},
			body:               `{"login": "admin", "password": "admin"}`,
			expectedStatusCode: 500,
		},
		{
			name: "Unauthorized",
			mockBehavior: func(r *servicemocks.MockIStorage) {
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
			repos := servicemocks.NewMockIStorage(c)
			test.mockBehavior(repos)
			logic := usecase.New(repos)
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
	type mockBehavior func(r *servicemocks.MockIStorage)
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
			mockBehavior: func(r *servicemocks.MockIStorage) {
				r.EXPECT().CheckID("admin", "12345678903").
					Return(nil).AnyTimes()
			},
			body:               `12345678903`,
			expectedStatusCode: 202,
		},
		{
			name: "Already created by this user",
			mockBehavior: func(r *servicemocks.MockIStorage) {
				r.EXPECT().CheckID("admin", "12345678903").
					Return(storage.ErrCreatedByThisUser).AnyTimes()
			},
			body:               `12345678903`,
			expectedStatusCode: 200,
		},
		{
			name: "Already created by another user",
			mockBehavior: func(r *servicemocks.MockIStorage) {
				r.EXPECT().CheckID("admin", "12345678903").
					Return(storage.ErrCreatedByAnotherUser).AnyTimes()
			},
			body:               `12345678903`,
			expectedStatusCode: 409,
		},
		{
			name: "Bad format",
			mockBehavior: func(r *servicemocks.MockIStorage) {
				r.EXPECT().CheckID("admin", "12345678902").
					Return(storage.ErrBadID).AnyTimes()
			},
			body:               `12345678902`,
			expectedStatusCode: 422,
		},
		{
			name: "Internal Server Error",
			mockBehavior: func(r *servicemocks.MockIStorage) {
				r.EXPECT().CheckID("admin", "12345678903").
					Return(errors.New("DB Error")).AnyTimes()
			},
			body:               `12345678903`,
			expectedStatusCode: 500,
		},
		{
			name: "Unauthorized",
			mockBehavior: func(r *servicemocks.MockIStorage) {
				r.EXPECT().CheckID("", "").
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
			repos := servicemocks.NewMockIStorage(c)
			test.mockBehavior(repos)
			logic := usecase.New(repos)
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
