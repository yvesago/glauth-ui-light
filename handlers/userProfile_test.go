//nolint
package handlers

import (
	//"bytes"
	//"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	//"regexp"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	. "glauth-ui-light/config"
)

func TestUserChgPass(t *testing.T) {

	cfg := WebConfig{
		//AppName: "test",
		DBfile: "_sample-simple.cfg",
		Locale: Locale{
			Lang: "fr",
			Path: "../locales/",
		},
		Debug: true,
		Tests: true,
		CfgUsers: CfgUsers{
			Start:         5000,
			GIDAdmin:      6501,
			GIDcanChgPass: 6502,
			GIDuseOtp:     6501,
		},
		PassPolicy: PassPolicy{
			Min:              2,
			Max:              8,
			AllowReadSSHA256: true,
		},
	}
	copyTmpFile(cfg.DBfile+".orig", cfg.DBfile)

	defer clean(cfg.DBfile)

	gin.SetMode(gin.TestMode)
	initUsersValues()

	Lock = 0

	// TEST errors

	u2 := InitRouterTest(cfg)
	u2.Use(SetUserTest("serviceapp", "5000", ""))
	u2.GET("/user/:id", UserProfile)
	u2.POST("/user/:id", UserChgPasswd)

	form2 := url.Values{}
	form2.Add("inputPassword", "pass1")
	form2.Add("inputPassword2", "pass1")
	req2, _ := http.NewRequest("POST", "/user/5001", strings.NewReader(form2.Encode()))
	req2.PostForm = form2
	req2.Header.Add("Content-Type", "application/x-www-form-Urlencoded")
	resp := httptest.NewRecorder()
	u2.ServeHTTP(resp, req2)
	assert.Equal(t, 302, resp.Code, "http POST no change to other profile")

	form2 = url.Values{}
	form2.Add("inputPassword", "pass1")
	form2.Add("inputPassword2", "pass1")
	req2, _ = http.NewRequest("POST", "/user/5000", strings.NewReader(form2.Encode()))
	req2.PostForm = form2
	req2.Header.Add("Content-Type", "application/x-www-form-Urlencoded")
	resp = httptest.NewRecorder()
	u2.ServeHTTP(resp, req2)
	assert.Equal(t, 200, resp.Code, "http POST serviceapp not allowed to change self pass")
	assert.Equal(t, 64, len(Data.Users[0].PassSHA256), "don't change sha256 pass")
	assert.Equal(t, 0, len(Data.Users[0].PassBcrypt), "don't set bcrypt")

	respA2, _ := testAccess(t, u2, "GET", "/user/5000")
	assert.Equal(t, 200, respA2.Code, "http GET allow access to self profile")
	//assert.Equal(t, "/user/5000", resurl2, "http GET profile")
	assert.Equal(t, 200, respA2.Code, "http Update invalid, redirect to self url: /user/5000")
	assert.Equal(t, 64, len(Data.Users[0].PassSHA256), "don't change sha256 pass")
	assert.Equal(t, 0, len(Data.Users[0].PassBcrypt), "don't set bcrypt")

	// Admin access
	u := InitRouterTest(cfg)
	u.Use(SetUserTest("user1", "5000", "admin"))
	u.Use(func(c *gin.Context) {
		c.Set("CanChgPass", true)
		c.Set("UseOtp", true)
		c.Next()
	})
	u.GET("/user/:id", UserProfile)
	u.POST("/user/:id", UserChgPasswd)

	respA, resurl := testAccess(t, u, "GET", "/user/5000")
	assert.Equal(t, 200, respA.Code, "http GET allow access to self profile")
	assert.Equal(t, "/user/5000", resurl, "http GET profile")
	assert.Equal(t, true, strings.Contains(respA.Body.String(), "id=\"nav-otp\""), "show otp nav")
	assert.Equal(t, true, strings.Contains(respA.Body.String(), "OTP"), "show otp img")
	assert.Equal(t, true, strings.Contains(respA.Body.String(), "id=\"nav-chgpwd\""), "show change password nav")
	fmt.Printf("%+v\n", Data.Users[0])

	form2 = url.Values{}
	form2.Add("inputPassword", "pass1")
	form2.Add("inputPassword2", "pass1")
	req2, _ = http.NewRequest("POST", "/user/6000", strings.NewReader(form2.Encode()))
	req2.PostForm = form2
	req2.Header.Add("Content-Type", "application/x-www-form-Urlencoded")
	resp = httptest.NewRecorder()
	u.ServeHTTP(resp, req2)
	assert.Equal(t, 200, resp.Code, "http POST reject access to unknown user")
	assert.Equal(t, true, strings.Contains(resp.Body.String(), "<H3>Erreur</H3>"), "print error unknown user")

	for _, s := range []string{"", "u", "va2ieYeidafee8Gi0", "pass1"} {
		for _, s2 := range []string{"", "u", "va2ieYeidafee8Gi0", "pass2"} {
			form := url.Values{}
			form.Add("inputPassword", s)
			form.Add("inputPassword2", s2)
			req, _ := http.NewRequest("POST", "/user/5000", strings.NewReader(form.Encode()))
			req.PostForm = form
			req.Header.Add("Content-Type", "application/x-www-form-Urlencoded")
			resp := httptest.NewRecorder()
			u.ServeHTTP(resp, req)
			assert.Equal(t, 200, resp.Code, "http Update invalid, redirect to self url: "+s+"/"+s2)
			assert.Equal(t, 64, len(Data.Users[0].PassSHA256), "don't change sha256 pass")
			assert.Equal(t, 0, len(Data.Users[0].PassBcrypt), "don't set bcrypt")
		}
	}

	// Test success
	Lock = 0
	form := url.Values{}
	form.Add("inputPassword", "test")
	form.Add("inputPassword2", "test")
	req, _ := http.NewRequest("POST", "/user/5000", strings.NewReader(form.Encode()))
	req.PostForm = form
	req.Header.Add("Content-Type", "application/x-www-form-Urlencoded")
	resp = httptest.NewRecorder()
	u.ServeHTTP(resp, req)
	assert.Equal(t, 200, resp.Code, "http Update success")
	assert.Equal(t, "user1", Data.Users[0].Name, "updated user1")
	assert.Equal(t, 120, len(Data.Users[0].PassBcrypt), "bcrypt pass length")
	assert.Equal(t, "", Data.Users[0].PassSHA256, "no more sha256")

	// Test error with lock
	Lock = 1
	oldpass := Data.Users[0].PassBcrypt
	form = url.Values{}
	form.Add("inputPassword", "testnew")
	form.Add("inputPassword2", "testnew")
	req, _ = http.NewRequest("POST", "/user/5000", strings.NewReader(form.Encode()))
	req.PostForm = form
	req.Header.Add("Content-Type", "application/x-www-form-Urlencoded")
	resp = httptest.NewRecorder()
	u.ServeHTTP(resp, req)
	assert.Equal(t, 200, resp.Code, "http Update success")
	assert.Equal(t, true, strings.Contains(resp.Body.String(), "Data locked by admin"), "Error message")
	assert.Equal(t, oldpass, Data.Users[0].PassBcrypt, "pass doesn't change")

}

