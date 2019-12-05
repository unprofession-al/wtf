package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
	"strings"

	"github.com/hashicorp/go-version"
)

type Terraform struct {
	location string
	verbose  bool
	versions version.Collection
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
		v, err := version.NewVersion(f.Name())
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

func (tf *Terraform) FindLatest(c version.Constraints) (*version.Version, error) {
	var latest *version.Version

	if len(tf.versions) == 0 {
		return latest, fmt.Errorf("No binaries available in %s", tf.location)
	}

	for _, v := range tf.versions {
		if c.Check(v) && (latest == nil || v.GreaterThan(latest)) {
			latest = v
		}
	}
	if latest == nil {
		return latest, fmt.Errorf("No matching version found for %s", c.String())
	}

	return latest, nil
}

func (tf *Terraform) Run(v *version.Version, args []string, w wrapper) (*os.ProcessState, error) {
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

func (tf *Terraform) DownloadVersion(v *version.Version) error {
	url := fmt.Sprintf("https://releases.hashicorp.com/terraform/%s/terraform_%s_%s_%s.zip", v.String(), v.String(), runtime.GOOS, runtime.GOARCH)
	filename := fmt.Sprintf("%s/%s", tf.location, v.String())

	fmt.Printf("Getting %s...\n", url)
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("Could not download %s: %s", url, err.Error())
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("Could not download %s: %s", url, resp.Status)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	zipReader, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
	if err != nil {
		return err
	}

	unpacked := false
	for _, zipped := range zipReader.File {
		if zipped.Name != "terraform" {
			continue
		}

		t, err := zipped.Open()
		if err != nil {
			return err
		}
		defer t.Close()

		b, err := ioutil.ReadAll(t)
		if err != nil {
			return err
		}

		dest, err := os.Create(filename)
		if err != nil {
			return err
		}
		defer dest.Close()

		_, err = dest.Write(b)
		if err != nil {
			return err
		}
		dest.Sync()
		unpacked = true
	}

	if !unpacked {
		return fmt.Errorf("Could not find file `terraform` not found in downloaded zip")
	}

	if err := os.Chmod(filename, 0700); err != nil {
		return err
	}
	fmt.Println(filename)

	return nil
}
