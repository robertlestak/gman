package utils

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
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

func getRemoteImageContent(u string) (string, error) {
	l := log.WithField("fn", "getRemoteImageContent")
	l.Debug("getting remote image content")
	u = strings.TrimSpace(u)
	ud, err := url.Parse(u)
	if err != nil {
		return "", err
	}
	c := &http.Client{}
	l.WithField("url", u).Debug("getting remote")
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
	// print the status code
	l.WithField("status_code", res.StatusCode).Debug("remote status code")
	// if status code is not in the 200 range, return error
	if res.StatusCode < 200 || res.StatusCode > 299 {
		l.WithError(err).Debug("error getting remote")
		err = errors.New("get error: " + strconv.Itoa(res.StatusCode))
	}
	l.Debug("remote read")
	// base64 encode the image
	encoded := base64.StdEncoding.EncodeToString(bd)
	// return the encoded image string
	enc := fmt.Sprintf("data:%s;base64,%s", res.Header.Get("Content-Type"), encoded)
	return enc, err
}

func fileIsImage(f string) bool {
	// get the file extension
	ext := filepath.Ext(f)
	// check if the extension is in the list of image extensions
	for _, e := range []string{".png", ".jpg", ".jpeg", ".gif", ".svg"} {
		if ext == e {
			return true
		}
	}
	return false
}

func rewriteRelativePaths(u string, data string, embedImages bool) string {
	l := log.WithField("fn", "rewriteRelativePaths")
	l.Debug("rewriting relative paths")
	rx := regexp.MustCompile(`\]\((\.\.\/)*([^\)]+)\)`)
	// find all matches
	matches := rx.FindAllStringSubmatch(data, -1)
	for _, match := range matches {
		// get the path
		p := match[2]
		// if the first character is a #, skip it
		if strings.HasPrefix(p, "#") {
			continue
		}
		// if it's not a url, rewrite it
		if !IsOnlyUrl(p) {
			// rewrite the path
			l.WithFields(log.Fields{
				"old_path": p,
				"new_path": u + "/" + p,
			}).Debug("rewriting relative path")
			if fileIsImage(p) && embedImages {
				ed, err := getRemoteImageContent(u + "/" + p)
				if err != nil {
					l.WithError(err).Error("error getting remote image content")
					continue
				}
				data = strings.ReplaceAll(data, p, ed)
			} else {
				data = strings.ReplaceAll(data, p, u+"/"+p)
			}
		}
	}
	// regex for finding relative paths in img tags
	rx = regexp.MustCompile(`<img src="(\.\.\/)*([^\"]+)"`)
	// find all matches
	matches = rx.FindAllStringSubmatch(data, -1)
	for _, match := range matches {
		// get the path
		p := match[2]
		// if the first character is a #, skip it
		if strings.HasPrefix(p, "#") {
			continue
		}
		// if it's not a url, rewrite it
		if !IsOnlyUrl(p) {
			// rewrite the path
			if fileIsImage(p) && embedImages {
				ed, err := getRemoteImageContent(u + "/" + p)
				if err != nil {
					l.WithError(err).Error("error getting remote image content")
					continue
				}
				data = strings.ReplaceAll(data, p, ed)
			} else {
				data = strings.ReplaceAll(data, p, u+"/"+p)
			}
		}
	}
	l.Debug("relative paths rewritten")
	return data
}

func GetRemote(u string, embedImages bool) (string, error) {
	l := log.WithField("fn", "GetRemote")
	l.Debug("getting remote")
	u = strings.TrimSpace(u)
	ud, err := url.Parse(u)
	if err != nil {
		return "", err
	}
	c := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// don't allow redirects
			return http.ErrUseLastResponse
		},
	}
	l.WithField("url", u).Debug("getting remote")
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
	// print the status code
	l.WithField("status_code", res.StatusCode).Debug("remote status code")
	// if status code is not in the 200 range, return error
	if res.StatusCode < 200 || res.StatusCode > 299 {
		l.WithError(err).Debug("error getting remote")
		err = errors.New("get error: " + strconv.Itoa(res.StatusCode))
	}
	l.Debug("remote read")
	// pop the last element off the path, and use that as the base url
	u = strings.Join(strings.Split(u, "/")[:len(strings.Split(u, "/"))-1], "/")
	// rewrite relative paths
	data := rewriteRelativePaths(u, string(bd), embedImages)
	return data, err
}

func SameFile(a, b string) (bool, error) {
	if a == b {
		return true, nil
	}

	aInfo, err := os.Lstat(a)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	bInfo, err := os.Lstat(b)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	return os.SameFile(aInfo, bInfo), nil
}

func Copydir(dst, src string) error {
	src, err := filepath.EvalSymlinks(src)
	if err != nil {
		return err
	}

	walkFn := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if path == src {
			return nil
		}

		if strings.HasPrefix(filepath.Base(path), ".") {
			// Skip any dot files
			if info.IsDir() {
				return filepath.SkipDir
			} else {
				return nil
			}
		}

		// The "path" has the src prefixed to it. We need to join our
		// destination with the path without the src on it.
		dstPath := filepath.Join(dst, path[len(src):])

		// we don't want to try and copy the same file over itself.
		if eq, err := SameFile(path, dstPath); eq {
			return nil
		} else if err != nil {
			return err
		}

		// If we have a directory, make that subdirectory, then continue
		// the walk.
		if info.IsDir() {
			if path == filepath.Join(src, dst) {
				// dst is in src; don't walk it.
				return nil
			}

			if err := os.MkdirAll(dstPath, 0755); err != nil {
				return err
			}

			return nil
		}

		// If the current path is a symlink, recreate the symlink relative to
		// the dst directory
		if info.Mode()&os.ModeSymlink == os.ModeSymlink {
			target, err := os.Readlink(path)
			if err != nil {
				return err
			}

			return os.Symlink(target, dstPath)
		}

		// If we have a file, copy the contents.
		srcF, err := os.Open(path)
		if err != nil {
			return err
		}
		defer srcF.Close()

		dstF, err := os.Create(dstPath)
		if err != nil {
			return err
		}
		defer dstF.Close()

		if _, err := io.Copy(dstF, srcF); err != nil {
			return err
		}

		// Chmod it
		return os.Chmod(dstPath, info.Mode())
	}

	return filepath.Walk(src, walkFn)
}

func CopyFile(src, dst string) (int64, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destination.Close()
	nBytes, err := io.Copy(destination, source)
	return nBytes, err
}
