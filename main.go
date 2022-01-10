package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/google/go-github/v41/github"
	"github.com/hashicorp/go-getter"
	"github.com/sethvargo/go-githubactions"
)

var downloadURLTemplate = "https://github.com/kubernetes-sigs/krew/releases/download/%s/krew-%s_%s.tar.gz"

func main() {
	// INPUT_VERSION
	version := os.Args[1]
	var err error
	if version == "latest" {
		version, err = getTheLatestVersion()
		if err != nil {
			githubactions.Fatalf("could not found the latest release of kubernetes-sigs/krew: %v", err)
		}
	}
	// https://github.com/kubernetes-sigs/krew/releases/download/v0.4.2/krew-darwin_amd64.tar.gz
	// https://github.com/kubernetes-sigs/krew/releases/download/v0.4.2/krew-darwin_amd64.tar.gz.sha256
	releaseDownloadURL := fmt.Sprintf(downloadURLTemplate, version, runtime.GOOS, runtime.GOARCH)
	td, err := ioutil.TempDir("", "setup-krew")
	if err != nil {
		githubactions.Fatalf("could not get working directory: %v", err)
	}

	u, err := url.Parse(releaseDownloadURL)
	if err != nil {
		githubactions.Fatalf("could not pare URL %s: %v", releaseDownloadURL, err)
	}

	fg := new(getter.HttpGetter)
	binaryTarGzFile := filepath.Join(td, fmt.Sprintf("krew-%s_%s.tar.gz", runtime.GOOS, runtime.GOARCH))
	err = fg.GetFile(binaryTarGzFile, u)
	if err != nil {
		log.Fatalf("could not download file from URL %s: %v", releaseDownloadURL, err)
	}
	currentDirFiles, err := exec.Command("ls", "-latr", td).CombinedOutput()
	if err != nil {
		log.Fatalf("could not run ls in directory %s: %v", td, err)
	}
	fmt.Println(string(currentDirFiles))

	tarDecompressor := new(getter.TarGzipDecompressor)
	err = tarDecompressor.Decompress(td, binaryTarGzFile, true, 0600)
	if err != nil {
		log.Fatalf("could not run extract .tar.gz file %s: %v", binaryTarGzFile, err)
	}

	gw := os.Getenv("HOME")
	fmt.Println(gw)
	installationPath := fmt.Sprintf("%s/%s/%s", gw, ".setup-krew", "bin")
	_, err = exec.Command("mkdir", "-p", installationPath).CombinedOutput()
	if err != nil {
		log.Fatalf("could not create the home directory for krew: %v", err)
	}

	_, err = exec.Command("mv", filepath.Join(td, fmt.Sprintf("krew-%s_%s", runtime.GOOS, runtime.GOARCH)),
		filepath.Join(installationPath, "krew")).CombinedOutput()
	if err != nil {
		log.Fatalf("could not rename the binary: %v", err)
	}

	currentDirFiles, err = exec.Command("ls", "-latr", installationPath).CombinedOutput()
	if err != nil {
		log.Fatalf("could not run ls in directory %s: %v", installationPath, err)
	}
	fmt.Println(string(currentDirFiles))

	_, err = exec.Command("sh", "-c", fmt.Sprintf("echo \"%s\" >> %s", installationPath, os.Getenv("GITHUB_PATH"))).CombinedOutput()
	if err != nil {
		log.Fatalf("could not add binary to \"GITHUB_PATH\": %v", err)
	}
	fmt.Println(os.Getenv("GITHUB_PATH"))

	// PSEUDO:
	// get input from user
	// if version == latest --> get the latest version from
}

func getTheLatestVersion() (string, error) {
	client := github.NewClient(nil)
	t, _, err := client.Repositories.ListTags(context.Background(), "kubernetes-sigs", "krew", &github.ListOptions{})
	if err != nil {
		return "", err
	}

	if len(t) == 0 {
		return "", fmt.Errorf("could not any valid tag for kubernetes-sigs/krew")
	}

	return *t[0].Name, nil
}
