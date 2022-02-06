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
		Verbose: false,
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
	resp := testLogin(t, router, "user1", "dogood")
	assert.Equal(t, 200, resp.Code, "http GET success first user profile")
	// assert.Equal(t, true, strings.Contains(resp.Body.String(), "Welcome user1"), "http GET success first user profile")
	assert.Equal(t, true, strings.Contains(resp.Body.String(), "class=\"navbar-brand\">user1</span>"), "http GET success first user profile")
	// fmt.Printf("%+v\n",resp)
	resp = testLogin(t, router, "xxuser1", "dogood")
	assert.Equal(t, 200, resp.Code, "http GET success first user profile")
	//newurl, _ := resp.Result().Location()
	//assert.Equal(t, "/auth/login", newurl.String(), "Bad login redirect to /auth/login")

	resp = testLogin(t, router, "user1", "dogoodxxx")
	assert.Equal(t, 200, resp.Code, "http GET success first user profile")
	//newurl, _ = resp.Result().Location()
	//assert.Equal(t, "/auth/login", newurl.String(), "Bad login redirect to /auth/login")

	resp = testLogin(t, router, "", "dogood")
	assert.Equal(t, 404, resp.Code, "Mandatory login and pass return error")
	resp = testLogin(t, router, "user1", "")
	assert.Equal(t, 404, resp.Code, "Mandatory login and pass return error")
	// fmt.Printf("%+v\n", resp)
}
