//nolint
package handlers

import (
	//"bytes"
	//"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	. "glauth-ui-light/config"
)

func TestUserValidate(t *testing.T) {
	defer resetData()

	cfg := WebConfig{
		Locale: Locale{
			Lang: "en",
			Path: "../locales/",
		},
		PassPolicy: PassPolicy{
			Min:              2,
			Max:              8,
			AllowReadSSHA256: true,
		},
	}
	InitRouterTest(cfg)
	initUsersValues()

	for _, s := range []string{"", "u", "va2ieYeidafee8Gi0", "uuu nn", "Aee"} {
		tf := UserForm{
			Name: s,
			Lang: cfg.Locale.Lang,
		}
		v := tf.Validate(cfg.PassPolicy)
		fmt.Printf(" test Name «%s» : %s\n", s, tf.Errors["Name"])
		assert.Equal(t, true, len(tf.Errors["Name"]) > 0, "set Name error")
		assert.Equal(t, false, v, "bad Name form: "+tf.Errors["Name"])
	}

	for _, s := range []string{"u", "va2ieYeidafee8Gi0"} {
		tf := UserForm{
			Name:     "test",
			Password: s,
			Lang:     cfg.Locale.Lang,
		}
		v := tf.Validate(cfg.PassPolicy)
		fmt.Printf(" test Password «%s» : %s\n", s, tf.Errors["Password"])
		assert.Equal(t, true, len(tf.Errors["Password"]) > 0, "set Password error")
		assert.Equal(t, false, v, "bad Password form: "+tf.Errors["Password"])
	}

	for _, s := range []string{"va2ieYqsqeii;dafee8Gi0", "uuu nn", "Aee", "4S62BZNFXXSZLCRO4S62BZNFXXSZLCRO4S62BZNFXXSZ"} {
		tf := UserForm{
			Name:      "test",
			OTPSecret: s,
			Lang:      cfg.Locale.Lang,
		}
		v := tf.Validate(cfg.PassPolicy)
		fmt.Printf(" test OTPSecret «%s» : %s\n", s, tf.Errors["OTPSecret"])
		assert.Equal(t, true, len(tf.Errors["OTPSecret"]) > 0, "set OTPSecret error on: "+s)
		assert.Equal(t, false, v, "bad OTPSecret form: "+tf.Errors["OTPSecret"])
	}

	userf := UserForm{
		UIDNumber: -1,
		Mail:      "sssss",
		Name:      "éé-- az",
		Password:  "somePassword",
		//PrimaryGroup: u.PrimaryGroup,
		//OtherGroups:  u.OtherGroups,
		SN:        "bad char <>",
		GivenName: "bad char %",
		Lang:      cfg.Locale.Lang,
	}
	v := userf.Validate(cfg.PassPolicy)
	assert.Equal(t, false, v, "unvalide user form")
	assert.Equal(t, "Unknown user", userf.Errors["UIDNumber"], "unvalide user form")
	assert.Equal(t, "Bad character", userf.Errors["Name"], "unvalide user form")
	assert.Equal(t, "Please enter a valid email address", userf.Errors["Mail"], "unvalide user form")
	assert.Equal(t, "Bad character", userf.Errors["SN"], "unvalide user form")
	assert.Equal(t, "Bad character", userf.Errors["GivenName"], "unvalide user form")

	userf = UserForm{
		Name: "user1",
		Lang: cfg.Locale.Lang,
	}
	v = userf.Validate(cfg.PassPolicy)
	assert.Equal(t, false, v, "unvalide user form")
	assert.Equal(t, "Name already used", userf.Errors["Name"], "unvalide user form")

	userf = UserForm{
		Name:      "",
		SN:        "quooZeiCei3ua9nae4eijoonae0Lahvix",
		GivenName: "quooZeiCei3ua9nae4eijoonae0Lahvix",
		Lang:      cfg.Locale.Lang,
	}
	v = userf.Validate(cfg.PassPolicy)
	assert.Equal(t, false, v, "unvalide user form")
	assert.Equal(t, "Mandatory", userf.Errors["Name"], "unvalide user form")
	assert.Equal(t, "Too long", userf.Errors["SN"], "unvalide user form")
	assert.Equal(t, "Too long", userf.Errors["GivenName"], "unvalide user form")

}