func TestUserChgOTP(t *testing.T) {

	cfg := WebConfig{
		AppName: "test",
		DBfile:  "_sample-simple.cfg",
		Locale: Locale{
			Lang: "ien",
			Path: "../locales/",
		},
		Debug: true,
		Tests: true,
		CfgUsers: CfgUsers{
			Start:         5000,
			GIDAdmin:      6501,
			GIDcanChgPass: 6501,
			GIDuseOtp:     6501,
		},
		PassPolicy: PassPolicy{
			Min:              2,
			Max:              8,
			AllowReadSSHA256: true,
		},
	}
	copyTmpFile(cfg.DBfile+".orig", cfg.DBfile)

	defer clean(cfg.DBfile)

	gin.SetMode(gin.TestMode)
	initUsersValues()

	Lock = 0

	// TEST errors

	u2 := InitRouterTest(cfg)
	u2.Use(SetUserTest("serviceapp", "5002", ""))
	u2.GET("/user/:id", UserProfile)
	u2.POST("/user/otp/:id", UserChgOTP)

	form2 := url.Values{}
	form2.Add("inputOTPSecret", "pass1")
	req2, _ := http.NewRequest("POST", "/user/otp/5001", strings.NewReader(form2.Encode()))
	req2.PostForm = form2
	req2.Header.Add("Content-Type", "application/x-www-form-Urlencoded")
	resp := httptest.NewRecorder()
	u2.ServeHTTP(resp, req2)
	assert.Equal(t, 302, resp.Code, "http POST no change to other profile")

	form2 = url.Values{}
	form2.Add("inputOTPSecret", "3hnvnk4ycv44glzigd6s25j4dougs3rk")
	req2, _ = http.NewRequest("POST", "/user/otp/5002", strings.NewReader(form2.Encode()))
	req2.PostForm = form2
	req2.Header.Add("Content-Type", "application/x-www-form-Urlencoded")
	resp = httptest.NewRecorder()
	u2.ServeHTTP(resp, req2)
	assert.Equal(t, 200, resp.Code, "http POST serviceapp not allowed to change otp")
	assert.Equal(t, 0, len(Data.Users[2].OTPSecret), "don't change secret")
	//fmt.Printf("%+v\n",Data.Users[2])
	//fmt.Printf("%+v\n",resp)

	respA2, _ := testAccess(t, u2, "GET", "/user/5002")
	assert.Equal(t, 200, respA2.Code, "http GET allow access to self profile")
	//assert.Equal(t, "/user/5000", resurl2, "http GET profile")
	assert.Equal(t, 200, respA2.Code, "http Update invalid, redirect to self url: /user/5000")
	assert.Equal(t, 64, len(Data.Users[0].PassSHA256), "don't change sha256 pass")
	assert.Equal(t, 0, len(Data.Users[0].PassBcrypt), "don't set bcrypt")

	// Admin access
	u := InitRouterTest(cfg)
	u.Use(SetUserTest("user1", "5000", "admin"))
	u.Use(func(c *gin.Context) {
		c.Set("CanChgPass", true)
		c.Set("UseOtp", true)
		c.Next()
	})
	u.GET("/user/:id", UserProfile)
	u.POST("/user/otp/:id", UserChgOTP)

	respA, resurl := testAccess(t, u, "GET", "/user/5000")
	assert.Equal(t, 200, respA.Code, "http GET allow access to self profile")
	assert.Equal(t, "/user/5000", resurl, "http GET profile")
	assert.Equal(t, true, strings.Contains(respA.Body.String(), "id=\"nav-otp\""), "show otp nav")
	assert.Equal(t, true, strings.Contains(respA.Body.String(), "OTP"), "show otp img")
	assert.Equal(t, true, strings.Contains(respA.Body.String(), "id=\"nav-chgpwd\""), "show change password nav")
	//fmt.Printf("%+v\n", Data.Users[0])

	form2 = url.Values{}
	form2.Add("inputOTPSecret", "pass1")
	req2, _ = http.NewRequest("POST", "/user/otp/6000", strings.NewReader(form2.Encode()))
	req2.PostForm = form2
	req2.Header.Add("Content-Type", "application/x-www-form-Urlencoded")
	resp = httptest.NewRecorder()
	u.ServeHTTP(resp, req2)
	assert.Equal(t, 200, resp.Code, "http POST reject access to unknown user")
	assert.Equal(t, true, strings.Contains(resp.Body.String(), "<H3>Error</H3>"), "print error unknown user")

	Data.Users[0].OTPSecret = ""
	for _, s := range []string{"va2ieYqsqeii;dafee8Gi0", "uuu nn", "Aee", "4S62BZNFXXSZLCRO4S62BZNFXXSZLCRO4S62BZNFXXSZ"} {
		form := url.Values{}
		form.Add("inputOTPSecret", s)
		req, _ := http.NewRequest("POST", "/user/otp/5000", strings.NewReader(form.Encode()))
		req.PostForm = form
		req.Header.Add("Content-Type", "application/x-www-form-Urlencoded")
		resp := httptest.NewRecorder()
		u.ServeHTTP(resp, req)
		assert.Equal(t, 200, resp.Code, "http Update invalid, redirect to self url: "+s)
		assert.Equal(t, 0, len(Data.Users[0].OTPSecret), "don't set otp secret")
	}

	// Test success
	//fmt.Printf("%+v\n",Data.Users[0])
	Lock = 0
	form := url.Values{}
	form.Add("inputOTPSecret", "3hnvnk4ycv44glzigd6s25j4dougs3rk")
	req, _ := http.NewRequest("POST", "/user/otp/5000", strings.NewReader(form.Encode()))
	req.PostForm = form
	req.Header.Add("Content-Type", "application/x-www-form-Urlencoded")
	resp = httptest.NewRecorder()
	u.ServeHTTP(resp, req)
	assert.Equal(t, 200, resp.Code, "http Update success")
	assert.Equal(t, "user1", Data.Users[0].Name, "updated user1")
	//fmt.Printf("%s\n",Data.Users[0].OTPSecret)
	assert.Equal(t, 32, len(Data.Users[0].OTPSecret), "otp secret pass length")

	// Test error with lock
	Lock = 1
	form = url.Values{}
	form.Add("inputOTPSecret", "")
	req, _ = http.NewRequest("POST", "/user/otp/5000", strings.NewReader(form.Encode()))
	req.PostForm = form
	req.Header.Add("Content-Type", "application/x-www-form-Urlencoded")
	resp = httptest.NewRecorder()
	u.ServeHTTP(resp, req)
	assert.Equal(t, 200, resp.Code, "http Update success")
	assert.Equal(t, true, strings.Contains(resp.Body.String(), "Data locked by admin"), "Error message")
	//assert.Equal(t, oldpass, Data.Users[0].PassBcrypt, "pass doesn't change")

}

