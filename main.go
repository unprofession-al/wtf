package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	bin := os.Args[0]
	args := []string{}
	if len(os.Args) > 1 {
		args = os.Args[1:]
	}

	if filepath.Base(bin) == "terraform" {
		runTerraform(args, false)
		os.Exit(0)
	}

	if err := NewApp().Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func runTerraform(args []string, verbose bool) {
	k, err := NewConfiguration()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	c, err := readConstraint()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	tf, err := NewTerraform(k.BinaryStorePath, verbose)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if verbose {
		fmt.Printf("Version constraint: %s\n", c.String())
	}

	latest, err := tf.FindLatest(c)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if verbose {
		fmt.Printf("Version used: %s\n\n", latest.String())
	}

	s, err := tf.Run(latest, args, k.Wrapper)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	os.Exit(s.ExitCode())
}
