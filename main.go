package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"slices"
	"sort"
	"strconv"
	"strings"
)

const fileName string = "go.tar.gz"
const oldGoBakDir string = "/usr/local/go-bak"

func main() {
	versions, err := getVersions()
	if err != nil {
		fmt.Println("failed to get versions: ", err)
		os.Exit(1)
	}
	fmt.Println("Current version: ", strings.TrimPrefix(runtime.Version(), "go"))
	fmt.Println("Versions:", strings.Join(versions, ", "))

	version := getVersion()
	if !slices.Contains(versions, version) {
		fmt.Println("Invalid version selected: ", version)
		os.Exit(1)
	}
	if version == strings.TrimPrefix(runtime.Version(), "go") {
		fmt.Println("Version already installed.")
		os.Exit(0)
	}

	url := fmt.Sprintf("https://go.dev/dl/go%s.linux-amd64.tar.gz", version)

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

	if err := moveOldVersion(); err != nil {
		fmt.Println("failed to move the old version: ", err)
		return
	}

	if err := installNewVersion(); err != nil {
		fmt.Println("failed to install the new version: ", err)
		return
	}

	if err := os.Remove(fileName); err != nil {
		fmt.Println("failed to remove the file: ", err)
		return
	}

	if err := removeOldVersionBak(); err != nil {
		fmt.Println("failed to remove old version: ", err)
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

func getVersions() ([]string, error) {
	type release struct {
		Version string `json:"version"`
		Stable  bool   `json:"stable"`
	}

	res, err := http.Get("https://go.dev/dl/?mode=json&include=all")
	if err != nil {
		return nil, fmt.Errorf("failed to get the tags: %w", err)
	}
	defer func(Body io.ReadCloser) {
		if err := Body.Close(); err != nil {
			fmt.Println("failed to close the response body: ", err)
		}
	}(res.Body)

	var releases []release
	if err := json.NewDecoder(res.Body).Decode(&releases); err != nil {
		return nil, fmt.Errorf("failed to parse the body: %w", err)
	}

	versions := make([]string, 0)
	for _, r := range releases {
		if r.Stable {
			versions = append(versions, strings.TrimPrefix(r.Version, "go"))
		}
	}

	sort.Slice(versions, func(i, j int) bool {
		parts1 := strings.Split(versions[i], ".")
		parts2 := strings.Split(versions[j], ".")

		for k := 0; k < len(parts1) && k < len(parts2); k++ {
			if parts1[k] != parts2[k] {
				p1, _ := strconv.Atoi(parts1[k])
				p2, _ := strconv.Atoi(parts2[k])
				return p1 > p2
			}
		}

		return len(parts1) > len(parts2)
	})

	return versions, nil
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

func moveOldVersion() error {
	mv := exec.Command("mv", "/usr/local/go", oldGoBakDir)

	_, err := mv.Output()
	if err != nil {
		return fmt.Errorf("failed to execute mv: %w", err)
	}

	fmt.Println("Moved previous installation to: ", oldGoBakDir)

	return nil
}

func removeOldVersionBak() error {
	rm := exec.Command("rm", "-rf", oldGoBakDir)

	_, err := rm.Output()
	if err != nil {
		return fmt.Errorf("failed to execute rm: %w", err)
	}

	fmt.Println("Removed previous installation.")

	return nil
}