func TestUserHandlers(t *testing.T) {
	defer resetData()

	cfg := WebConfig{
		AppName: "test",
		DBfile:  "sample-simple.cfg",
		Locale: Locale{
			Lang: "en",
			Path: "../locales/",
		},
		Debug: true,
		Tests: true,
		CfgUsers: CfgUsers{
			Start:         5000,
			GIDAdmin:      5501,
			GIDcanChgPass: 5500,
			GIDuseOtp:     5501,
		},
		PassPolicy: PassPolicy{
			Min:              2,
			Max:              8,
			AllowReadSSHA256: true,
		},
	}

	gin.SetMode(gin.TestMode)
	router := InitRouterTest(cfg)

	var Url = "/auth/crud/user"
	router.GET("/login", LoginHandlerForm)
	router.Use(SetUserTest("user1", "5000", "admin"))
	router.GET(Url+"/create", UserAdd)
	router.POST(Url+"/create", UserCreate)
	router.GET(Url, UserList)
	router.GET(Url+"/:id", UserEdit)
	router.POST(Url+"/del/:id", UserDel)
	router.POST(Url+"/:id", UserUpdate)

	//fmt.Printf("%+v\n",Data)

	// Add
	fmt.Println("= http Add User")
	form := url.Values{}
	form.Add("inputName", "user1")
	req, err := http.NewRequest("POST", Url+"/create", strings.NewReader(form.Encode()))
	req.PostForm = form
	req.Header.Add("Content-Type", "application/x-www-form-Urlencoded")
	if err != nil {
		fmt.Println(err)
	}
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	assert.Equal(t, 302, resp.Code, "http POST success redirect to Edit")
	fmt.Println(resp.Body)

	// Add second user
	fmt.Println("= http Add more User")
	form = url.Values{}
	form.Add("inputName", "user2")
	req, err = http.NewRequest("POST", Url+"/create", strings.NewReader(form.Encode()))
	req.PostForm = form
	req.Header.Add("Content-Type", "application/x-www-form-Urlencoded")
	resp = httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	assert.Equal(t, 302, resp.Code, "http POST success redirect to Edit")

	// Get all
	fmt.Println("= http GET all Users")
	req, err = http.NewRequest("GET", Url, nil)
	if err != nil {
		fmt.Println(err)
	}
	resp = httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	assert.Equal(t, 200, resp.Code, "http success")
	//fmt.Println(resp.Body)
	re := regexp.MustCompile(`href="/auth/crud/user/(\d+)">Edit</a>`)
	matches := re.FindAllStringSubmatch(resp.Body.String(), -1)
	fmt.Printf("===\n%+v\n===\n", matches)
	assert.Equal(t, 2, len(matches), "2 results")

	// Get one
	fmt.Println("= http GET one User")
	req, err = http.NewRequest("GET", Url+"/5001", nil)
	if err != nil {
		fmt.Println(err)
	}
	resp = httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	assert.Equal(t, 200, resp.Code, "http success")
	re = regexp.MustCompile(`id="inputName" value="(.*?)" required`)
	matches = re.FindAllStringSubmatch(resp.Body.String(), -1)
	assert.Equal(t, 1, len(matches), "1 result for user")
	fmt.Printf("===\n%+v\n===\n", matches[0][1])
	assert.Equal(t, "user2", matches[0][1], "Name user2")

	// Delete one
	fmt.Println("= http DELETE one User")
	req, _ = http.NewRequest("POST", Url+"/del/5000", nil)
	resp = httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	//fmt.Println(resp)
	assert.Equal(t, 302, resp.Code, "http Del success, redirect to list")

	req, _ = http.NewRequest("GET", Url, nil)
	resp = httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	assert.Equal(t, 200, resp.Code, "http success")
	re = regexp.MustCompile(`href="/auth/crud/user/(\d+)">Edit</a>`)
	matches = re.FindAllStringSubmatch(resp.Body.String(), -1)
	//fmt.Println(resp.Body)
	fmt.Printf("===\n%+v\n===\n", matches)
	assert.Equal(t, 1, len(matches), "1 result")

	// Update one
	// with old sha256 pass
	Data.Users[0].PassSHA256 = "652c7dc687d98c9889304ed2e408c74b611e86a40caa51c4b43f1dd5913c5cd0"
	fmt.Println("= http Update one User")
	form = url.Values{}
	form.Add("inputName", "user2")
	form.Add("inputMail", "test@exemple.com")
	form.Add("inputGroup", "0")
	form.Add("inputOtherGroup", "1")
	form.Add("inputOtherGroup", "2")
	form.Add("inputOTPSecret", "gvxdgn3hpfvwu2lhmz3gmm3z")
	req, err = http.NewRequest("POST", Url+"/5001", strings.NewReader(form.Encode()))
	req.PostForm = form
	req.Header.Add("Content-Type", "application/x-www-form-Urlencoded")
	resp = httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	assert.Equal(t, 200, resp.Code, "http Update success")
	assert.Equal(t, "user2", Data.Users[0].Name, "updated user2")
	assert.Equal(t, "test@exemple.com", Data.Users[0].Mail, "new user2 mail")
	assert.Equal(t, 64, len(Data.Users[0].PassSHA256), "don't  change sha256 pass")
	assert.Equal(t, "", Data.Users[0].PassBcrypt, "no bcrypt, don't change sha256 pass")
	assert.Equal(t, true, strings.Contains(resp.Body.String(), "gvxdgn3hpfvwu2lhmz3gmm3z"), "OTP secret")

	fmt.Println("= http Update only Password")
	form = url.Values{}
	form.Add("inputName", "user2")                         // Mandatory
	form.Add("inputMail", "test@exemple.com")              // to be set
	form.Add("inputOTPSecret", "gvxdgn3hpfvwu2lhmz3gmm3z") // to be set
	form.Add("inputPassword", "somePass")
	req, err = http.NewRequest("POST", Url+"/5001", strings.NewReader(form.Encode()))
	req.PostForm = form
	req.Header.Add("Content-Type", "application/x-www-form-Urlencoded")
	resp = httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	assert.Equal(t, 200, resp.Code, "http Update success")
	assert.Equal(t, "user2", Data.Users[0].Name, "updated user2")
	assert.Equal(t, "test@exemple.com", Data.Users[0].Mail, "new user2 mail")
	assert.Equal(t, "", Data.Users[0].PassSHA256, "no more sha256 pass")
	assert.Equal(t, 120, len(Data.Users[0].PassBcrypt), "bcrypt pass length")

	req, _ = http.NewRequest("GET", Url+"/5001", nil)
	resp = httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	assert.Equal(t, 200, resp.Code, "http success")
	re = regexp.MustCompile(`id="inputMail" value="(.*?)"`)
	matches = re.FindAllStringSubmatch(resp.Body.String(), -1)
	assert.Equal(t, 1, len(matches), "1 result for user")
	//fmt.Println(resp.Body)
	fmt.Printf("===\n%+v\n===\n", matches[0][1])
	assert.Equal(t, "test@exemple.com", matches[0][1], "mail user2")
	assert.Equal(t, true, strings.Contains(resp.Body.String(), "gvxdgn3hpfvwu2lhmz3gmm3z"), "OTP secret")

	// TEST good access
	fmt.Println("= TEST good access")
	respA, resurl := testAccess(t, router, "GET", "/login")
	assert.Equal(t, 200, respA.Code, "http GET login")
	assert.Equal(t, true, strings.Contains(respA.Body.String(), "<h3>Connection</h3>"), "print login template")

	respA, resurl = testAccess(t, router, "GET", Url+"/create")
	assert.Equal(t, 200, respA.Code, "http GET create user")
	assert.Equal(t, true, strings.Contains(respA.Body.String(), "Add user"), "print Add user template")

	// TEST errors
	fmt.Println("= TEST errors")
	respA, resurl = testAccess(t, router, "GET", Url+"/5099")
	assert.Equal(t, 200, respA.Code, "http GET print error unknown user")
	assert.Equal(t, true, strings.Contains(respA.Body.String(), "<H3>Error</H3>"), "print error unknown user")

	respA, resurl = testAccess(t, router, "POST", Url+"/5099")
	assert.Equal(t, 200, respA.Code, "http GET print error unknown user")
	assert.Equal(t, true, strings.Contains(respA.Body.String(), "<H3>Error</H3>"), "print error unknown user")

	respA, resurl = testAccess(t, router, "POST", Url+"/del/5099")
	assert.Equal(t, 200, respA.Code, "http GET print error unknown user")
	assert.Equal(t, true, strings.Contains(respA.Body.String(), "<H3>Error</H3>"), "print error unknown user")

	form = url.Values{}
	form.Add("inputName", "bad login<>")
	req, _ = http.NewRequest("POST", Url+"/create", strings.NewReader(form.Encode()))
	req.PostForm = form
	req.Header.Add("Content-Type", "application/x-www-form-Urlencoded")
	resp = httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	assert.Equal(t, 200, resp.Code, "http POST invalid redirect to self url")

	form = url.Values{}
	form.Add("inputName", "user2")
	form.Add("inputMail", "testBadexemple.com")
	form.Add("inputDisabled", "on")
	form.Add("inputPassword", "s")
	form.Add("inputGroup", "0")
	req, err = http.NewRequest("POST", Url+"/5001", strings.NewReader(form.Encode()))
	req.PostForm = form
	req.Header.Add("Content-Type", "application/x-www-form-Urlencoded")
	resp = httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	assert.Equal(t, 200, resp.Code, "http Update invalid, redirect to self url")
	assert.Equal(t, "user2", Data.Users[0].Name, "updated user2")
	assert.Equal(t, "test@exemple.com", Data.Users[0].Mail, "doni't change user2 mail")
	assert.Equal(t, 120, len(Data.Users[0].PassBcrypt), "bcrypt pass length")

	form = url.Values{}
	form.Add("inputName", "user2")
	form.Add("inputMail", "testBadexemple.com")
	form.Add("inputDisabled", "on")
	form.Add("inputPassword", "s")
	form.Add("inputGroup", "b")
	form.Add("inputOtherGroup", "a")
	form.Add("inputOtherGroup", "2")
	req, err = http.NewRequest("POST", Url+"/5001", strings.NewReader(form.Encode()))
	req.PostForm = form
	req.Header.Add("Content-Type", "application/x-www-form-Urlencoded")
	resp = httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	assert.Equal(t, 200, resp.Code, "http Update invalid, redirect to self url")
	assert.Equal(t, "user2", Data.Users[0].Name, "updated user2")
	assert.Equal(t, "test@exemple.com", Data.Users[0].Mail, "doni't change user2 mail")
	assert.Equal(t, 120, len(Data.Users[0].PassBcrypt), "bcrypt pass length")

	// TEST bad access
	fmt.Println("= TEST bad access")
	Url = "/auth/crud/user"
	r := InitRouterTest(cfg)
	r.GET("/auth/logout", LogoutHandler)
	r.Use(SetUserTest("user1", "5000", "user"))
	r.GET("/auth/user/:id", UserProfile)
	r.GET(Url+"/create", UserAdd)
	r.POST(Url+"/create", UserCreate)
	r.GET(Url, UserList)
	r.GET(Url+"/:id", UserEdit)
	r.POST(Url+"/del/:id", UserDel)
	r.POST(Url+"/:id", UserUpdate)

	respA, resurl = testAccess(t, r, "GET", Url)
	assert.Equal(t, 302, respA.Code, "http GET reject non admin access")
	assert.Equal(t, "/auth/logout", resurl, "http GET redirect to logout")

	respA, resurl = testAccess(t, r, "GET", Url+"/create")
	assert.Equal(t, 302, respA.Code, "http GET reject non admin access")
	assert.Equal(t, "/auth/logout", resurl, "http GET redirect to logout")

	respA, resurl = testAccess(t, r, "POST", Url+"/create")
	assert.Equal(t, 302, respA.Code, "http POST reject non admin access")
	assert.Equal(t, "/auth/logout", resurl, "http POST redirect to logout")

	respA, resurl = testAccess(t, r, "GET", "/auth/user/5001")
	assert.Equal(t, 302, respA.Code, "http GET reject access to other user")
	assert.Equal(t, "/auth/logout", resurl, "http GET redirect to logout")
	//fmt.Printf("%+v\n",Data)
	respA, resurl = testAccess(t, r, "GET", Url+"/5000")
	assert.Equal(t, 302, respA.Code, "http GET reject non admin access")
	assert.Equal(t, "/auth/logout", resurl, "http GET redirect to logout")

	respA, resurl = testAccess(t, r, "GET", "/auth/user/5000")
	assert.Equal(t, 200, respA.Code, "http GET allow access to self profile")
	assert.Equal(t, "/auth/user/5000", resurl, "http GET profile")

	respA, resurl = testAccess(t, r, "POST", Url+"/5000")
	assert.Equal(t, 302, respA.Code, "http GET reject non admin access")
	assert.Equal(t, "/auth/logout", resurl, "http GET redirect to logout")

	respA, resurl = testAccess(t, r, "POST", Url+"/del/5000")
	assert.Equal(t, 302, respA.Code, "http GET reject non admin access")
	assert.Equal(t, "/auth/logout", resurl, "http GET redirect to logout")

}

func TestUserChgPass(t *testing.T) {

	cfg := WebConfig{
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
	//oldpass := Data.Users[0].PassBcrypt
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
	//assert.Equal(t, oldpass, Data.Users[0].PassBcrypt, "pass doesn't change")

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
