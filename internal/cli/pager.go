package cli

import (
	"io"
	"os"
	"os/exec"
	"strings"

	"golang.org/x/term"
)

func writeMaybePaged(text string) error {
	if !shouldPage(text) {
		_, err := io.WriteString(os.Stdout, text+"\n")
		return err
	}

	args := pagerArgs()
	if len(args) == 0 {
		_, err := io.WriteString(os.Stdout, text+"\n")
		return err
	}

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdin = strings.NewReader(text)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func shouldPage(text string) bool {
	if text == "" {
		return false
	}
	if os.Getenv("KUBEMEMO_NO_PAGER") == "1" {
		return false
	}
	if !term.IsTerminal(int(os.Stdout.Fd())) || !term.IsTerminal(int(os.Stdin.Fd())) {
		return false
	}
	_, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || height <= 0 {
		return false
	}
	lineCount := strings.Count(text, "\n") + 1
	return lineCount >= height-1
}

func pagerArgs() []string {
	pager := strings.TrimSpace(os.Getenv("PAGER"))
	if pager == "" {
		pager = "less -R"
	}
	args := strings.Fields(pager)
	if len(args) == 0 {
		return nil
	}
	if args[0] == "less" && !containsArg(args[1:], "-R") && !containsArg(args[1:], "-r") {
		args = append(args, "-R")
	}
	if _, err := exec.LookPath(args[0]); err != nil {
		return nil
	}
	return args
}

func containsArg(args []string, needle string) bool {
	for _, arg := range args {
		if arg == needle {
			return true
		}
	}
	return false
}
