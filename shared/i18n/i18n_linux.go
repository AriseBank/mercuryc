// +build linux

package i18n

import (
	"github.com/gosexy/gettext"
)

// G returns the translated string
func G(msgid string) string {
	return gettext.DGettext("apollo", msgid)
}

func init() {
	gettext.SetLocale(gettext.LC_ALL, "")
}
