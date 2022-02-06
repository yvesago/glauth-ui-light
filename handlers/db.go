package handlers

import (
	"fmt"

	"github.com/gin-gonic/gin"

	. "glauth-ui-light/config"
	. "glauth-ui-light/helpers"
)

func CancelChanges(c *gin.Context) {
	cfg := c.MustGet("Cfg").(WebConfig)
	lang := cfg.Locale.Lang

	if !isAdminAccess(c, "CancelChanges", "-") {
		return
	}

	if Lock != 0 {
		DataRead, _, err := ReadDB(&cfg)
		if err == nil {
			Data = DataRead
			Lock = 0
			SetFlashCookie(c, "success", Tr(lang, "Changes canceled"))
			Log.Info(fmt.Sprintf("%s -- [%s] changes canceled", c.ClientIP(), c.MustGet("Login").(string)))
		} else {
			SetFlashCookie(c, "warning", err.Error())
		}
	} else {
		SetFlashCookie(c, "warning", Tr(lang, "Nothing to cancel"))
	}

	c.Redirect(302, "/auth/crud/user")
}

func SaveChanges(c *gin.Context) {
	cfg := c.MustGet("Cfg").(WebConfig)
	lang := cfg.Locale.Lang

	if !isAdminAccess(c, "SaveChanges", "-") {
		return
	}

	if Lock != 0 {
		username := c.MustGet("Login").(string)
		err := WriteDB(&cfg, Data, username)
		if err == nil {
			Lock = 0
			SetFlashCookie(c, "success", Tr(lang, "Changes saved"))
			Log.Info(fmt.Sprintf("%s -- [%s] changes saved", c.ClientIP(), username))
		} else {
			SetFlashCookie(c, "warning", err.Error())
		}
	} else {
		SetFlashCookie(c, "warning", Tr(lang, "Nothing to save"))
	}

	c.Redirect(302, "/auth/crud/user")
}
