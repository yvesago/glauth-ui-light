package handlers

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	. "glauth-ui-light/config"
	. "glauth-ui-light/helpers"
)

// Validate entries

// var rxName = regexp.MustCompile("^[a-z0-9]+$")

type GroupForm struct {
	GIDNumber int
	Name      string
	Errors    map[string]string
	Lang      string
}

func (msg *GroupForm) Validate() bool {
	lang := msg.Lang
	msg.Errors = make(map[string]string)

	n := msg.Name
	matchAscii := rxName.MatchString(n)
	switch {
	case strings.TrimSpace(n) == "":
		msg.Errors["Name"] = Tr(lang, "Mandatory")
	case len(n) < 2:
		msg.Errors["Name"] = Tr(lang, "Too short")
	case len(n) > 16:
		msg.Errors["Name"] = Tr(lang, "Too long")
	case !matchAscii:
		msg.Errors["Name"] = Tr(lang, "Bad character")
	}
	for k := range Data.Groups {
		if Data.Groups[k].Name == n && Data.Groups[k].GIDNumber != msg.GIDNumber {
			msg.Errors["Name"] = Tr(lang, "Name already used")
			break
		}
	}

	if msg.GIDNumber < 0 {
		msg.Errors["GIDNumber"] = Tr(lang, "Unknown group")
	}

	return len(msg.Errors) == 0
}

// Helpers

func ctlGroupExist(c *gin.Context, lang string, id string) int {
	k := GetGroupKey(id)
	if k < 0 {
		render(c, gin.H{"title": Tr(lang, "Error"), "currentPage": "group", "error": Tr(lang, "Unknown group")}, "home/error.tmpl")
		return -1
	}
	return k
}

func GetGroupKey(id string) int {
	i := -1
	intId, _ := strconv.Atoi(id)
	for k := range Data.Groups {
		if Data.Groups[k].GIDNumber == intId {
			i = k
			break
		}
	}
	return i
}

func GetGroupByID(id int) (Group, error) {
	for k := range Data.Groups {
		if Data.Groups[k].GIDNumber == id {
			return Data.Groups[k], nil
		}
	}
	return Group{}, fmt.Errorf("unknown group")
}

type SpecialGroups struct {
	Admins string
	Users  string
	OTP    string
}

func GetSpecialGroups(c *gin.Context) SpecialGroups {
	cfg := c.MustGet("Cfg").(WebConfig)
	s := SpecialGroups{}
	g, err := GetGroupByID(cfg.CfgUsers.GIDAdmin)
	if err == nil {
		s.Admins = g.Name
	}
	g, err = GetGroupByID(cfg.CfgUsers.GIDcanChgPass)
	if err == nil {
		s.Users = g.Name
	}
	g, err = GetGroupByID(cfg.CfgUsers.GIDuseOtp)
	if err == nil {
		s.OTP = g.Name
	}

	return s
}

/*func GetGroupByName(name string) (Group, error) {
	for _, v := range Data.Groups {
		if v.Name == name {
			return v, nil
		}
	}
	return Group{}, fmt.Errorf("unknown group")
}*/

func contains(s []int, e int) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func isGroupEmpty(id int) bool {
	for k := range Data.Users {
		if Data.Users[k].PrimaryGroup == id {
			return false
		}
		if contains(Data.Users[k].OtherGroups, id) {
			return false
		}
	}
	return true
}

// Handlers

func GroupList(c *gin.Context) {
	cfg := c.MustGet("Cfg").(WebConfig)
	lang := cfg.Locale.Lang

	if !isAdminAccess(c, "GroupList", "-") {
		return
	}

	hg := make(map[int]string)
	for k := range Data.Groups {
		hg[Data.Groups[k].GIDNumber] = Data.Groups[k].Name
	}
	render(c, gin.H{"title": Tr(lang, "Groups page"), "currentPage": "group", "groupdata": Data.Groups, "hashgroups": hg}, "group/list.tmpl")
}

func GroupEdit(c *gin.Context) {
	cfg := c.MustGet("Cfg").(WebConfig)
	lang := cfg.Locale.Lang
	id := c.Params.ByName("id")

	if !isAdminAccess(c, "GroupEdit", id) {
		return
	}

	k := ctlGroupExist(c, lang, id)
	if k < 0 {
		return
	}

	u := Data.Groups[k]
	groupf := GroupForm{
		GIDNumber: u.GIDNumber,
		Name:      u.Name,
		Lang:      lang,
	}

	render(c, gin.H{"title": Tr(lang, "Edit group"), "currentPage": "group", "u": groupf, "groupdata": Data.Groups}, "group/edit.tmpl")
}

