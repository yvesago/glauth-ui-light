package helpers

import (
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"

	"github.com/gorilla/securecookie"
)

var blockKey = securecookie.GenerateRandomKey(32)

var CookieSessionName = "appsession"

func SetSession(c *gin.Context, status string) {
	session := sessions.Default(c)
	session.Set("status", status)
	session.Save() //nolint:errcheck // no session
}

func MiddlewareSession(secure bool) gin.HandlerFunc {
	// store := cookie.NewStore([]byte(secret))
	store := cookie.NewStore(blockKey)
	store.Options(sessions.Options{
		//Domain:   "localhost",
		Path:     "/auth/",
		HttpOnly: true,
		Secure:   secure,
		MaxAge:   3600,
		SameSite: http.SameSiteStrictMode,
	})
	return sessions.Sessions(CookieSessionName, store)
}

func GetUserID(c *gin.Context) (userName string, userId string) {
	s := GetSession(c)
	return s.User, s.UserID
}

func GetSession(c *gin.Context) Status {
	session := sessions.Default(c)
	var s = Status{}
	t := session.Get("status")
	if t != nil {
		s = StrToStatus(t.(string))
	}
	return s
}

func ClearSession(c *gin.Context) {
	cookie := &http.Cookie{
		Name:   CookieSessionName,
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	}

	http.SetCookie(c.Writer, cookie)
}

// Encodage de la valeur du cookie.
func encode(value string) string {
	encode := &url.URL{Path: value}
	return encode.String()
}

// Décodage de la valeur du cookie.
func decode(value string) string {
	decode, _ := url.QueryUnescape(value)
	return decode
}

func SetFlashCookie(c *gin.Context, name string, value string) {
	cookie := &http.Cookie{
		Name:     name,
		Value:    encode(value),
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
		MaxAge:   1,
	}

	http.SetCookie(c.Writer, cookie)
}

func GetFlashCookie(c *gin.Context, name string) (value string) {
	cookie, err := c.Request.Cookie(name)

	var cookieValue string
	if err == nil {
		cookieValue = cookie.Value
	}

	return decode(cookieValue)
}
