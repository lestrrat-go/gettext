/*
Package gettext implements GNU gettext utilities.
*/
package gettext

import "fmt"

func format(str string, vars ...interface{}) string {
	return fmt.Sprintf(str, vars...)
}