func TestUserPassApp(t *testing.T) {

	cfg := WebConfig{
		AppName: "test",
		DBfile:  "_sample-simple.cfg",
		Locale: Locale{
			Lang: "ien",
			Path: "../locales/",
		},
		Debug: true,
		Tests: true,
		CfgUsers: CfgUsers{
			Start:         5000,
			GIDAdmin:      6501,
			GIDcanChgPass: 6501,
			GIDuseOtp:     6501,
		},
		PassPolicy: PassPolicy{
			Min:              2,
			Max:              8,
			AllowReadSSHA256: true,
		},
	}
	copyTmpFile(cfg.DBfile+".orig", cfg.DBfile)

	defer clean(cfg.DBfile)

	gin.SetMode(gin.TestMode)
	initUsersValues()

	// TEST errors

	u2 := InitRouterTest(cfg)
	u2.Use(SetUserTest("serviceapp", "5002", ""))
	u2.GET("/user/:id", UserProfile)
	u2.POST("/user/passapp/:id", UserPassApp)

	form2 := url.Values{}
	form2.Add("inputNewPassApp", "pass1")
	req2, _ := http.NewRequest("POST", "/user/passapp/5001", strings.NewReader(form2.Encode()))
	req2.PostForm = form2
	req2.Header.Add("Content-Type", "application/x-www-form-Urlencoded")
	resp := httptest.NewRecorder()
	u2.ServeHTTP(resp, req2)
	assert.Equal(t, 302, resp.Code, "http POST no change to other profile")

	form2 = url.Values{}
	form2.Add("inputNewPassApp", "3hnvnk4ycv44glzigd6s25j4dougs3rk")
	req2, _ = http.NewRequest("POST", "/user/passapp/5002", strings.NewReader(form2.Encode()))
	req2.PostForm = form2
	req2.Header.Add("Content-Type", "application/x-www-form-Urlencoded")
	resp = httptest.NewRecorder()
	u2.ServeHTTP(resp, req2)
	assert.Equal(t, 200, resp.Code, "http POST serviceapp not allowed to change otp")
	assert.Equal(t, 0, len(Data.Users[2].PassAppBcrypt), "don't add passapp")
	//fmt.Printf("%+v\n",Data.Users[2])
	//fmt.Printf("%+v\n",resp)

	respA2, _ := testAccess(t, u2, "GET", "/user/5002")
	assert.Equal(t, 200, respA2.Code, "http GET allow access to self profile")
	//assert.Equal(t, "/user/5000", resurl2, "http GET profile")
	assert.Equal(t, 200, respA2.Code, "http Update invalid, redirect to self url: /user/5000")
	assert.Equal(t, 64, len(Data.Users[0].PassSHA256), "don't change sha256 pass")
	assert.Equal(t, 0, len(Data.Users[0].PassBcrypt), "don't set bcrypt")

	// Admin access
	u := InitRouterTest(cfg)
	u.Use(SetUserTest("user1", "5000", "admin"))
	u.Use(func(c *gin.Context) {
		c.Set("CanChgPass", true)
		c.Set("UseOtp", true)
		c.Next()
	})
	u.GET("/user/:id", UserProfile)
	u.POST("/user/passapp/:id", UserPassApp)

	respA, resurl := testAccess(t, u, "GET", "/user/5000")
	assert.Equal(t, 200, respA.Code, "http GET allow access to self profile")
	assert.Equal(t, "/user/5000", resurl, "http GET profile")
	assert.Equal(t, true, strings.Contains(respA.Body.String(), "id=\"nav-otp\""), "show otp nav")
	assert.Equal(t, true, strings.Contains(respA.Body.String(), "OTP"), "show otp img")
	assert.Equal(t, true, strings.Contains(respA.Body.String(), "id=\"nav-chgpwd\""), "show change password nav")
	//fmt.Printf("%+v\n", Data.Users[0])

	form2 = url.Values{}
	form2.Add("inputNewPassApp", "pass1")
	req2, _ = http.NewRequest("POST", "/user/passapp/6000", strings.NewReader(form2.Encode()))
	req2.PostForm = form2
	req2.Header.Add("Content-Type", "application/x-www-form-Urlencoded")
	resp = httptest.NewRecorder()
	u.ServeHTTP(resp, req2)
	assert.Equal(t, 200, resp.Code, "http POST reject access to unknown user")
	assert.Equal(t, true, strings.Contains(resp.Body.String(), "<H3>Error</H3>"), "print error unknown user")

	Data.Users[0].OTPSecret = "3hnvnk4ycv44glzigd6s25j4dougs3rk"
	for _, s := range []string{"u", "4S62BZNFXXSZLCRO4S62BZNFXXSZLCRO4S62BZNFXXSZ"} {
		form := url.Values{}
		form.Add("inputNewPassApp", s)
		req, _ := http.NewRequest("POST", "/user/passapp/5000", strings.NewReader(form.Encode()))
		req.PostForm = form
		req.Header.Add("Content-Type", "application/x-www-form-Urlencoded")
		resp := httptest.NewRecorder()
		u.ServeHTTP(resp, req)
		assert.Equal(t, 200, resp.Code, "http Update invalid, redirect to self url: "+s)
		assert.Equal(t, 0, len(Data.Users[0].PassAppBcrypt), "don't set pass app")
		//assert.Equal(t, true, strings.Contains(resp.Body.String(), "Too "), "print error msg")
	}
	//fmt.Printf("%+v\n",Data.Users[0])

	// Test success
	Lock = 0
	form := url.Values{}
	form.Add("inputNewPassApp", "passapp")
	req, _ := http.NewRequest("POST", "/user/passapp/5000", strings.NewReader(form.Encode()))
	req.PostForm = form
	req.Header.Add("Content-Type", "application/x-www-form-Urlencoded")
	resp = httptest.NewRecorder()
	u.ServeHTTP(resp, req)
	assert.Equal(t, 200, resp.Code, "http Update success")
	assert.Equal(t, "user1", Data.Users[0].Name, "updated user1")
	//fmt.Printf("%s\n",Data.Users[0].OTPSecret)
	assert.Equal(t, 1, len(Data.Users[0].PassAppBcrypt), "pass app length")

	// Test error with lock
	Lock = 1
	form = url.Values{}
	form.Add("inputNewPassApp", "sss")
	req, _ = http.NewRequest("POST", "/user/passapp/5000", strings.NewReader(form.Encode()))
	req.PostForm = form
	req.Header.Add("Content-Type", "application/x-www-form-Urlencoded")
	resp = httptest.NewRecorder()
	u.ServeHTTP(resp, req)
	assert.Equal(t, 200, resp.Code, "http Update success")
	assert.Equal(t, true, strings.Contains(resp.Body.String(), "Data locked by admin"), "Error message")
	assert.Equal(t, 1, len(Data.Users[0].PassAppBcrypt), "pass app length")

	// Test del
	Lock = 0
	form = url.Values{}
	form.Add("inputDelPassApp0", "on")
	req, _ = http.NewRequest("POST", "/user/passapp/5000", strings.NewReader(form.Encode()))
	req.PostForm = form
	req.Header.Add("Content-Type", "application/x-www-form-Urlencoded")
	resp = httptest.NewRecorder()
	u.ServeHTTP(resp, req)
	assert.Equal(t, 200, resp.Code, "http Update success")
	assert.Equal(t, "user1", Data.Users[0].Name, "updated user1")
	//fmt.Printf("%s\n",Data.Users[0].OTPSecret)
	assert.Equal(t, 0, len(Data.Users[0].PassAppBcrypt), "no more pass app")

	form = url.Values{}
	form.Add("inputNewPassApp", "s")
	req, _ = http.NewRequest("POST", "/user/passapp/5000", strings.NewReader(form.Encode()))
	req.PostForm = form
	req.Header.Add("Content-Type", "application/x-www-form-Urlencoded")
	resp = httptest.NewRecorder()
	u.ServeHTTP(resp, req)
	assert.Equal(t, 200, resp.Code, "http Update success")
	assert.Equal(t, true, strings.Contains(resp.Body.String(), "Too "), "Error message")
}
