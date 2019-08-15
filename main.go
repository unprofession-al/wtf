package main

import (
	"fmt"
	"os"

	"github.com/hashicorp/go-version"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(execCmd)
	rootCmd.AddCommand(configureCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(installCmd)
}

var rootCmd = &cobra.Command{
	Use:   "wtf",
	Short: "Wrapper for Terraform: Transparently work with multiple terraform versions",
}

var execCmd = &cobra.Command{
	Use:   "exec",
	Short: "run correct version of terraform",
	Run: func(cmd *cobra.Command, args []string) {
		k, err := NewConfiguration()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		c, err := readConstraint(k.DetectSyntax)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		tf, err := NewTerraform(k.BinaryStorePath)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Printf("Version constraint: %s\n", c.String())

		latest, err := tf.FindLatest(c)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Printf("Version used: %s\n\n", latest.String())

		s, err := tf.Run(latest, args, k.Wrapper)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		os.Exit(s.ExitCode())
	},
}

var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "configure interactively",
	Run: func(cmd *cobra.Command, args []string) {
		k, err := NewConfiguration()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		k.Interactive()
	},
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "print configuration",
	Run: func(cmd *cobra.Command, args []string) {
		k, err := NewConfiguration()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		out, err := k.ToYAML()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Println(string(out))
	},
}

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "install a version of terraform",
	Run: func(cmd *cobra.Command, args []string) {
		k, err := NewConfiguration()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		tf, err := NewTerraform(k.BinaryStorePath)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		errs := []error{}

		for _, ver := range args {
			fmt.Printf("Processing %s...\n", ver)
			v, err := version.NewVersion(ver)
			if err != nil {
				err = fmt.Errorf("Version string could not be parsed: %s", err.Error())
				fmt.Println(err.Error())
				errs = append(errs, err)
				continue
			}

			err = tf.DownloadVersion(v)
			if err != nil {
				err = fmt.Errorf("Version could not be downloaded: %s", err.Error())
				fmt.Println(err.Error())
				errs = append(errs, err)
				continue
			}

		}

		if len(errs) > 0 {
			fmt.Printf("%d error(s) occured\n", len(errs))
			os.Exit(1)
		}
	},
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}