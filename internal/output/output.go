package output

import (
	"bytes"
	"os"
	"os/exec"
	"strings"

	log "github.com/sirupsen/logrus"
)

type OutputType string

const (
	Text OutputType = "text"
	JSON OutputType = "json"
	YAML OutputType = "yaml"
)

func PagerPrint(pager string, data string) error {
	cmd := exec.Command(pager)
	cmd.Stdin = strings.NewReader(data)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func RenderPandoc(data string) (string, error) {
	// pipe data to:
	// pandoc -s -f markdown -t man | groff -T utf8 -man
	// then print that
	cmds := []string{
		"pandoc",
		"-s",
		"-f",
		"markdown",
		"-t",
		"man",
	}
	cmd := exec.Command(cmds[0], cmds[1:]...)
	cmd.Stdin = strings.NewReader(data)
	if log.GetLevel() == log.DebugLevel {
		cmd.Stderr = os.Stderr
	}
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	cmds = []string{
		"groff",
		"-T",
		"utf8",
		"-man",
	}
	cmd = exec.Command(cmds[0], cmds[1:]...)
	cmd.Stdin = strings.NewReader(string(out))
	outData := bytes.Buffer{}
	cmd.Stdout = &outData
	if log.GetLevel() == log.DebugLevel {
		cmd.Stderr = os.Stderr
	}
	err = cmd.Run()
	if err != nil {
		return "", err
	}
	return outData.String(), nil
}

func Print(render bool, pager string, data string) error {
	if render {
		var err error
		data, err = RenderPandoc(data)
		if err != nil {
			return err
		}
	}
	if pager != "" {
		return PagerPrint(pager, data)
	}
	println(data)
	return nil
}
