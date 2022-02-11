package handlers

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"encoding/base64"

	"github.com/skip2/go-qrcode"
	"github.com/xlzd/gotp"

	. "glauth-ui-light/config"
	. "glauth-ui-light/helpers"
)

// Validate entries

var rxEmail = regexp.MustCompile(".+@.+\\..+") //nolint
var rxName = regexp.MustCompile("^[a-z0-9]+$")
var rxASCII = regexp.MustCompile("^[A-Za-z0-9]+$")
var rxBadChar = regexp.MustCompile("[<>&*%$'«».,;:!` ]+")

type UserForm struct {
	UIDNumber    int
	Name         string
	Mail         string
	SN           string
	GivenName    string
	Password     string
	OTPSecret    string
	PrimaryGroup int
	OtherGroups  []int
	Disabled     bool
	Errors       map[string]string
	Lang         string
}

func (msg *UserForm) Validate(cfg PassPolicy) bool {
	lang := msg.Lang
	msg.Errors = make(map[string]string)

	match := rxEmail.MatchString(msg.Mail)
	if msg.Mail != "" && !match {
		msg.Errors["Mail"] = Tr(lang, "Please enter a valid email address")
	}

	p := msg.Password
	if p != "" {
		switch {
		case len(p) < cfg.Min:
			msg.Errors["Password"] = Tr(lang, "Too short")
		case len(p) > cfg.Max:
			msg.Errors["Password"] = Tr(lang, "Too long")
		}
	}

	o := msg.OTPSecret
	matchAscii := rxASCII.MatchString(o)
	if o != "" {
		switch {
		case len(o) < 16:
			msg.Errors["OTPSecret"] = Tr(lang, "Too short")
		case len(o) > 33:
			msg.Errors["OTPSecret"] = Tr(lang, "Too long")
		case !matchAscii:
			msg.Errors["OTPSecret"] = Tr(lang, "Bad character")
		}
	}

	n := msg.Name
	matchName := rxName.MatchString(n)
	switch {
	case strings.TrimSpace(n) == "":
		msg.Errors["Name"] = Tr(lang, "Mandatory")
	case len(n) < 2:
		msg.Errors["Name"] = Tr(lang, "Too short")
	case len(n) > 16:
		msg.Errors["Name"] = Tr(lang, "Too long")
	case !matchName:
		msg.Errors["Name"] = Tr(lang, "Bad character")
	}
	for k := range Data.Users {
		if Data.Users[k].Name == n && Data.Users[k].UIDNumber != msg.UIDNumber {
			msg.Errors["Name"] = Tr(lang, "Name already used")
			break
		}
	}

	matchBadSN := rxBadChar.MatchString(msg.SN)
	if msg.SN != "" && len(msg.SN) > 32 {
		msg.Errors["SN"] = Tr(lang, "Too long")
	}
	if msg.SN != "" && matchBadSN {
		msg.Errors["SN"] = Tr(lang, "Bad character")
	}

	matchBadGname := rxBadChar.MatchString(msg.GivenName)
	if msg.GivenName != "" && len(msg.GivenName) > 32 {
		msg.Errors["GivenName"] = Tr(lang, "Too long")
	}
	if msg.GivenName != "" && matchBadGname {
		msg.Errors["GivenName"] = Tr(lang, "Bad character")
	}

	if msg.UIDNumber < 0 {
		msg.Errors["UIDNumber"] = Tr(lang, "Unknown user")
	}

	return len(msg.Errors) == 0
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
	render(c, gin.H{"title": Tr(lang, "Users Page"), "currentPage": "user", "userdata": Data.Users, "hashgroups": hg}, "user/list.tmpl")
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
		UIDNumber:    u.UIDNumber,
		Mail:         u.Mail,
		Name:         u.Name,
		PrimaryGroup: u.PrimaryGroup,
		OtherGroups:  u.OtherGroups,
		SN:           u.SN,
		GivenName:    u.GivenName,
		Disabled:     u.Disabled,
		OTPSecret:    u.OTPSecret,
		Lang:         lang,
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
		UIDNumber:    u.UIDNumber,
		Mail:         u.Mail,
		Name:         u.Name,
		PrimaryGroup: u.PrimaryGroup,
		OtherGroups:  u.OtherGroups,
		SN:           u.SN,
		GivenName:    u.GivenName,
		Disabled:     u.Disabled,
		Lang:         lang,
	}

	if u.OTPSecret != "" {
		totp := gotp.NewDefaultTOTP(u.OTPSecret)
		sec := totp.ProvisioningUri(u.Name, cfg.AppName)
		var png []byte
		png, _ = qrcode.Encode(sec, qrcode.Medium, 256)
		img := base64.StdEncoding.EncodeToString(png)
		userf.OTPSecret = img
	}

	render(c, gin.H{"title": u.Name, "u": userf, "currentPage": "profile", "groupdata": Data.Groups}, "user/profile.tmpl")
}

func UserChgPasswd(c *gin.Context) {
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
	role := c.MustGet("Role").(string)
	// application accounts don't change their password
	if role != "admin" && role != "user" { // users and admins are defined by group set by GIDcanChgPass, GIDAdmin config
		render(c, gin.H{"title": u.Name, "currentPage": "profile", "u": u, "groupdata": Data.Groups}, "user/profile.tmpl")
		return
	}

	userf := &UserForm{
		UIDNumber:    u.UIDNumber,
		Mail:         u.Mail,
		Name:         u.Name,
		PrimaryGroup: u.PrimaryGroup,
		OtherGroups:  u.OtherGroups,
		SN:           u.SN,
		GivenName:    u.GivenName,
		Disabled:     u.Disabled,
		Lang:         lang,
	}
	userf.Errors = make(map[string]string)

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

	if Lock != 0 {
		render(c, gin.H{
			"title":       u.Name,
			"currentPage": "profile",
			"warning":     Tr(lang, "Data locked by admin."),
			"u":           userf,
			"groupdata":   Data.Groups},
			"user/profile.tmpl")
	} else {
		err := WriteDB(&cfg, Data, username)
		if err != nil {
			render(c, gin.H{"title": Tr(lang, "Error"), "currentPage": "profile", "error": err.Error()}, "home/error.tmpl")
		} else {
			render(c, gin.H{
				"title":       u.Name,
				"currentPage": "profile",
				"success":     Tr(lang, "Password updated"),
				"u":           userf,
				"groupdata":   Data.Groups},
				"user/profile.tmpl")
		}
	}
}
