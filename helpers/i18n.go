package helpers

import (
	"github.com/kataras/i18n"
)

var I18n *i18n.I18n

func Tr(lang string, x string, o ...interface{}) string {
	res := I18n.Tr(lang, x, o...)
	if res != "" {
		return res
	}
	return x
}
