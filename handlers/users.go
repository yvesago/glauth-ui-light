package handlers

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"encoding/base32"
	"encoding/base64"
	"image/png"

	"github.com/pquerna/otp"

	. "glauth-ui-light/config"
	. "glauth-ui-light/helpers"
)

// Validate entries

var rxEmail = regexp.MustCompile(".+@.+\\..+") //nolint
var rxName = regexp.MustCompile("^[a-z0-9]+$")

var rxBadChar = regexp.MustCompile("[<>&*%$'«».,;:!` ]+")

type UserForm struct {
	UIDNumber     int
	Name          string
	Mail          string
	SN            string
	GivenName     string
	Password      string
	OTPSecret     string
	OTPImg        string
	PassAppBcrypt []string
	NewPassApp    string
	PrimaryGroup  int
	OtherGroups   []int
	Disabled      bool
	Errors        map[string]string
	Lang          string
}

func (userf *UserForm) CreateOTPimg(appname string) {
	url := fmt.Sprintf("otpauth://totp/%s%%3A%s?secret=%s&issuer=%s", appname, userf.Name, userf.OTPSecret, appname)
	if appname == "" {
		fmt.Println("totp.Generate: Mandatory AppName")
		return
	}
	key, _ := otp.NewKeyFromURL(url)
	var buf bytes.Buffer
	img, _ := key.Image(200, 200)
	e := png.Encode(&buf, img)
	if e != nil {
		fmt.Println("png.Encode: " + e.Error())
		return
	}
	userf.OTPImg = base64.StdEncoding.EncodeToString(buf.Bytes())
}

func (userf *UserForm) Validate(cfg PassPolicy) bool {
	lang := userf.Lang
	userf.Errors = make(map[string]string)

	match := rxEmail.MatchString(userf.Mail)
	if userf.Mail != "" && !match {
		userf.Errors["Mail"] = Tr(lang, "Please enter a valid email address")
	}

	p := userf.Password
	if p != "" {
		switch {
		case len(p) < cfg.Min:
			userf.Errors["Password"] = Tr(lang, "Too short")
		case len(p) > cfg.Max:
			userf.Errors["Password"] = Tr(lang, "Too long")
		}
	}

	np := userf.NewPassApp
	if np != "" {
		switch {
		case len(np) < cfg.Min:
			userf.Errors["NewPassApp"] = Tr(lang, "Too short")
		case len(np) > cfg.Max:
			userf.Errors["NewPassApp"] = Tr(lang, "Too long")
		}
	}

	o := userf.OTPSecret
	if o != "" {
		_, err := base32.StdEncoding.DecodeString(strings.ToUpper(o))
		switch {
		case len(o) < 16:
			userf.Errors["OTPSecret"] = Tr(lang, "Too short")
		case len(o) > 33:
			userf.Errors["OTPSecret"] = Tr(lang, "Too long")
		case err != nil:
			userf.Errors["OTPSecret"] = Tr(lang, "Wrong base32")
		}
	}

	n := userf.Name
	matchName := rxName.MatchString(n)
	switch {
	case strings.TrimSpace(n) == "":
		userf.Errors["Name"] = Tr(lang, "Mandatory")
	case len(n) < 2:
		userf.Errors["Name"] = Tr(lang, "Too short")
	case len(n) > 16:
		userf.Errors["Name"] = Tr(lang, "Too long")
	case !matchName:
		userf.Errors["Name"] = Tr(lang, "Bad character")
	}
	for k := range Data.Users {
		if Data.Users[k].Name == n && Data.Users[k].UIDNumber != userf.UIDNumber {
			userf.Errors["Name"] = Tr(lang, "Name already used")
			break
		}
	}

	matchBadSN := rxBadChar.MatchString(userf.SN)
	if userf.SN != "" && len(userf.SN) > 32 {
		userf.Errors["SN"] = Tr(lang, "Too long")
	}
	if userf.SN != "" && matchBadSN {
		userf.Errors["SN"] = Tr(lang, "Bad character")
	}

	matchBadGname := rxBadChar.MatchString(userf.GivenName)
	if userf.GivenName != "" && len(userf.GivenName) > 32 {
		userf.Errors["GivenName"] = Tr(lang, "Too long")
	}
	if userf.GivenName != "" && matchBadGname {
		userf.Errors["GivenName"] = Tr(lang, "Bad character")
	}

	if userf.UIDNumber < 0 {
		userf.Errors["UIDNumber"] = Tr(lang, "Unknown user")
	}

	return len(userf.Errors) == 0
}

