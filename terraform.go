package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
	"sort"
	"strings"

	ver "github.com/hashicorp/go-version"
)

type Terraform struct {
	location string
	verbose  bool
	versions ver.Collection
}

func NewTerraform(location string, verbose bool) (*Terraform, error) {
	location = expandPath(location)
	createDir(location)
	tf := &Terraform{
		location: strings.TrimRight(location, "/"),
		verbose:  verbose,
	}
	files, err := ioutil.ReadDir(tf.location)
	if err != nil {
		return tf, err
	}

	for _, f := range files {
		v, err := ver.NewVersion(f.Name())
		if err != nil {
			return tf, err
		}
		tf.versions = append(tf.versions, v)
	}
	return tf, nil
}

func (tf *Terraform) String() string {
	o := []string{}
	for _, v := range tf.versions {
		o = append(o, v.String())
	}
	return strings.Join(o, "\n")
}

func (tf *Terraform) FindLatest(c ver.Constraints) (*ver.Version, error) {
	var latest *ver.Version

	if len(tf.versions) == 0 {
		return latest, fmt.Errorf("no binaries available in %s", tf.location)
	}

	for _, v := range tf.versions {
		if c.Check(v) && (latest == nil || v.GreaterThan(latest)) {
			latest = v
		}
	}
	if latest == nil {
		return latest, fmt.Errorf("no matching version found for %s", c.String())
	}

	return latest, nil
}

func (tf *Terraform) ListInstalled() ver.Collection {
	return tf.versions
}

func (tf *Terraform) ListAvailable() (ver.Collection, error) {
	out := ver.Collection{}

	type releaseInfo struct {
		Versions map[string]struct {
			Builds []struct {
				Os   string `json:"os"`
				Arch string `json:"arch"`
			} `json:"builds"`
		} `json:"versions"`
	}

	url := "https://releases.hashicorp.com/terraform/index.json"
	resp, err := http.Get(url)
	if err != nil {
		return out, fmt.Errorf("could not download %s: %s", url, err.Error())
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return out, fmt.Errorf("could not download %s: %s", url, resp.Status)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return out, err
	}

	releases := releaseInfo{}
	err = json.Unmarshal(body, &releases)
	if err != nil {
		return out, err
	}

	for vs, spec := range releases.Versions {
		version, err := ver.NewVersion(vs)
		if err != nil {
			continue
		}
		for _, build := range spec.Builds {
			if build.Arch == runtime.GOARCH && build.Os == runtime.GOOS {
				out = append(out, version)
			}
		}
	}

	sort.Sort(out)
	return out, nil
}

func (tf *Terraform) Run(v *ver.Version, args []string, w wrapper) (*os.ProcessState, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	pa := os.ProcAttr{
		Files: []*os.File{os.Stdin, os.Stdout, os.Stderr},
		Dir:   wd,
	}

	bin := fmt.Sprintf("%s/%s", tf.location, v.String())

	cmd, args, err := w.Wrap(bin, args, tf.verbose)
	if err != nil {
		return nil, err
	}

	proc, err := os.StartProcess(cmd, append([]string{"terraform"}, args...), &pa)
	if err != nil {
		return nil, err
	}

	status, err := proc.Wait()
	if err != nil {
		return status, err
	}

	return status, w.Cleanup()
}

func (tf *Terraform) DownloadVersion(v *ver.Version) (string, error) {
	url := fmt.Sprintf("https://releases.hashicorp.com/terraform/%s/terraform_%s_%s_%s.zip", v.String(), v.String(), runtime.GOOS, runtime.GOARCH)
	filename := fmt.Sprintf("%s/%s", tf.location, v.String())

	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("could not download %s: %s", url, err.Error())
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return "", fmt.Errorf("could not download %s: %s", url, resp.Status)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	zipReader, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
	if err != nil {
		return "", err
	}

	unpacked := false
	for _, zipped := range zipReader.File {
		if zipped.Name != "terraform" {
			continue
		}

		t, err := zipped.Open()
		if err != nil {
			return "", err
		}
		defer t.Close()

		b, err := ioutil.ReadAll(t)
		if err != nil {
			return "", err
		}

		dest, err := os.Create(filename)
		if err != nil {
			return "", err
		}
		defer dest.Close()

		_, err = dest.Write(b)
		if err != nil {
			return "", err
		}
		dest.Sync()
		unpacked = true
	}

	if !unpacked {
		return "", fmt.Errorf("could not find file `terraform` not found in downloaded zip")
	}

	if err := os.Chmod(filename, 0700); err != nil {
		return "", err
	}

	return filename, nil
}
