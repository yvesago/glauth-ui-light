package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"glauth-ui-light/config"
	"glauth-ui-light/helpers"
)

func LoginHandlerForm(c *gin.Context) {
	cfg := c.MustGet("Cfg").(config.WebConfig)
	userName, userId := helpers.GetUserID(c)
	s := helpers.GetSession(c)

	c.HTML(200, "home/login.tmpl", gin.H{
		"userName":    userName,
		"userId":      userId,
		"currentPage": "login",
		"version":     Version,
		"appname":     cfg.AppName,
		"otp":         s.ReqOTP,
		"warning":     helpers.GetFlashCookie(c, "warning"),
		"error":       helpers.GetFlashCookie(c, "error"),
	})
}

func LoginHandler(c *gin.Context) {
	cfg := c.MustGet("Cfg").(config.WebConfig)
	Log.Debug(c.ClientIP(), " - POST /login")
	lang := cfg.Locale.Lang

	s := helpers.GetSession(c)
	s = helpers.FailLimiter(s, 30) // lock 30s after 4 failed logins

	username := c.PostForm("username")
	password := c.PostForm("password")
	code := c.PostForm("code")

	switch {
	case s.Lock: // == true
		s.User = username
		s.UserID = ""
		helpers.SetSession(c, s.ToJSONStr())
		Log.Debug(c.ClientIP(), " - Lock Status for ", username)
		helpers.SetFlashCookie(c, "error", helpers.Tr(lang, "Too many errors, come back later"))
		c.Redirect(302, "/auth/login")
	case username != "" && password != "":
		valid := false
		u, err := GetUserByName(username)
		if err == nil {
			valid = u.ValidPass(password, cfg.PassPolicy.AllowReadSSHA256)
		} else {
			Log.Info(c.ClientIP(), " - No user ", username)
		}
		/*if *backend == "test" {
			valid = testValidateUser(username, password)
		}
		if *backend == "ldap" {
			valid = ldapValidateUser(username, password, config)
		}*/
		if valid && !u.Disabled {
			tmpid := strconv.Itoa(u.UIDNumber)
			s.UserID = tmpid
			s.Count = 0

			groups := u.OtherGroups
			groups = append(groups, u.PrimaryGroup)
			useOtp := contains(groups, cfg.CfgUsers.GIDuseOtp)

			// redirect to otp if otp group and secret
			if u.OTPSecret != "" && useOtp {
				s.ReqOTP = true
				s.User = ""
				helpers.SetSession(c, s.ToJSONStr())
				c.Redirect(302, "/auth/login")
			} else { // Auth success
				s.ReqOTP = false
				s.User = username
				helpers.SetSession(c, s.ToJSONStr())
				c.Redirect(302, "/auth/user/"+tmpid)
			}
		} else { // Auth failed
			s.User = ""
			s.UserID = ""
			helpers.SetSession(c, s.ToJSONStr())
			if u.Disabled {
				Log.Info(c.ClientIP(), " - AUTH failed for ", username, " : Account disabled")
				helpers.SetFlashCookie(c, "warning", helpers.Tr(lang, "Account disabled"))
			} else {
				Log.Info(c.ClientIP(), " - AUTH failed for ", username, "Bad credentials")
				helpers.SetFlashCookie(c, "warning", helpers.Tr(lang, "Bad credentials"))
			}
			c.Redirect(302, "/auth/login")
		}
	case s.UserID != "" && code != "":
		u := Data.Users[GetUserKey(s.UserID)]
		valid := u.ValidOTP(code, !cfg.Tests)
		if !valid {
			c.Redirect(302, "/auth/login")
		} else { // Auth success
			s.ReqOTP = false
			s.User = u.Name
			helpers.SetSession(c, s.ToJSONStr())
			tmpid := strconv.Itoa(u.UIDNumber)
			c.Redirect(302, "/auth/user/"+tmpid)
		}
	default:
		Log.Error(c.ClientIP(), " - Bad Post params")
		c.HTML(404, "home/login.tmpl", nil)
	}
}

func LogoutHandler(c *gin.Context) {
	cfg := c.MustGet("Cfg").(config.WebConfig)
	lang := cfg.Locale.Lang

	helpers.ClearSession(c)
	helpers.SetFlashCookie(c, "success", helpers.Tr(lang, "You are disconnected"))
	c.Redirect(302, "/")
}
