package handlers

import (
	"fmt"

	"github.com/gin-gonic/gin"

	. "glauth-ui-light/config"
	. "glauth-ui-light/helpers"
)

// Self user handlers

func UserProfile(c *gin.Context) {
	cfg := c.MustGet("Cfg").(WebConfig)
	lang := cfg.Locale.Lang
	id := c.Params.ByName("id")

	if !isSelfAccess(c, "UserProfile", id) {
		return
	}

	k := ctlUserExist(c, lang, id)
	if k < 0 {
		return
	}

	u := Data.Users[k]
	userf := UserForm{
		UIDNumber:     u.UIDNumber,
		Mail:          u.Mail,
		Name:          u.Name,
		PrimaryGroup:  u.PrimaryGroup,
		OtherGroups:   u.OtherGroups,
		SN:            u.SN,
		GivenName:     u.GivenName,
		Disabled:      u.Disabled,
		OTPSecret:     u.OTPSecret,
		PassAppBcrypt: u.PassAppBcrypt,
		Lang:          lang,
	}

	if userf.OTPSecret != "" {
		userf.CreateOTPimg(cfg.AppName)
	}

	render(c, gin.H{"title": u.Name, "u": userf, "currentPage": "profile", "groupdata": Data.Groups}, "user/profile.tmpl")
}

func UserChgPasswd(c *gin.Context) {
	cfg := c.MustGet("Cfg").(WebConfig)
	lang := cfg.Locale.Lang
	id := c.Params.ByName("id")

	// Ctrl access
	if !isSelfAccess(c, "UserChgPasswd", id) {
		return
	}

	k := ctlUserExist(c, lang, id)
	if k < 0 {
		return
	}

	// Ctrl access with message
	u := Data.Users[k]
	role := c.MustGet("Role").(string)

	userf := &UserForm{
		UIDNumber:     u.UIDNumber,
		Mail:          u.Mail,
		Name:          u.Name,
		PrimaryGroup:  u.PrimaryGroup,
		OtherGroups:   u.OtherGroups,
		SN:            u.SN,
		GivenName:     u.GivenName,
		Disabled:      u.Disabled,
		OTPSecret:     u.OTPSecret,
		PassAppBcrypt: u.PassAppBcrypt,
		Lang:          lang,
	}
	userf.Errors = make(map[string]string)

	if userf.OTPSecret != "" {
		userf.CreateOTPimg(cfg.AppName)
	}

	// application accounts don't change their password
	// users and admins are defined by group set by GIDcanChgPass, GIDAdmin config
	if (role != "admin" && role != "user") || Lock != 0 {
		warning := ""
		if Lock != 0 {
			warning = Tr(lang, "Data locked by admin.")
		}
		render(c, gin.H{
			"title":       u.Name,
			"currentPage": "profile",
			"warning":     warning,
			"u":           userf,
			"groupdata":   Data.Groups},
			"user/profile.tmpl")
		return
	}

	pass1 := c.PostForm("inputPassword")
	pass2 := c.PostForm("inputPassword2")

	// Validate entries
	if pass1 == "" {
		userf.Errors["Password"] = Tr(lang, "Mandatory")
	}
	if pass1 != pass2 {
		userf.Errors["Password2"] = Tr(lang, "Passwords mismatch")
	}
	if pass2 == "" {
		userf.Errors["Password2"] = Tr(lang, "Mandatory")
	}
	if len(userf.Errors) != 0 {
		render(c, gin.H{"title": u.Name, "currentPage": "profile", "u": userf, "groupdata": Data.Groups}, "user/profile.tmpl")
		return
	}
	userf.Password = pass1

	// Validate new password
	if !userf.Validate(cfg.PassPolicy) {
		render(c, gin.H{"title": u.Name, "currentPage": "profile", "u": userf, "groupdata": Data.Groups}, "user/profile.tmpl")
		return
	}

	(&Data.Users[k]).SetBcryptPass(pass1)
	(&Data.Users[k]).PassSHA256 = "" // no more use of SHA256

	username := c.MustGet("Login").(string)
	Log.Info(fmt.Sprintf("%s -- %s password changed by %s", c.ClientIP(), u.Name, username))

	err := WriteDB(&cfg, Data, username)
	if err != nil {
		render(c, gin.H{"title": Tr(lang, "Error"), "currentPage": "profile", "error": err.Error()}, "home/error.tmpl")
		return
	}

	render(c, gin.H{
		"title":       u.Name,
		"currentPage": "profile",
		"success":     Tr(lang, "Password updated"),
		"u":           userf,
		"groupdata":   Data.Groups},
		"user/profile.tmpl")
}

