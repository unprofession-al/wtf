package main

import (
	"bytes"
	"os"
	"strings"
	"text/template"
)

type wrapper struct {
	ScriptTemplate string `yaml:"script_template"`
	tmpfile        *os.File
}

func (w *wrapper) Wrap(command string, args []string, verbose bool) (string, []string, error) {
	if w.ScriptTemplate == "" {
		return command, args, nil
	}
	c := strings.Join(append([]string{command}, args...), " ")

	data := struct {
		Command string
		Verbose bool
	}{
		Command: c,
		Verbose: verbose,
	}

	var out bytes.Buffer
	tmpl, err := template.New("wrapper").Parse(w.ScriptTemplate)
	if err != nil {
		return command, args, err
	}

	err = tmpl.Execute(&out, data)
	if err != nil {
		return command, args, err
	}

	w.tmpfile, err = os.CreateTemp("", "wrapped.terraform.*.wtf")
	if err != nil {
		return command, args, err
	}

	if _, err := w.tmpfile.Write(out.Bytes()); err != nil {
		w.tmpfile.Close()
		return command, args, err
	}
	if err := w.tmpfile.Close(); err != nil {
		return command, args, err
	}

	err = os.Chmod(w.tmpfile.Name(), 0700)
	if err != nil {
		return command, args, err
	}
	return w.tmpfile.Name(), []string{}, nil
}

func (w *wrapper) Cleanup() error {
	if w.tmpfile == nil {
		return nil
	}
	return os.Remove(w.tmpfile.Name())
}
