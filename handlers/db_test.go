//nolint
package handlers

import (
	"fmt"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	. "glauth-ui-light/config"
	"glauth-ui-light/helpers"
)

func TestDB(t *testing.T) {
	// defer deleteFile(config.DBname)

	cfg := WebConfig{
		DBfile: "_sample-simple.cfg",
		Locale: Locale{
			Lang: "en",
			Path: "../locales/",
		},
		Debug: true,
		Tests: false,
		CfgUsers: CfgUsers{
			Start:         5000,
			GIDAdmin:      6501,
			GIDcanChgPass: 6500,
		},
		PassPolicy: PassPolicy{
			AllowReadSSHA256: true,
		},
	}
	copyTmpFile(cfg.DBfile+".orig", cfg.DBfile)

	defer clean(cfg.DBfile)

	initUsersValues()
	gin.SetMode(gin.TestMode)
	router := InitRouterTest(cfg)

	Url := "/auth/crud"
	router.Use(SetUserTest("user1", "5000", "admin"))
	router.GET(Url+"/user", UserList)
	router.GET(Url+"/reload", CancelChanges)
	router.GET(Url+"/save", SaveChanges)

	// reload
	fmt.Println("= Reload")
	resp, resurl := testAccessSimple(t, router, "GET", Url+"/reload")
	assert.Equal(t, 200, resp.Code, "http GET reload success")
	assert.Equal(t, Url+"/user", resurl, "http GET reload success")
	assert.Equal(t, true, strings.Contains(resp.Body.String(), "Nothing to cancel"), "Nothing to cancel")

	Lock = 1
	resp, resurl = testAccessSimple(t, router, "GET", Url+"/reload")
	assert.Equal(t, 200, resp.Code, "http GET reload success")
	assert.Equal(t, Url+"/user", resurl, "http GET reload success")
	assert.Equal(t, true, strings.Contains(resp.Body.String(), "Changes canceled"), "db reloaded")
	assert.Equal(t, 0, Lock, "No more Lock")
	assert.Equal(t, 5, len(Data.Users), "5 users in db")
	assert.Equal(t, 3, len(Data.Groups), "3 groups in db")

	resp, resurl = testAccessSimple(t, router, "GET", Url+"/save")
	assert.Equal(t, 200, resp.Code, "http GET reload success")
	assert.Equal(t, Url+"/user", resurl, "http GET reload success")
	assert.Equal(t, true, strings.Contains(resp.Body.String(), "Nothing to save"), "Nothing to save")

	Lock = 1
	resp, resurl = testAccessSimple(t, router, "GET", Url+"/save")
	assert.Equal(t, 200, resp.Code, "http GET reload success")
	assert.Equal(t, Url+"/user", resurl, "http GET reload success")
	assert.Equal(t, true, strings.Contains(resp.Body.String(), "Changes saved"), "Changes saved")
	assert.Equal(t, 0, Lock, "No more Lock")

	_, head, _ := helpers.ReadDB(&cfg)
	assert.Equal(t, true, strings.Contains(head[0], "# Updated by user1 on "), "Changed by saved on first line")

	// Test errors
	cfg.DBfile = "_badfile.cfg"
	rr := InitRouterTest(cfg)
	rr.Use(SetUserTest("user1", "5000", "admin"))
	rr.GET(Url+"/user", UserList)
	rr.GET(Url+"/reload", CancelChanges)
	rr.GET(Url+"/save", SaveChanges)

	Lock = 1
	resp, resurl = testAccessSimple(t, rr, "GET", Url+"/reload")
	assert.Equal(t, 200, resp.Code, "http GET reload success")
	assert.Equal(t, Url+"/user", resurl, "http GET reload failed")
	assert.Equal(t, true, strings.Contains(resp.Body.String(), "Non-existent config path: _badfile.cfg"), "db reloaded")
	assert.Equal(t, 1, Lock, "always Lock")

	r := InitRouterTest(cfg)

	r.GET("/auth/logout", LogoutHandler)
	r.Use(SetUserTest("user1", "5000", "user"))
	r.GET(Url+"/user", UserList)
	r.GET(Url+"/reload", CancelChanges)
	r.GET(Url+"/save", SaveChanges)

	Lock = 1
	resp, resurl = testAccessSimple(t, r, "GET", Url+"/reload")
	assert.Equal(t, 302, resp.Code, "http GET reload failed for user")
	assert.Equal(t, "/auth/logout", resurl, "http GET redirect to logout")

	resp, resurl = testAccessSimple(t, r, "GET", Url+"/save")
	assert.Equal(t, 302, resp.Code, "http GET save failed for user")
	assert.Equal(t, "/auth/logout", resurl, "http GET redirect to logout")
	//fmt.Printf("%+v\n",resp)
}
