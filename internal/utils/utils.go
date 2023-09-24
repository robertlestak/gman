package utils

import (
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/fhs/go-netrc/netrc"
	log "github.com/sirupsen/logrus"
)

func IsOnlyUrl(s string) bool {
	l := log.WithField("fn", "IsOnlyUrl")
	l.Debug("checking if string is only url")
	// if the string is only one line, and that
	// line is a url, then we can assume it's a url
	lines := strings.Split(s, "\n")
	if len(lines) > 1 {
		l.Debug("string is not only one line")
		return false
	}
	// parse string as url
	_, err := url.ParseRequestURI(strings.TrimSpace(s))
	if err != nil {
		l.Debug("string is not a url")
		return false
	}
	l.Debug("string is only url")
	return true
}

func OpenURL(u string) error {
	openCmd := "xdg-open"
	if runtime.GOOS == "darwin" {
		openCmd = "open"
	} else if runtime.GOOS == "windows" {
		openCmd = "start"
	}
	cmd := exec.Command(openCmd, u)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func StringSearch(s string, search string) bool {
	if strings.Contains(s, search) {
		return true
	}
	// try regex
	rx := regexp.MustCompile(search)
	if rx.MatchString(s) {
		return true
	}
	return false
}

func AuthForDomain(domain string) (login *string, password *string) {
	// check if there is a ~/.netrc file
	// if so, check if there is a machine entry for the domain
	// if so, return the token
	// if not, return nil
	// if there is no ~/.netrc file, return nil
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, nil
	}
	netrcPath := path.Join(home, ".netrc")
	if _, err := os.Stat(netrcPath); os.IsNotExist(err) {
		return nil, nil
	}
	m, err := netrc.FindMachine(netrcPath, domain)
	if err != nil {
		return nil, nil
	}
	return &m.Login, &m.Password
}

func GetRemote(u string) (string, error) {
	l := log.WithField("fn", "GetRemote")
	l.Debug("getting remote")
	u = strings.TrimSpace(u)
	ud, err := url.Parse(u)
	if err != nil {
		return "", err
	}
	c := &http.Client{}
	req, err := http.NewRequest("GET", u, nil)
	// check if we have a token for the domain
	_, token := AuthForDomain(ud.Host)
	if token != nil {
		req.Header.Add("Authorization", "token "+*token)
	}
	res, err := c.Do(req)
	if err != nil {
		l.WithError(err).Error("error getting remote")
		return "", err
	}
	l.Debug("remote gotten")
	defer res.Body.Close()
	bd, err := io.ReadAll(res.Body)
	if err != nil {
		l.WithError(err).Error("error reading remote")
		return "", err
	}
	// if status code is not in the 200 range, return error
	if res.StatusCode < 200 || res.StatusCode > 299 {
		l.WithError(err).Debug("error getting remote")
		err = errors.New("get error: " + strconv.Itoa(res.StatusCode))
	}
	l.Debug("remote read")
	return string(bd), err
}
