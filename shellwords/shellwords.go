package shellwords

import (
	"regexp"
)

var rexShellChars = regexp.MustCompile(`([\\'" \*\?\[\$\(\)\!\@\&\{])`)

func Escape(str string) string {

	return rexShellChars.ReplaceAllString(str, `\$1`)
}
