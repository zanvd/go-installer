package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

const fileName string = "go.tar.gz"

func main() {
	url := fmt.Sprintf("https://go.dev/dl/go%s.linux-amd64.tar.gz", getVersion())

	f, err := os.Create(fileName)
	if err != nil {
		fmt.Println("failed to create the output file: ", err)
		os.Exit(2)
	}
	defer func(f *os.File) {
		if _, err := os.Stat(fileName); errors.Is(err, os.ErrNotExist) {
			return
		}
		if err := f.Close(); err != nil {
			fmt.Println("failed to close the file: ", err)
		}
	}(f)

	if err := downloadFile(f, url); err != nil {
		fmt.Println("failed to download the file: ", err)
		return
	}

	if err := removeCurrentVersion(); err != nil {
		fmt.Println("failed to remove current version: ", err)
		return
	}

	if err := installNewVersion(); err != nil {
		fmt.Println("failed to install the new version: ", err)
		return
	}

	if err := f.Close(); err != nil {
		fmt.Println("failed to close the file: ", err)
		return
	}
	if err := os.Remove(fileName); err != nil {
		fmt.Println("failed remove the file: ", err)
		return
	}

	fmt.Println("Done. Please run go version to confirm correct installation.")
}

func downloadFile(file io.Writer, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to get the file: %w", err)
	}
	defer func(Body io.ReadCloser) {
		if err := Body.Close(); err != nil {
			fmt.Println("failed to close the response body: ", err)
		}
	}(resp.Body)

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write to the file: %w", err)
	}

	return nil
}

func getVersion() string {
	var v string

	r := bufio.NewReader(os.Stdin)
	for {
		_, err := fmt.Fprint(os.Stdout, "Which version?\n")
		if err != nil {
			fmt.Println("failed to write to stdout: ", err)
			os.Exit(1)
		}

		v, _ = r.ReadString('\n')
		v = strings.TrimSpace(v)
		if v != "" {
			break
		}
	}

	return v
}

func installNewVersion() error {
	tar := exec.Command("tar", "-C", "/usr/local", "-xzf", fileName)

	_, err := tar.Output()
	if err != nil {
		return fmt.Errorf("failed to execute tar: %w", err)
	}

	fmt.Println("Extracted the new version.")

	return nil
}

func removeCurrentVersion() error {
	rm := exec.Command("rm", "-rf", "/usr/local/go")

	_, err := rm.Output()
	if err != nil {
		return fmt.Errorf("failed to execute rm: %w", err)
	}

	fmt.Println("Removed previous installation.")

	return nil
}
