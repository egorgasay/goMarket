package cookies

import (
	"github.com/docker/distribution/uuid"
	"gomarket/internal/loyalty/cookies"
	"log"
	"net/http"
	"time"
)

func SetCookie(readyCookie ...string) *http.Cookie {
	log.Println("Setting new cookie...")
	fid := uuid.Generate()
	cookie := new(http.Cookie)
	cookie.Name = "session"
	if len(readyCookie) > 0 {
		cookie.Value = readyCookie[0]
	} else {
		cookie.Value = cookies.NewCookie(fid.String()).Value + fid.String()
	}
	cookie.Expires = time.Now().Add(24 * time.Hour * 365)
	return cookie
}
