package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	. "glauth-ui-light/config"
	"glauth-ui-light/helpers"
)

var (
	Data    Ctmp
	Lock    int     // number of waiting changes in memory
	Version = "dev" // will be set on build
)

var Log logrus.Logger

func isAdminAccess(c *gin.Context, ressource string, id string) bool {
	login := c.MustGet("Login").(string)
	loginid := c.MustGet("LoginID").(string)
	role := c.MustGet("Role").(string)
	// admin access
	if role != "admin" {
		Log.Info(fmt.Sprintf("-- [%s] (%s) denied admin access to %s : %s", login, loginid, ressource, id))
		c.Redirect(302, "/auth/logout")
		c.Abort()
		return false
	}
	return true
}

func isSelfAccess(c *gin.Context, ressource string, id string) bool {
	login := c.MustGet("Login").(string)
	loginid := c.MustGet("LoginID").(string)
	role := c.MustGet("Role").(string)

	// Self access
	if role != "admin" && loginid != id {
		Log.Info(fmt.Sprintf("-- [%s] (%s) denied self access to %s : %s", login, loginid, ressource, id))
		c.Redirect(302, "/auth/logout")
		c.Abort()
		return false
	}
	return true
}

func render(c *gin.Context, data gin.H, templateName string) {
	// Set user
	role, _ := c.Get("Role")
	data["userName"], data["userId"] = helpers.GetUserID(c)
	if role != nil && role.(string) == "admin" {
		data["roleAdmin"] = true
	}

	// Set CSRF token in forms
	data["Csrf"], _ = c.Get("Csrf")

	// Set view elements
	data["lock"] = Lock
	data["version"] = Version
	data["appname"], _ = c.Get("AppName")
	data["MaskOTP"], _ = c.Get("MaskOTP")

	canChgPass, _ := c.Get("CanChgPass")
	if canChgPass != nil {
		data["canChgPass"] = canChgPass.(bool)
	}

	useOtp, _ := c.Get("UseOtp")
	if useOtp != nil {
		data["useOtp"] = useOtp.(bool)
	}

	data["groupsinfo"] = GetSpecialGroups(c)

	if data["success"] == nil {
		data["success"] = helpers.GetFlashCookie(c, "success")
	}
	if data["warning"] == nil {
		data["warning"] = helpers.GetFlashCookie(c, "warning")
	}
	if data["error"] == nil {
		data["error"] = helpers.GetFlashCookie(c, "error")
	}

	c.HTML(http.StatusOK, templateName, data)

	/*switch c.Request.Header.Get("Accept") {
	case "application/json":
	          // Respond with JSON
	          c.JSON(http.StatusOK, data["payload"])
	  case "application/xml":
	          // Respond with XML
	          c.XML(http.StatusOK, data["payload"])
	default:
		// Respond with HTML
		c.HTML(http.StatusOK, templateName, data)
	}*/
}