func UserChgOTP(c *gin.Context) {
	cfg := c.MustGet("Cfg").(WebConfig)
	lang := cfg.Locale.Lang
	id := c.Params.ByName("id")

	// Ctrl access
	if !isSelfAccess(c, "UserChgOTP", id) {
		return
	}

	k := ctlUserExist(c, lang, id)
	if k < 0 {
		return
	}

	// Ctrl access with message
	u := Data.Users[k]

	userf := &UserForm{
		UIDNumber:     u.UIDNumber,
		Mail:          u.Mail,
		Name:          u.Name,
		PrimaryGroup:  u.PrimaryGroup,
		OtherGroups:   u.OtherGroups,
		SN:            u.SN,
		GivenName:     u.GivenName,
		Disabled:      u.Disabled,
		OTPSecret:     u.OTPSecret,
		PassAppBcrypt: u.PassAppBcrypt,
		Lang:          lang,
	}
	userf.Errors = make(map[string]string)

	if userf.OTPSecret != "" {
		userf.CreateOTPimg(cfg.AppName)
	}

	groups := u.OtherGroups
	groups = append(groups, u.PrimaryGroup)
	useOtp := contains(groups, cfg.CfgUsers.GIDuseOtp)

	if !useOtp || Lock != 0 { // only for members of GIDuseOtp
		warning := ""
		if Lock != 0 {
			warning = Tr(lang, "Data locked by admin.")
		}
		render(c, gin.H{
			"title":       u.Name,
			"currentPage": "profile",
			"warning":     warning,
			"navotp":      true,
			"u":           userf,
			"groupdata":   Data.Groups},
			"user/profile.tmpl")
		return
	}

	otp := c.PostForm("inputOTPSecret")
	userf.OTPSecret = otp

	// Validate new otpsecret or no change
	if !userf.Validate(cfg.PassPolicy) || otp == (&Data.Users[k]).OTPSecret {
		userf.OTPSecret = (&Data.Users[k]).OTPSecret
		render(c, gin.H{"title": u.Name,
			"currentPage": "profile",
			"navotp":      true,
			"u":           userf,
			"groupdata":   Data.Groups}, "user/profile.tmpl")
		return
	}

	(&Data.Users[k]).OTPSecret = userf.OTPSecret

	username := c.MustGet("Login").(string)
	Log.Info(fmt.Sprintf("%s -- %s otp secret changed by %s", c.ClientIP(), u.Name, username))

	err := WriteDB(&cfg, Data, username)
	if err != nil {
		render(c, gin.H{"title": Tr(lang, "Error"), "currentPage": "profile", "error": err.Error()}, "home/error.tmpl")
		return
	}

	if userf.OTPSecret != "" {
		userf.CreateOTPimg(cfg.AppName)
	}

	render(c, gin.H{
		"title":       u.Name,
		"currentPage": "profile",
		"success":     Tr(lang, "OTP updated"),
		"navotp":      true,
		"u":           userf,
		"groupdata":   Data.Groups},
		"user/profile.tmpl")
}

func UserPassApp(c *gin.Context) {
	cfg := c.MustGet("Cfg").(WebConfig)
	lang := cfg.Locale.Lang
	id := c.Params.ByName("id")

	// Ctrl access
	if !isSelfAccess(c, "UserPassApp", id) {
		return
	}

	k := ctlUserExist(c, lang, id)
	if k < 0 {
		return
	}

	// Ctrl access with message
	u := Data.Users[k]

	userf := &UserForm{
		UIDNumber:     u.UIDNumber,
		Mail:          u.Mail,
		Name:          u.Name,
		PrimaryGroup:  u.PrimaryGroup,
		OtherGroups:   u.OtherGroups,
		SN:            u.SN,
		GivenName:     u.GivenName,
		Disabled:      u.Disabled,
		OTPSecret:     u.OTPSecret,
		PassAppBcrypt: u.PassAppBcrypt,
		Lang:          lang,
	}
	userf.Errors = make(map[string]string)

	if userf.OTPSecret != "" {
		userf.CreateOTPimg(cfg.AppName)
	}

	groups := u.OtherGroups
	groups = append(groups, u.PrimaryGroup)
	useOtp := contains(groups, cfg.CfgUsers.GIDuseOtp)

	if !useOtp || Lock != 0 { // only for members of GIDuseOtp
		warning := ""
		if Lock != 0 {
			warning = Tr(lang, "Data locked by admin.")
		}
		render(c, gin.H{
			"title":       u.Name,
			"currentPage": "profile",
			"warning":     warning,
			"navotp":      true,
			"u":           userf,
			"groupdata":   Data.Groups},
			"user/profile.tmpl")
		return
	}

	// Read input
	username := c.MustGet("Login").(string)

	userf.NewPassApp = c.PostForm("inputNewPassApp")

	change := false
	// Remove pass app
	for d := 0; d < 3; d++ {
		input := fmt.Sprintf("inputDelPassApp%d", d)
		delpass := c.PostForm(input)
		if delpass != "" {
			(&Data.Users[k]).DelPassApp(d)
			change = true
			Log.Info(fmt.Sprintf("%s -- %s passapp removed %d by %s", c.ClientIP(), u.Name, d, username))
		}
	}

	// Validate and register newpass
	if userf.NewPassApp != "" {
		if !userf.Validate(cfg.PassPolicy) {
			render(c, gin.H{"title": u.Name,
				"currentPage": "profile",
				"navotp":      true,
				"u":           userf,
				"groupdata":   Data.Groups}, "user/profile.tmpl")
			return
		}

		(&Data.Users[k]).AddPassApp(userf.NewPassApp)
		change = true
		Log.Info(fmt.Sprintf("%s -- %s passapp added by %s", c.ClientIP(), u.Name, username))
	}

	if change {
		userf.PassAppBcrypt = Data.Users[k].PassAppBcrypt
	}

	err := WriteDB(&cfg, Data, username)
	if err != nil {
		render(c, gin.H{"title": Tr(lang, "Error"), "currentPage": "profile", "error": err.Error()}, "home/error.tmpl")
		return
	}

	render(c, gin.H{
		"title":       u.Name,
		"currentPage": "profile",
		"success":     Tr(lang, "Tokens changed"),
		"navotp":      true,
		"u":           userf,
		"groupdata":   Data.Groups},
		"user/profile.tmpl")
}