// Helpers

func ctlUserExist(c *gin.Context, lang string, id string) int {
	k := GetUserKey(id)
	if k < 0 {
		render(c, gin.H{"title": Tr(lang, "Error"), "currentPage": "user", "error": Tr(lang, "Unknown user")}, "home/error.tmpl")
		return -1
	}
	return k
}

func GetUserKey(id string) int {
	i := -1
	intId, _ := strconv.Atoi(id)
	for k := range Data.Users {
		if Data.Users[k].UIDNumber == intId {
			i = k
			break
		}
	}
	return i
}

func GetUserByName(name string) (User, error) {
	for k := range Data.Users {
		if Data.Users[k].Name == name {
			return Data.Users[k], nil
		}
	}
	return User{}, fmt.Errorf("unknown user")
}

// Handlers

func UserList(c *gin.Context) {
	cfg := c.MustGet("Cfg").(WebConfig)
	lang := cfg.Locale.Lang

	if !isAdminAccess(c, "UserList", "-") {
		return
	}

	hg := make(map[int]string)
	for k := range Data.Groups {
		hg[Data.Groups[k].GIDNumber] = Data.Groups[k].Name
	}
	render(c, gin.H{"title": Tr(lang, "Users page"), "currentPage": "user", "userdata": Data.Users, "hashgroups": hg}, "user/list.tmpl")
}

