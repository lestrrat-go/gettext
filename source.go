package gettext

import "io/ioutil"

func (f FileSystemSource) ReadFile(s string) ([]byte, error) {
	return ioutil.ReadFile(s)
}
