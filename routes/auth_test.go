//nolint
package routes

import (
	//"bytes"
	//"encoding/json"
	"fmt"
	//"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	//"regexp"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	. "glauth-ui-light/config"
	. "glauth-ui-light/handlers"
)

func initUsersValues() {
	v1 := User{
		Name:         "user",
		UIDNumber:    5000,
		PrimaryGroup: 6500,
		PassSHA256:   "6478579e37aff45f013e14eeb30b3cc56c72ccdc310123bcdf53e0333e3f416a", //dogood
	}
	Data.Users = append(Data.Users, v1)
	v2 := User{
		Name:         "admin",
		UIDNumber:    5001,
		PrimaryGroup: 6501,
		PassSHA256:   "6478579e37aff45f013e14eeb30b3cc56c72ccdc310123bcdf53e0333e3f416a",
	}
	Data.Users = append(Data.Users, v2)
	v3 := User{
		Name:         "serviceapp",
		UIDNumber:    5002,
		PrimaryGroup: 6502,
		PassSHA256:   "6478579e37aff45f013e14eeb30b3cc56c72ccdc310123bcdf53e0333e3f416a",
	}
	Data.Users = append(Data.Users, v3)
}

func TestSession(t *testing.T) {

	cfg := WebConfig{
		Locale: Locale{
			Lang: "en",
			Path: "../locales",
		},
		Debug:   true,
		Verbose: false,
		Sec: Sec{
			CSRFrandom:  "secret",
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
	router := SetRoutes(&cfg)

	// Public access
	fmt.Println("= Public access")
	respA, url := testAccess(t, router, "GET", "/", nil)
	assert.Equal(t, 200, respA.Code, "http GET public access")
	assert.Equal(t, "/", url, "http GET public access")

	respA, url = testAccess(t, router, "GET", "/favicon.ico", nil)
	headers := respA.Header()
	//fmt.Printf("=====\n%+v\n",headers["Cache-Control"][0])
	assert.Equal(t, "public, max-age=604800, immutable", headers["Cache-Control"][0], "http GET cached headers")

	// Login
	fmt.Println("= Logins")
	// user login
	resp, cookie, location := testLogin(t, router, "user", "dogood")
	usercookie := cookie
	assert.Equal(t, 200, resp.Code, "http GET success user login")
	//assert.Equal(t, true, strings.Contains(resp.Body.String(), "Welcome user"), "http GET success first access user")
	assert.Equal(t, true, strings.Contains(resp.Body.String(), "class=\"navbar-brand\">user</span>"), "http GET success first access user")
	// test badlogin
	resp, cookie, location = testLogin(t, router, "baduser", "dogood")
	assert.Equal(t, 200, resp.Code, "http GET success first user profile")
	assert.Equal(t, "/auth/login", location, "Bad login redirect to /login")

	// admin login
	resp, cookie, location = testLogin(t, router, "admin", "dogood")
	admincookie := cookie
	assert.Equal(t, 200, resp.Code, "http GET success admin login")
	//assert.Equal(t, true, strings.Contains(resp.Body.String(), "Welcome admin"), "http GET success first access admin")
	assert.Equal(t, true, strings.Contains(resp.Body.String(), "class=\"navbar-brand\">admin</span>"), "http GET success first access user")

	// serviceapp login
	resp, cookie, location = testLogin(t, router, "serviceapp", "dogood")
	serviceappcookie := cookie
	assert.Equal(t, 200, resp.Code, "http GET success serviceapp login")
	//assert.Equal(t, true, strings.Contains(resp.Body.String(), "Welcome serviceapp"), "http GET success first access serviceapp")
	assert.Equal(t, true, strings.Contains(resp.Body.String(), "class=\"navbar-brand\">serviceapp</span>"), "http GET success first access user")

	// Admin access
	fmt.Println("= Admin access")
	Url := "/auth"
	respA, url = testAccess(t, router, "GET", Url+"/crud/user/", usercookie)
	assert.Equal(t, 302, respA.Code, "http GET bad user access to admin url")
	assert.Equal(t, "/auth/logout", url, "http GET bad user access to admin url")
	//fmt.Printf("%+v\n", respA)
	respA, url = testAccess(t, router, "GET", Url+"/crud/user/", admincookie)
	assert.Equal(t, 200, respA.Code, "http GET admin access to admin url")
	assert.Equal(t, Url+"/crud/user/", url, "http GET admin access to admin url")
	respA, url = testAccess(t, router, "GET", Url+"/user/5001", admincookie)
	assert.Equal(t, 200, respA.Code, "http GET admin access to profile")
	assert.Equal(t, Url+"/user/5001", url, "http GET admin access to profile")
	assert.Equal(t, true, strings.Contains(respA.Body.String(), ">Change password</button>"), "http GET success allow Change pass")
	respA, url = testAccess(t, router, "GET", Url+"/user/5000", admincookie)
	assert.Equal(t, 200, respA.Code, "http GET admin access to user profile")
	assert.Equal(t, Url+"/user/5000", url, "http GET admin access to user profile")

	// User access
	fmt.Println("= User access")
	respA, _ = testAccess(t, router, "GET", Url+"/user/5000", usercookie)
	assert.Equal(t, 200, respA.Code, "http GET user access to user profile")
	assert.Equal(t, true, strings.Contains(respA.Body.String(), ">Change password</button>"), "http GET success allow Change pass")

	respA, url = testAccess(t, router, "GET", Url+"/user/5001", usercookie)
	assert.Equal(t, 302, respA.Code, "http GET restrict user access to other profile")
	assert.Equal(t, "/auth/logout", url, "http GET restrict user access to other profile")
	respA, _ = testAccess(t, router, "GET", Url+"/user/5002", usercookie)
	assert.Equal(t, 302, respA.Code, "http GET restrict user access to other profile")
	assert.Equal(t, "/auth/logout", url, "http GET restrict user access to other profile")
	//fmt.Printf("%+v\n", respA)

	// Service access
	fmt.Println("= Service access")
	respA, _ = testAccess(t, router, "GET", Url+"/user/5002", serviceappcookie)
	assert.Equal(t, 200, respA.Code, "http GET user access to serviceapp profile")
	assert.Equal(t, false, strings.Contains(respA.Body.String(), ">Change password</button>"), "http GET success don't show Change pass")
	//fmt.Printf("%+v\n", respA)

	var oldcookie []*http.Cookie
	badcookie := &http.Cookie{
		Name:   "session",
		Value:  "",
		Path:   Url + "/",
		MaxAge: 1,
	}
	oldcookie = append(oldcookie, badcookie)

	respA, url = testAccess(t, router, "GET", Url+"/user/5002", oldcookie)
	assert.Equal(t, 302, respA.Code, "http GET reject access with old or bad cookie")
	assert.Equal(t, "/auth/logout", url, "http GET reject access with old or bad cookie")
}

func testAccess(t *testing.T, router *gin.Engine, method string, url string, cookie []*http.Cookie) (*httptest.ResponseRecorder, string) {
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

func testLogin(t *testing.T, router *gin.Engine, login string, pass string) (*httptest.ResponseRecorder, []*http.Cookie, string) {
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
