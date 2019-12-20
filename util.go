package chainrunner

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strings"
)

func readFile(session Session, path string) ([]byte, error) {
	return getShellResult(fmt.Sprintf("cat %s", path), session)
}

func getSshDir(session Session) (string, error) {
	home, err := getShellResult("echo $USER", session)
	path := fmt.Sprintf("/home/%s/.ssh", strings.Replace(string(home), "\n", "", -1))
	return path, err
}

func getShellResult(cmd string, session Session) ([]byte, error) {
	buf := bytes.NewBuffer([]byte(``))
	oldStdout := session.GetStdout()
	session.SetStdout(buf)
	err := session.Run(cmd)
	session.SetStdout(oldStdout)
	if err != nil {
		return []byte{}, err
	}
	data, err := ioutil.ReadAll(buf)
	if err != nil {
		return []byte{}, err
	}

	return data, nil
}
