//nolint
package handlers

import (
	// "bytes"
	// "encoding/json".
	"fmt"

	"net/http"

	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	. "glauth-ui-light/config"
)

func TestLogin(t *testing.T) {
	// defer deleteFile(config.DBname)
	defer resetData()

	cfg := WebConfig{
		DBfile: "sample-simple.cfg",
		Locale: Locale{
			Lang: "fr",
			Path: "../locales/",
		},
		Debug:   true,
		Tests:   true,
		CfgUsers: CfgUsers{
			Start:         5000,
			GIDAdmin:      6501,
			GIDcanChgPass: 6500,
			GIDuseOtp:     6501,
		},
		PassPolicy: PassPolicy{
			AllowReadSSHA256: true,
		},
	}

	initUsersValues()
	gin.SetMode(gin.TestMode)
	router := InitRouterTest(cfg)

	router.GET("/", func(c *gin.Context) { c.HTML(http.StatusOK, "home/index.tmpl", nil) })

	router.GET("/auth/login", LoginHandlerForm)
	router.POST("/auth/login", LoginHandler)
	router.GET("/auth/logout", LogoutHandler)
	router.Use(SetUserTest("user1", "5000", "admin"))
	router.GET("/auth/user/:id", UserProfile)
	// u.POST("/:id", UserChgPasswd)

	// Login
	fmt.Println("= Login")
	resp, _ := testLogin(t, router, "serviceapp", "dogood", nil) // user without otp
	assert.Equal(t, 200, resp.Code, "http GET success first user profile")
	assert.Equal(t, true, strings.Contains(resp.Body.String(), "class=\"navbar-brand\">serviceapp</span>"), "http GET success first user profile")

	Data.Users[2].Disabled = true // serviceapp user disabled
	resp, _ = testLogin(t, router, "serviceapp", "dogood", nil) // user without otp
	assert.Equal(t, 200, resp.Code, "http GET success first user profile")
	assert.Equal(t, true, strings.Contains(resp.Body.String(), "alert-warning"), "account disabled")
	//fmt.Printf("%+v\n",resp)

	var cookie []*http.Cookie
	resp, cookie = testLogin(t, router, "user1", "dogood", nil) // user with otp
	assert.Equal(t, 200, resp.Code, "http GET success request OTP")
	assert.Equal(t, true, strings.Contains(resp.Body.String(), "id=\"code\""), " waiting totp code")
	fmt.Printf("*** => Cookie: %+v\n", cookie)
	//fmt.Printf("%+v\n",resp)
	resp = testCode(t, router, "123456", cookie) // user with otp
	assert.Equal(t, 302, resp.Code, "not valid code")
	newurl, _ := resp.Result().Location()
	assert.Equal(t, "/auth/login", newurl.String(), "Bad login redirect to /auth/login")

	resp = testCode(t, router, "147756", cookie) // user with otp : hotp test
	assert.Equal(t, 302, resp.Code, "not valid code")
	newurl, _ = resp.Result().Location()
	assert.Equal(t, "/auth/user/5000", newurl.String(), "Bad login redirect to /auth/login")

	resp, _ = testLogin(t, router, "xxuser1", "dogood", nil)
	assert.Equal(t, 200, resp.Code, "http GET success first user profile")
	//newurl, _ := resp.Result().Location()
	//assert.Equal(t, "/auth/login", newurl.String(), "Bad login redirect to /auth/login")

	resp, _ = testLogin(t, router, "user1", "dogoodxxx", nil)
	assert.Equal(t, 200, resp.Code, "http GET success first user profile")
	//newurl, _ = resp.Result().Location()
	//assert.Equal(t, "/auth/login", newurl.String(), "Bad login redirect to /auth/login")

	resp, _ = testLogin(t, router, "", "dogood", nil)
	assert.Equal(t, 404, resp.Code, "Mandatory login and pass return error")
	resp, _ = testLogin(t, router, "user1", "", nil)
	assert.Equal(t, 404, resp.Code, "Mandatory login and pass return error")
	// fmt.Printf("%+v\n", resp)

	badpass := [4]string{"bad1", "bad2", "bad3", "bad4"}
	var c []*http.Cookie
	for i, b := range badpass {
		fmt.Println(i)
		resp, c = testLogin(t, router, "user1", b, c)
		assert.Equal(t, 200, resp.Code, "http GET success first user profile")
		if i >= 3 {
			assert.Equal(t, true, strings.Contains(resp.Body.String(), "retentez plus tard"), "locked")
			//fmt.Printf("%+v\n", resp)
		}
	}
}
