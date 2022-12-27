//go:build ignore
// +build ignore

package main

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func queryReleases(user, repo string) ([]byte, error) {
	resp, err := http.Get(
		fmt.Sprintf("https://api.github.com/repos/%s/%s/releases", user, repo))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadUrl string `json:"browser_download_url"`
	tagName            string
}

type Release struct {
	TagName string   `json:"tag_name"`
	Assets  []*Asset `json:"assets"`
}

func getReleases(name, repo string) ([]*Release, error) {
	releasesStr, err := queryReleases(name, repo)
	if err != nil {
		return nil, fmt.Errorf("getReleases: %w", err)
	}
	var releases []*Release
	if err := json.Unmarshal(releasesStr, &releases); err != nil {
		return nil, fmt.Errorf("json.Unmarshal: %w", err)
	}
	return releases, nil
}

func seekAssets(releases []*Release, fname string) (string, string) {
	for _, rel := range releases {
		for _, a := range rel.Assets {
			if a.Name == fname {
				return a.BrowserDownloadUrl, rel.TagName
			}
		}
	}
	return "", ""
}

func getHash(fname string) (string, error) {
	fd, err := os.Open(fname)
	if err != nil {
		return "", err
	}
	defer fd.Close()

	h := sha256.New()
	io.Copy(h, fd)
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

type Archtecture struct {
	Url  string `json:"url"`
	Hash string `json:"hash,omitempty"`
}

type AutoUpdate struct {
	Archtectures map[string]*Archtecture `json:"architecture"`
}

type Manifest struct {
	Version      string                  `json:"version"`
	Description  string                  `json:"description,omitempty"`
	Homepage     string                  `json:"homepage,omitempty"`
	License      string                  `json:"license,omitempty"`
	Archtectures map[string]*Archtecture `json:"architecture"`
	Bin          string                  `json:"bin"`
	CheckVer     map[string]string       `json:"checkver"`
	AutoUpdate   AutoUpdate              `json:"autoupdate"`
}

func getBits(s string) string {
	if strings.Contains(s, "386") {
		return "32bit"
	} else if strings.Contains(s, "amd64") {
		return "64bit"
	}
	return ""
}

func queryDescription(user, repo string) ([]byte, error) {
	resp, err := http.Get(
		fmt.Sprintf("https://api.github.com/repos/%s/%s", user, repo))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

type Description struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	License     map[string]string `json:"license"`
}

func getDescription(user, repo string) (*Description, error) {
	bin, err := queryDescription(user, repo)
	if err != nil {
		return nil, err
	}
	var desc *Description
	if err = json.Unmarshal(bin, &desc); err != nil {
		return nil, err
	}
	return desc, nil
}

var flagInlineTemplate = flag.String("inline", "", "Set template inline")

func mains(args []string) error {
	if len(args) < 3 {
		return errors.New("Usage: go run mkmanifest.go USER REPO ZIP1...")
	}
	name := args[0]
	repo := args[1]

	releases, err := getReleases(name, repo)
	if err != nil {
		return fmt.Errorf("getReleases: %w", err)
	}
	arch := make(map[string]*Archtecture)
	var url, tag string
	for _, arg1 := range args[2:] {
		files, err := filepath.Glob(arg1)
		if err != nil {
			files = []string{arg1}
		}
		for _, fname := range files {
			bits := getBits(fname)
			if bits == "" {
				return fmt.Errorf("%s: can not find `386` nor `amd64`", fname)
			}
			url, tag = seekAssets(releases, fname)
			if url == "" {
				return fmt.Errorf("%s not found in remote repository", fname)
			}
			hash, err := getHash(fname)
			if err != nil {
				return err
			}
			arch[bits] = &Archtecture{
				Url:  url,
				Hash: hash,
			}
		}
	}
	var input []byte
	if *flagInlineTemplate != "" {
		input = []byte(*flagInlineTemplate)
	} else {
		input, err = io.ReadAll(os.Stdin)
		if err != nil && err != io.EOF {
			return err
		}
	}
	var manifest Manifest
	if err = json.Unmarshal(input, &manifest); err != nil {
		return err
	}
	if manifest.Archtectures == nil {
		manifest.Archtectures = make(map[string]*Archtecture)
	}
	if manifest.AutoUpdate.Archtectures == nil {
		manifest.AutoUpdate.Archtectures = map[string]*Archtecture{}
	}
	if manifest.Homepage == "" {
		manifest.Homepage = fmt.Sprintf(
			"https://github.com/%s/%s", name, repo)
	}
	if manifest.CheckVer == nil {
		manifest.CheckVer = make(map[string]string)
	}
	if _, ok := manifest.CheckVer["github"]; !ok {
		manifest.CheckVer["github"] = fmt.Sprintf(
			"https://github.com/%s/%s", name, repo)
	}
	if _, ok := manifest.CheckVer["regex"]; !ok {
		manifest.CheckVer["regex"] = "tag/([\\d\\._]+)"
	}
	for name, val := range arch {
		manifest.Archtectures[name] = val
		manifest.Version = strings.TrimPrefix(tag, "v")

		autoupdate := strings.ReplaceAll(val.Url, manifest.Version, "$version")
		bits := getBits(val.Url)
		manifest.AutoUpdate.Archtectures[bits] = &Archtecture{Url: autoupdate}
	}
	if desc, err := getDescription(name, repo); err == nil {
		if manifest.Description == "" {
			manifest.Description = desc.Description
		}
		if manifest.License == "" {
			manifest.License = desc.License["name"]
		}
	}

	output, err := json.MarshalIndent(&manifest, "", "    ")
	if err != nil {
		return err
	}
	os.Stdout.Write(output)
	return nil
}

func main() {
	flag.Parse()
	if err := mains(flag.Args()); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
