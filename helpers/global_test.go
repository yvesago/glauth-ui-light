//nolint
package helpers

/**
 Common functions for helpers tests
**/

import (
	"fmt"
	"html/template"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/kataras/i18n"
	"github.com/stretchr/testify/assert"

	. "glauth-ui-light/config"
)

// tools

var Data Ctmp

func resetData() {
	Data.Users = []User{}
	Data.Groups = []Group{}
}

// testAccessSimple : access without cookie
func testAccessSimple(t *testing.T, router *gin.Engine, method string, url string) (*httptest.ResponseRecorder, string) {
	req, _ := http.NewRequest(method, url, nil)
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

// testAccess with cookie from login
func testAccess(t *testing.T, router *gin.Engine, method string, testurl string) (*httptest.ResponseRecorder, string) {
	req, _ := http.NewRequest(method, testurl, nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code == 302 {
		location, _ := resp.Result().Location()
		fmt.Printf("=> Redirect to: %s\n", location.String())
		testurl = location.String()
		req, _ = http.NewRequest("GET", testurl, nil)
		resp = httptest.NewRecorder()
		router.ServeHTTP(resp, req)
	}
	return resp, testurl
}

// testLogin
func testLogin(t *testing.T, router *gin.Engine, login string, pass string) *httptest.ResponseRecorder {
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
	// fmt.Printf("%+v\n",resp)
	if login == "" || pass == "" {
		return resp
	}
	assert.Equal(t, 302, resp.Code, "http POST success redirect to Edit")
	newurl, _ := resp.Result().Location()
	fmt.Printf("=> Redirect to: %s\n", newurl.String())
	cookie := resp.Result().Cookies()
	fmt.Printf("=> Cookie: %+v\n", cookie[0])
	if strings.Contains(cookie[0].String(), "session") {
		req, _ = http.NewRequest("GET", newurl.String(), nil)
		for _, c := range cookie {
			req.Header.Add("Cookie", c.String())
		}
		resp = httptest.NewRecorder()
		router.ServeHTTP(resp, req)
	}
	return resp
}

// mock routes function

func setConfigTest(cfg WebConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("Cfg", cfg)
		c.Next()
	}
}

func SetUserTest(login string, loginID string, role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("Login", login)
		c.Set("LoginID", loginID)
		c.Set("Role", role)
		c.Next()
	}
}

func InitRouterTest(cfg WebConfig) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	basePath := cfg.Locale.Path

	r.Use(MiddlewareSession(false))

	var err error
	I18n, err = i18n.New(i18n.Glob(basePath+"/*/*"), cfg.Locale.Langs...)
	if err != nil {
		panic(err)
	}

	translateLangFunc := func(x string) string { return Tr(cfg.Locale.Lang, x) }

	r.SetFuncMap(template.FuncMap{
		"tr": translateLangFunc,
	})
	r.LoadHTMLGlob("../routes/web/templates/**/*.tmpl")

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "home/index.tmpl", gin.H{"appname": cfg.AppName, "appdesc": cfg.AppDesc})
	})
	r.Static("/css", basePath+"/web/assets/css")
	r.Static("/fonts", basePath+"/web/assets/fonts")
	r.Static("/js", basePath+"/web/assets/js")

	r.Use(setConfigTest(cfg))
	return r

}

// some default test values

func initUsersValues() {
	v1 := User{
		Name:         "user1",
		UIDNumber:    5000,
		PrimaryGroup: 6501,
		PassSHA256:   "6478579e37aff45f013e14eeb30b3cc56c72ccdc310123bcdf53e0333e3f416a",
	}
	Data.Users = append(Data.Users, v1)
	v2 := User{
		Name:         "user2",
		UIDNumber:    5001,
		PrimaryGroup: 6504,
		OtherGroups:  []int{6501, 6503},
	}
	Data.Users = append(Data.Users, v2)
	v3 := User{
		Name:         "serviceapp",
		UIDNumber:    5002,
		PrimaryGroup: 6502,
		PassSHA256:   "6478579e37aff45f013e14eeb30b3cc56c72ccdc310123bcdf53e0333e3f416a",
	}
	Data.Users = append(Data.Users, v3)
	g1 := Group{
		Name:      "group1",
		GIDNumber: 6502,
	}
	Data.Groups = append(Data.Groups, g1)
	g2 := Group{
		Name:      "group2",
		GIDNumber: 6503,
	}
	Data.Groups = append(Data.Groups, g2)
}
