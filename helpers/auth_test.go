//nolint
package helpers

import (
	//"bytes"
	//"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	. "glauth-ui-light/config"
)

func TestI18n(t *testing.T) {
	cfg := WebConfig{
		Locale: Locale{
			Lang: "fr",
			Path: "../locales",
		},
	}
	InitRouterTest(cfg)
	assert.Equal(t, "Mail", Tr("fr", "Mail"), "i18n")
	assert.Equal(t, "Identifiant", Tr("fr", "Login"), "i18n")
}

func TestHelpersSession(t *testing.T) {

	cfg := WebConfig{
		Locale: Locale{
			Lang: "fr",
			Path: "../locales",
		},
		Debug: true,
		Sec: Sec{
			CSRFrandom: "secret",
		},
		CfgUsers: CfgUsers{
			Start:         5000,
			GIDAdmin:      6501,
			GIDcanChgPass: 6500,
		},
		PassPolicy: PassPolicy{
			AllowReadSSHA256: true,
		},
	}

	initUsersValues()
	//fmt.Printf("%+v\n",Data)
	gin.SetMode(gin.TestMode)
	router := InitRouterTest(cfg)

	router.GET("/auth/login", LoginTestHandlerForm)
	router.POST("/auth/login", LoginTestHandler)
	router.GET("/auth/logout", LogoutTestHandler)

	// Public access
	fmt.Println("= Public access")
	respA, url := testCookieAccess(t, router, "GET", "/", nil)
	assert.Equal(t, 200, respA.Code, "http GET public access")
	assert.Equal(t, "/", url, "http GET public access")

	// Login
	fmt.Println("= Logins")
	// user login
	resp, cookie, location := testCookieLogin(t, router, "user", "dogood")
	usercookie := cookie
	fmt.Println(usercookie)
	assert.Equal(t, 200, resp.Code, "http GET success user login")
	//assert.Equal(t, true, strings.Contains(resp.Body.String(), "Welcome user"), "http GET success first access user")
	//assert.Equal(t, true, strings.Contains(resp.Body.String(), "class=\"navbar-brand\">user</span>"), "http GET success first access user")
	// test badlogin
	resp, cookie, location = testCookieLogin(t, router, "baduser", "dogood")
	assert.Equal(t, 200, resp.Code, "http GET success first user profile")
	assert.Equal(t, "/auth/login", location, "Bad login redirect to /auth/login")

	// User access
	fmt.Println("= User access")
	respA, url = testCookieAccess(t, router, "GET", "/auth/logout", usercookie)
	assert.Equal(t, 200, respA.Code, "http GET user access to user profile")
	assert.Equal(t, "/", url, "http GET logout success")
	fmt.Printf("%+v\n", respA)

	/*respA, url = testCookieAccess(t, router, "GET", "/user/5001", usercookie)
	assert.Equal(t, 302, respA.Code, "http GET restrict user access to other profile")
	assert.Equal(t, "/auth/logout", url, "http GET restrict user access to other profile")
	respA, _ = testCookieAccess(t, router, "GET", "/user/5002", usercookie)
	assert.Equal(t, 302, respA.Code, "http GET restrict user access to other profile")
	assert.Equal(t, "/auth/logout", url, "http GET restrict user access to other profile")
	//fmt.Printf("%+v\n", respA)
	*/
}

func testCookieAccess(t *testing.T, router *gin.Engine, method string, url string, cookie []*http.Cookie) (*httptest.ResponseRecorder, string) {
	req, _ := http.NewRequest(method, url, nil)
	if cookie != nil {
		for _, c := range cookie {
			req.Header.Add("Cookie", c.String())
		}
	}
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code == 302 {
		location, _ := resp.Result().Location()
		fmt.Printf("=> Redirect to: %s\n", location.String())
		url = location.String()
		cookie := resp.Result().Cookies()
		req, _ = http.NewRequest("GET", url, nil)
		if len(cookie) != 0 {
			for _, c := range cookie {
				req.Header.Add("Cookie", c.String())
			}
		}
		resp = httptest.NewRecorder()
		router.ServeHTTP(resp, req)
	}
	return resp, url
}

func testCookieLogin(t *testing.T, router *gin.Engine, login string, pass string) (*httptest.ResponseRecorder, []*http.Cookie, string) {
	form := url.Values{}
	form.Add("username", login)
	form.Add("password", pass)
	req, err := http.NewRequest("POST", "/auth/login", strings.NewReader(form.Encode()))
	req.PostForm = form
	req.Header.Add("Content-Type", "application/x-www-form-Urlencoded")
	if err != nil {
		fmt.Println(err)
	}
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	//fmt.Printf("%+v\n",resp)
	if login == "" || pass == "" {
		return resp, nil, ""
	}
	assert.Equal(t, 302, resp.Code, "http POST success redirect to Edit")
	location, _ := resp.Result().Location()
	fmt.Printf("=> Redirect to: %s\n", location.String())
	cookie := resp.Result().Cookies()
	req, _ = http.NewRequest("GET", location.String(), nil)
	for _, c := range cookie {
		req.Header.Add("Cookie", c.String())
	}
	resp = httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	return resp, cookie, location.String()
}