func GroupUpdate(c *gin.Context) {
	cfg := c.MustGet("Cfg").(WebConfig)
	lang := cfg.Locale.Lang
	id := c.Params.ByName("id")

	if !isAdminAccess(c, "GroupUpdate", id) {
		return
	}

	k := ctlGroupExist(c, lang, id)
	if k < 0 {
		return
	}

	// Bind form to struct
	groupf := &GroupForm{
		GIDNumber: Data.Groups[k].GIDNumber,
		Name:      c.PostForm("inputName"),
		Lang:      lang,
	}
	// fmt.Printf("%+v\n", groupf)

	// Validate entries
	if !groupf.Validate() {
		render(c, gin.H{"title": Tr(lang, "Edit group"), "currentPage": "group", "u": groupf, "groupdata": Data.Groups}, "group/edit.tmpl")
		return
	}

	// Update Data
	(&Data.Groups[k]).Name = groupf.Name

	Lock++

	Log.Info(fmt.Sprintf("%s -- %s updated by %s", c.ClientIP(), groupf.Name, c.MustGet("Login").(string)))

	render(c, gin.H{
		"title":       Tr(lang, "Edit group"),
		"currentPage": "group",
		"success":     "«" + groupf.Name + "» updated",
		"u":           groupf,
		"groupdata":   Data.Groups},
		"group/edit.tmpl")
}

func GroupAdd(c *gin.Context) {
	cfg := c.MustGet("Cfg").(WebConfig)
	lang := cfg.Locale.Lang

	if !isAdminAccess(c, "GroupAdd", "-") {
		return
	}

	render(c, gin.H{"title": Tr(lang, "Add group"), "currentPage": "group"}, "group/create.tmpl")
}

func GroupCreate(c *gin.Context) {
	cfg := c.MustGet("Cfg").(WebConfig)
	lang := cfg.Locale.Lang

	if !isAdminAccess(c, "GroupCreate", "-") {
		return
	}

	// Bind form to struct
	groupf := &GroupForm{
		Name: c.PostForm("inputName"),
		Lang: lang,
	}
	// Validate entries
	if !groupf.Validate() {
		render(c, gin.H{"title": Tr(lang, "Add group"), "currentPage": "group", "u": groupf, "groupdata": Data.Groups}, "group/create.tmpl")
		return
	}

	// Create new id
	nextID := cfg.CfgGroups.Start - 1 // start uidnumber via config
	for k := range Data.Groups {
		if Data.Groups[k].GIDNumber >= nextID {
			nextID = Data.Groups[k].GIDNumber
		}
	}
	groupf.GIDNumber = nextID + 1
	// Add Group to Data
	newGroup := Group{GIDNumber: groupf.GIDNumber, Name: groupf.Name}
	Data.Groups = append(Data.Groups, newGroup)

	Lock++

	Log.Info(fmt.Sprintf("%s -- %s created by %s", c.ClientIP(), newGroup.Name, c.MustGet("Login").(string)))

	SetFlashCookie(c, "success", "«"+newGroup.Name+"» added")
	c.Redirect(302, fmt.Sprintf("/auth/crud/group/%d", newGroup.GIDNumber))
}

func GroupDel(c *gin.Context) {
	cfg := c.MustGet("Cfg").(WebConfig)
	lang := cfg.Locale.Lang
	id := c.Params.ByName("id")

	if !isAdminAccess(c, "GroupDel", id) {
		return
	}

	k := ctlGroupExist(c, lang, id)
	if k < 0 {
		return
	}

	deletedGroup := Data.Groups[k]

	if isGroupEmpty(deletedGroup.GIDNumber) {
		Data.Groups = append(Data.Groups[:k], Data.Groups[k+1:]...)

		Lock++

		Log.Info(fmt.Sprintf("%s -- %s deleted by %s", c.ClientIP(), deletedGroup.Name, c.MustGet("Login").(string)))

		SetFlashCookie(c, "success", "«"+deletedGroup.Name+"» deleted")
	} else {
		SetFlashCookie(c, "warning", Tr(lang, "Group must be empty before being deleted"))
	}
	c.Redirect(302, "/auth/crud/group")
}
