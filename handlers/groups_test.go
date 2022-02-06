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

func TestGroupValidate(t *testing.T) {
	defer resetData()

	cfg := WebConfig{
		Locale: Locale{
			Lang: "en",
			Path: "../locales/",
		},
	}
	InitRouterTest(cfg)
	initUsersValues()

	for _, s := range []string{"", "u", "va2ieYeidafee8Gi0", "uuu nn", "Aee"} {
		tf := GroupForm{
			Name: s,
			Lang: cfg.Locale.Lang,
		}
		v := tf.Validate()
		fmt.Printf(" test Name «%s» : %s\n", s, tf.Errors["Name"])
		assert.Equal(t, true, len(tf.Errors["Name"]) > 0, "set Name error")
		assert.Equal(t, false, v, "bad Name form: "+tf.Errors["Name"])
	}

	groupf := GroupForm{
		GIDNumber: -1,
		Name:      "éé-- az",
		Lang:      cfg.Locale.Lang,
	}
	v := groupf.Validate()
	assert.Equal(t, false, v, "unvalide group form")
	assert.Equal(t, "Unknown group", groupf.Errors["GIDNumber"], "unvalide group form")
	assert.Equal(t, "Bad character", groupf.Errors["Name"], "unvalide group form")

	groupf = GroupForm{
		Name: "group1",
		Lang: cfg.Locale.Lang,
	}
	v = groupf.Validate()
	assert.Equal(t, false, v, "unvalide group form")
	assert.Equal(t, "Name already used", groupf.Errors["Name"], "unvalide group form")

	groupf = GroupForm{
		Name: "",
		Lang: cfg.Locale.Lang,
	}
	v = groupf.Validate()
	assert.Equal(t, false, v, "unvalide group form")
	assert.Equal(t, "Mandatory", groupf.Errors["Name"], "unvalide group form")

}

