package cookies

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"
	"time"
)

var secretkey []byte = []byte("change me")

func SetSecret(secret []byte) {
	secretkey = secret
}

func NewCookie(username string) *http.Cookie {
	h := hmac.New(sha256.New, secretkey)
	src := []byte(username)
	h.Write(src)

	value := hex.EncodeToString(h.Sum(nil)) + "-" + hex.EncodeToString(src)
	cookie := &http.Cookie{
		Name:       "session",
		Value:      value,
		Path:       "",
		Domain:     "localhost",
		Expires:    time.Time{},
		RawExpires: "",
		MaxAge:     3600,
		Secure:     false,
		HttpOnly:   true,
		SameSite:   0,
		Raw:        "",
		Unparsed:   nil,
	}

	return cookie
}

func Get(r *http.Request) (cookie string, err error) {
	cookie = r.Header.Get("Authorization")
	if cookie != "" {
		return cookie, nil
	}

	sessionCookie, err := r.Cookie("session")
	if err == nil && sessionCookie.Value != "" {
		return sessionCookie.Value, nil
	}

	return "", errors.New("no cookies was provided")
}

func Set(w http.ResponseWriter, username string) {
	cookie := NewCookie(username)
	w.Header().Set("Authorization", cookie.Value)
}

func Check(cookie string) bool {
	arr := strings.Split(cookie, "-")

	if len(arr) < 2 {
		return false
	}

	k, v := arr[0], arr[1]

	sign, err := hex.DecodeString(k)
	if err != nil {
		return false
	}

	data, err := hex.DecodeString(v)
	if err != nil {
		return false
	}

	h := hmac.New(sha256.New, secretkey)
	h.Write(data)

	return hmac.Equal(sign, h.Sum(nil))
}