func UserEdit(c *gin.Context) {
	cfg := c.MustGet("Cfg").(WebConfig)
	lang := cfg.Locale.Lang
	id := c.Params.ByName("id")

	if !isAdminAccess(c, "UserEdit", id) {
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

	render(c, gin.H{"title": Tr(lang, "Edit user"), "currentPage": "user", "u": userf, "groupdata": Data.Groups}, "user/edit.tmpl")
}

func UserUpdate(c *gin.Context) {
	cfg := c.MustGet("Cfg").(WebConfig)
	lang := cfg.Locale.Lang
	id := c.Params.ByName("id")

	if !isAdminAccess(c, "UserUpdate", id) {
		return
	}

	k := ctlUserExist(c, lang, id)
	if k < 0 {
		return
	}

	// Convert string to right format
	var err error
	var pg int

	if c.PostForm("inputGroup") != "" {
		pg, err = strconv.Atoi(c.PostForm("inputGroup"))
	}
	ogStr := c.PostFormArray("inputOtherGroup")
	d := false
	if c.PostForm("inputDisabled") == "on" {
		d = true
	}
	og := []int{}
	for k := range ogStr {
		i, e := strconv.Atoi(ogStr[k])
		if e != nil {
			err = e
		}
		og = append(og, i)
	}
	if err != nil {
		render(c, gin.H{"title": Tr(lang, "Error"), "currentPage": "user", "error": err.Error()}, "home/error.tmpl")
		return
	}

	// Bind form to struct
	userf := &UserForm{
		UIDNumber:    Data.Users[k].UIDNumber,
		Mail:         c.PostForm("inputMail"),
		Name:         c.PostForm("inputName"),
		SN:           c.PostForm("inputSN"),
		GivenName:    c.PostForm("inputGivenName"),
		Password:     c.PostForm("inputPassword"),
		OTPSecret:    c.PostForm("inputOTPSecret"),
		PrimaryGroup: pg,
		OtherGroups:  og,
		Disabled:     d,
		Lang:         lang,
	}
	// fmt.Printf("%+v\n", userf)
	if userf.OTPSecret != "" {
		userf.CreateOTPimg(cfg.AppName)
	}

	// Validate entries
	if !userf.Validate(cfg.PassPolicy) {
		render(c, gin.H{"title": Tr(lang, "Edit user"), "currentPage": "user", "u": userf, "groupdata": Data.Groups}, "user/edit.tmpl")
		return
	}

	// Update Data
	// updateUser := &Data.Users[k]
	(&Data.Users[k]).Name = userf.Name
	(&Data.Users[k]).PrimaryGroup = userf.PrimaryGroup
	(&Data.Users[k]).OtherGroups = og
	(&Data.Users[k]).SN = userf.SN
	(&Data.Users[k]).GivenName = userf.GivenName
	(&Data.Users[k]).Mail = userf.Mail
	(&Data.Users[k]).Disabled = d
	(&Data.Users[k]).OTPSecret = userf.OTPSecret
	if userf.Password != "" { // optional set password
		(&Data.Users[k]).PassSHA256 = "" // no more use of SHA256
		(&Data.Users[k]).SetBcryptPass(userf.Password)
	}

	Lock++

	Log.Info(fmt.Sprintf("%s -- %s updated by %s", c.ClientIP(), userf.Name, c.MustGet("Login").(string)))

	render(c, gin.H{
		"title":       Tr(lang, "Edit user"),
		"currentPage": "user",
		"success":     "«" + userf.Name + "» updated",
		"u":           userf,
		"groupdata":   Data.Groups},
		"user/edit.tmpl")
}

func UserAdd(c *gin.Context) {
	cfg := c.MustGet("Cfg").(WebConfig)
	lang := cfg.Locale.Lang

	if !isAdminAccess(c, "UserAdd", "-") {
		return
	}

	render(c, gin.H{"title": Tr(lang, "Add user"), "currentPage": "user"}, "user/create.tmpl")
}

func UserCreate(c *gin.Context) {
	cfg := c.MustGet("Cfg").(WebConfig)
	lang := cfg.Locale.Lang

	if !isAdminAccess(c, "UserCreate", "-") {
		return
	}

	// Bind form to struct
	userf := &UserForm{
		Name: c.PostForm("inputName"),
		Lang: lang,
	}
	// Validate entries
	if !userf.Validate(cfg.PassPolicy) {
		render(c, gin.H{"title": Tr(lang, "Add user"), "currentPage": "user", "u": userf, "groupdata": Data.Groups}, "user/create.tmpl")
		return
	}

	// Create new id
	nextID := cfg.CfgUsers.Start - 1 // start uidnumber via config
	for k := range Data.Users {
		if Data.Users[k].UIDNumber >= nextID {
			nextID = Data.Users[k].UIDNumber
		}
	}
	userf.UIDNumber = nextID + 1
	// Add User to Data
	newUser := User{UIDNumber: userf.UIDNumber, Name: userf.Name}
	Data.Users = append(Data.Users, newUser)

	Lock++

	Log.Info(fmt.Sprintf("%s -- %s created by %s", c.ClientIP(), newUser.Name, c.MustGet("Login").(string)))

	SetFlashCookie(c, "success", "«"+newUser.Name+"» added")
	c.Redirect(302, fmt.Sprintf("/auth/crud/user/%d", newUser.UIDNumber))
}

func UserDel(c *gin.Context) {
	cfg := c.MustGet("Cfg").(WebConfig)
	lang := cfg.Locale.Lang
	id := c.Params.ByName("id")

	if !isAdminAccess(c, "UserDel", id) {
		return
	}

	k := ctlUserExist(c, lang, id)
	if k < 0 {
		return
	}

	deletedUser := Data.Users[k]

	Data.Users = append(Data.Users[:k], Data.Users[k+1:]...)

	Lock++

	Log.Info(fmt.Sprintf("%s -- %s deleted by %s", c.ClientIP(), deletedUser.Name, c.MustGet("Login").(string)))

	SetFlashCookie(c, "success", "«"+deletedUser.Name+"» deleted")
	c.Redirect(302, "/auth/crud/user")
}

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

	// Validate and register newpass
	change := false
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

	// Manage removed pass app
	for d := 0; d < 3; d++ {
		input := fmt.Sprintf("inputDelPassApp%d", d)
		delpass := c.PostForm(input)
		if delpass != "" {
			(&Data.Users[k]).DelPassApp(d)
			change = true
			Log.Info(fmt.Sprintf("%s -- %s passapp removed %d by %s", c.ClientIP(), u.Name, d, username))
		}
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
