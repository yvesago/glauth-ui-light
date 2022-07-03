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
			Entropy:          60,
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

	for _, s := range []string{"u", "test", "va2ieYeidafee8Gi0"} {
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

	for _, s := range []string{"u", "va2ieYeidafee8Gi0"} {
		tf := UserForm{
			Name:       "test",
			NewPassApp: s,
			Lang:       cfg.Locale.Lang,
		}
		v := tf.Validate(cfg.PassPolicy)
		fmt.Printf(" test NewPassApp «%s» : %s\n", s, tf.Errors["NewPassApp"])
		assert.Equal(t, true, len(tf.Errors["NewPassApp"]) > 0, "set NewPassApp error")
		assert.Equal(t, false, v, "bad NewPassApp form: "+tf.Errors["NewPassApp"])
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
		DefaultHomedir: "/home",
		DefaultLoginShell: "/bin/false",
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

	// test ui for unix fields
	//fmt.Println(resp.Body)
	re = regexp.MustCompile(`/bin/false`)
	matches = re.FindAllStringSubmatch(resp.Body.String(), -1)
	assert.Equal(t, 2, len(matches), "2 /bin/false for user2")
	re = regexp.MustCompile(`/home/user2`)
	matches = re.FindAllStringSubmatch(resp.Body.String(), -1)
	assert.Equal(t, 1, len(matches), "1 homedir result for user2")

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
	form.Add("inputNewPassApp", "pass1")
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
	assert.Equal(t, 1, len(Data.Users[0].PassAppBcrypt), "1 PassApp")

	fmt.Println("= http Update only Password")
	form = url.Values{}
	form.Add("inputName", "user2")                         // Mandatory
	form.Add("inputMail", "test@exemple.com")              // to be set
	form.Add("inputOTPSecret", "gvxdgn3hpfvwu2lhmz3gmm3z") // to be set
	form.Add("inputPassword", "somePass")
	form.Add("inputDelPassApp0", "on")
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
	assert.Equal(t, 0, len(Data.Users[0].PassAppBcrypt), "no more PassApp")

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
