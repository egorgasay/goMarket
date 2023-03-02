package cookies

import (
	"gomarket/internal/loyalty/cookies"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
)

var lastUID int64 = 0
var idMu = &sync.Mutex{}

func SetCookie() *http.Cookie {
	log.Println("Setting new cookie...")
	idMu.Lock()
	lastUID++
	var uid = lastUID
	idMu.Unlock()

	fid := strconv.FormatInt(uid, 10)
	cookie := new(http.Cookie)
	cookie.Name = "session"
	cookie.Value = cookies.NewCookie(fid).Value + fid
	cookie.Expires = time.Now().Add(24 * time.Hour * 365)
	return cookie
}
