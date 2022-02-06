package helpers

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"

	"glauth-ui-light/config"
)

func GetUserByName(name string) (config.User, error) {
	for k := range Data.Users {
		if Data.Users[k].Name == name {
			return Data.Users[k], nil
		}
	}
	return config.User{}, fmt.Errorf("unknown user")
}

func LoginTestHandlerForm(c *gin.Context) {
	cfg := c.MustGet("Cfg").(config.WebConfig)
	userName, userId := GetUserID(c)

	c.HTML(200, "home/login.tmpl", gin.H{
		"userName":    userName,
		"userId":      userId,
		"currentPage": "login",
		"appname":     cfg.AppName,
		"warning":     GetFlashCookie(c, "warning"),
		"error":       GetFlashCookie(c, "error"),
	})
}

func LoginTestHandler(c *gin.Context) {
	cfg := c.MustGet("Cfg").(config.WebConfig)
	fmt.Println(" - POST /login")
	lang := cfg.Locale.Lang

	s := GetSession(c)
	s = FailLimiter(s, 30) // lock 30s after 4 failed logins

	username := c.PostForm("username")
	password := c.PostForm("password")

	switch {
	case s.Lock: // == true
		s.User = username
		s.UserID = ""
		SetSession(c, s.ToJSONStr())
		fmt.Println(" - Lock Status for ", username)
		SetFlashCookie(c, "error", Tr(lang, "Too many errors, come back later"))
		c.Redirect(302, "/auth/login")
	case username != "" && password != "":
		valid := false
		u, err := GetUserByName(username)
		if err == nil {
			valid = u.ValidPass(password, cfg.PassPolicy.AllowReadSSHA256)
		} else {
			fmt.Println(" - No user ", username)
		}
		/*if *backend == "test" {
			valid = testValidateUser(username, password)
		}
		if *backend == "ldap" {
			valid = ldapValidateUser(username, password, config)
		}*/
		if valid {
			tmpid := strconv.Itoa(u.UIDNumber)
			s.User = username
			s.UserID = tmpid
			s.Count = 0

			SetSession(c, s.ToJSONStr())
			c.Redirect(302, "/user/"+tmpid)
		} else {
			fmt.Println(" - AUTHENTICATION failed for ", username)
			s.User = username
			s.UserID = ""
			SetSession(c, s.ToJSONStr())
			SetFlashCookie(c, "warning", Tr(lang, "Bad credentials"))
			c.Redirect(302, "/auth/login")
		}
	default:
		fmt.Println(" - Bad Post params")
		c.HTML(404, "home/login.tmpl", nil)
	}
}

func LogoutTestHandler(c *gin.Context) {
	cfg := c.MustGet("Cfg").(config.WebConfig)
	lang := cfg.Locale.Lang

	ClearSession(c)
	SetFlashCookie(c, "success", Tr(lang, "You are disconnected"))
	c.Redirect(302, "/")
}