func TestGroupHandlers(t *testing.T) {
	defer resetData()

	cfg := WebConfig{
		DBfile: "sample-simple.cfg",
		Locale: Locale{
			Lang: "en",
			Path: "../locales/",
		},
		Debug:   true,
		Verbose: false,
		CfgUsers: CfgUsers{
			Start:         5000,
			GIDAdmin:      6501,
			GIDcanChgPass: 6650,
		},
		CfgGroups: CfgGroups{
			Start: 6500,
		},
	}

	gin.SetMode(gin.TestMode)
	router := InitRouterTest(cfg)

	var Url = "/auth/crud/group"
	router.GET("/auth/login", LoginHandlerForm)
	router.Use(SetUserTest("user1", "5000", "admin"))
	router.GET(Url+"/create", GroupAdd)
	router.POST(Url+"/create", GroupCreate)
	router.GET(Url, GroupList)
	router.GET(Url+"/:id", GroupEdit)
	router.POST(Url+"/del/:id", GroupDel)
	router.POST(Url+"/:id", GroupUpdate)

	//fmt.Printf("%+v\n",Data)

	// Add
	fmt.Println("= http Add Group")
	form := url.Values{}
	form.Add("inputName", "group1")
	req, err := http.NewRequest("POST", Url+"/create", strings.NewReader(form.Encode()))
	req.PostForm = form
	req.Header.Add("Content-Type", "application/x-www-form-Urlencoded")
	if err != nil {
		fmt.Println(err)
	}
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	assert.Equal(t, 302, resp.Code, "http POST success redirect to Edit")
	//fmt.Println(resp.Body)

	// Add second group
	fmt.Println("= http Add more Group")
	form = url.Values{}
	form.Add("inputName", "group2")
	req, err = http.NewRequest("POST", Url+"/create", strings.NewReader(form.Encode()))
	req.PostForm = form
	req.Header.Add("Content-Type", "application/x-www-form-Urlencoded")
	resp = httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	assert.Equal(t, 302, resp.Code, "http POST success redirect to Edit")

	// Get all
	fmt.Println("= http GET all Groups")
	req, err = http.NewRequest("GET", Url, nil)
	if err != nil {
		fmt.Println(err)
	}
	resp = httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	assert.Equal(t, 200, resp.Code, "http success")
	//fmt.Println(resp.Body)
	re := regexp.MustCompile(`href="/auth/crud/group/(\d+)">Edit</a>`)
	matches := re.FindAllStringSubmatch(resp.Body.String(), -1)
	fmt.Printf("===\n%+v\n===\n", matches)
	assert.Equal(t, 2, len(matches), "2 results")

	// Get one
	fmt.Println("= http GET one Group")
	req, err = http.NewRequest("GET", Url+"/6501", nil)
	if err != nil {
		fmt.Println(err)
	}
	resp = httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	assert.Equal(t, 200, resp.Code, "http success")
	//fmt.Println(resp.Body)
	re = regexp.MustCompile(`id="inputName" value="(.*?)" required`)
	matches = re.FindAllStringSubmatch(resp.Body.String(), -1)
	assert.Equal(t, 1, len(matches), "1 result for group")
	fmt.Printf("===\n%+v\n===\n", matches[0][1])
	assert.Equal(t, "group2", matches[0][1], "Name group2")

	// Delete one
	fmt.Println("= http DELETE one Group")
	req, _ = http.NewRequest("POST", Url+"/del/6500", nil)
	resp = httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	//fmt.Println(resp)
	assert.Equal(t, 302, resp.Code, "http Del success, redirect to list")

	req, _ = http.NewRequest("GET", Url, nil)
	resp = httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	assert.Equal(t, 200, resp.Code, "http success")
	re = regexp.MustCompile(`href="/auth/crud/group/(\d+)">Edit</a>`)
	matches = re.FindAllStringSubmatch(resp.Body.String(), -1)
	//fmt.Println(resp.Body)
	fmt.Printf("===\n%+v\n===\n", matches)
	assert.Equal(t, 1, len(matches), "1 result")

	// Update one
	fmt.Println("= http Update one Group")
	form = url.Values{}
	form.Add("inputName", "group2a")
	req, err = http.NewRequest("POST", Url+"/6501", strings.NewReader(form.Encode()))
	req.PostForm = form
	req.Header.Add("Content-Type", "application/x-www-form-Urlencoded")
	resp = httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	assert.Equal(t, 200, resp.Code, "http Update success")
	assert.Equal(t, true, strings.Contains(resp.Body.String(), "group2a"), "group in list")

	// TEST good access
	fmt.Println("= TEST good access")
	respA, resurl := testAccess(t, router, "GET", "/auth/login")
	assert.Equal(t, 200, respA.Code, "http GET login")
	assert.Equal(t, true, strings.Contains(respA.Body.String(), "<h3>Connection</h3>"), "print login template")

	respA, resurl = testAccess(t, router, "GET", Url+"/create")
	assert.Equal(t, 200, respA.Code, "http GET create group")
	assert.Equal(t, true, strings.Contains(respA.Body.String(), "Add group"), "print Add group template")

	// TEST errors
	fmt.Println("= TEST errors")
	respA, resurl = testAccess(t, router, "GET", Url+"/5099")
	assert.Equal(t, 200, respA.Code, "http GET print error unknown group")
	assert.Equal(t, true, strings.Contains(respA.Body.String(), "<H3>Error</H3>"), "print error unknown group")

	respA, resurl = testAccess(t, router, "POST", Url+"/5099")
	assert.Equal(t, 200, respA.Code, "http GET print error unknown group")
	assert.Equal(t, true, strings.Contains(respA.Body.String(), "<H3>Error</H3>"), "print error unknown group")

	respA, resurl = testAccess(t, router, "POST", Url+"/del/5099")
	assert.Equal(t, 200, respA.Code, "http GET print error unknown group")
	assert.Equal(t, true, strings.Contains(respA.Body.String(), "<H3>Error</H3>"), "print error unknown group")

	form = url.Values{}
	form.Add("inputName", "bad name<>")
	req, _ = http.NewRequest("POST", Url+"/create", strings.NewReader(form.Encode()))
	req.PostForm = form
	req.Header.Add("Content-Type", "application/x-www-form-Urlencoded")
	resp = httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	assert.Equal(t, 200, resp.Code, "http POST create invalid redirect to self url")

	form = url.Values{}
	form.Add("inputName", "group2<>")
	req, err = http.NewRequest("POST", Url+"/6501", strings.NewReader(form.Encode()))
	req.PostForm = form
	req.Header.Add("Content-Type", "application/x-www-form-Urlencoded")
	resp = httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	assert.Equal(t, 200, resp.Code, "http Update invalid, redirect to self url")
	assert.Equal(t, "group2a", Data.Groups[0].Name, "updated group2")

	initUsersValues()
	//fmt.Printf("%+v\n",Data)
	respA, resurl = testAccess(t, router, "POST", Url+"/del/6502")
	assert.Equal(t, 200, respA.Code, "http POST reject non empty group, used in PrimaryGroup ")
	assert.Equal(t, "/auth/crud/group", resurl, "http GET redirect to logout")
	assert.Equal(t, true, strings.Contains(respA.Body.String(), "href=\"/auth/crud/group/6502\">Edit</a>"), "group in list")

	respA, resurl = testAccess(t, router, "POST", Url+"/del/6503")
	assert.Equal(t, 200, respA.Code, "http POST reject non empty group, used in OtherGroups")
	assert.Equal(t, "/auth/crud/group", resurl, "http GET redirect to logout")
	assert.Equal(t, true, strings.Contains(respA.Body.String(), "href=\"/auth/crud/group/6503\">Edit</a>"), "group in list")
	//fmt.Println(respA.Body)

	// TEST bad access
	fmt.Println("= TEST bad access")
	Url = "/auth/crud/group"
	r := InitRouterTest(cfg)
	r.GET("/auth/logout", LogoutHandler)
	r.Use(SetUserTest("user1", "5000", "user"))
	r.GET(Url+"/create", GroupAdd)
	r.POST(Url+"/create", GroupCreate)
	r.GET(Url, GroupList)
	r.GET(Url+"/:id", GroupEdit)
	r.POST(Url+"/del/:id", GroupDel)
	r.POST(Url+"/:id", GroupUpdate)

	respA, resurl = testAccess(t, r, "GET", Url)
	assert.Equal(t, 302, respA.Code, "http GET reject non admin access")
	assert.Equal(t, "/auth/logout", resurl, "http GET redirect to logout")

	respA, resurl = testAccess(t, r, "GET", Url+"/create")
	assert.Equal(t, 302, respA.Code, "http GET reject non admin access")
	assert.Equal(t, "/auth/logout", resurl, "http GET redirect to logout")

	respA, resurl = testAccess(t, r, "POST", Url+"/create")
	assert.Equal(t, 302, respA.Code, "http POST reject non admin access")
	assert.Equal(t, "/auth/logout", resurl, "http POST redirect to logout")

	//fmt.Printf("%+v\n",Data)
	respA, resurl = testAccess(t, r, "GET", Url+"/6500")
	assert.Equal(t, 302, respA.Code, "http GET reject non admin access")
	assert.Equal(t, "/auth/logout", resurl, "http GET redirect to logout")

	respA, resurl = testAccess(t, r, "POST", Url+"/6500")
	assert.Equal(t, 302, respA.Code, "http GET reject non admin access")
	assert.Equal(t, "/auth/logout", resurl, "http GET redirect to logout")

	respA, resurl = testAccess(t, r, "POST", Url+"/del/6500")
	assert.Equal(t, 302, respA.Code, "http GET reject non admin access")
	assert.Equal(t, "/auth/logout", resurl, "http GET redirect to logout")

}
