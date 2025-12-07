package main

import (
	"fmt"
	"os"

	ver "github.com/hashicorp/go-version"
	"github.com/spf13/cobra"
)

type App struct {
	// entry point
	Execute func() error
}

func NewApp() *App {
	a := &App{}

	// root
	rootCmd := &cobra.Command{
		Use:           "wtf",
		Short:         "Wrapper for Terraform: Transparently work with multiple terraform versions",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	a.Execute = func() error {
		if err := rootCmd.Execute(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return nil
	}

	// exec
	execCmd := &cobra.Command{
		Use:                "exec",
		Short:              "Run correct version of terraform",
		DisableFlagParsing: true,
		Run:                a.execCmd,
	}
	rootCmd.AddCommand(execCmd)

	// install
	installCmd := &cobra.Command{
		Use:   "install",
		Short: "install a version of terraform",
		RunE:  a.installCmd,
	}
	rootCmd.AddCommand(installCmd)

	// list-versions
	listVersionsCmd := &cobra.Command{
		Use:   "list-versions",
		Short: "list versions of terraform",
		RunE:  a.listVersionsCmd,
	}
	rootCmd.AddCommand(listVersionsCmd)

	// version
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print version info",
		Run:   a.versionCmd,
	}
	rootCmd.AddCommand(versionCmd)

	return a
}

func (a *App) execCmd(cmd *cobra.Command, args []string) {
	runTerraform(args, true)
}

func (a *App) installCmd(cmd *cobra.Command, args []string) error {
	k, err := NewConfiguration()
	if err != nil {
		return err
	}

	tf, err := NewTerraform(k.BinaryStorePath, true)
	if err != nil {
		return err
	}

	errs := []error{}

	for _, v := range args {
		fmt.Printf("Processing %s...\n", v)
		this, err := ver.NewVersion(v)
		if err != nil {
			err = fmt.Errorf("version string '%s' could not be parsed: %s", v, err.Error())
			fmt.Println(err.Error())
			errs = append(errs, err)
			continue
		}

		filepath, err := tf.DownloadVersion(this)
		if err != nil {
			err = fmt.Errorf("version '%s' could not be downloaded: %s", v, err.Error())
			fmt.Println(err.Error())
			errs = append(errs, err)
			continue
		} else {
			fmt.Printf("version '%s' installed at '%s'\n", v, filepath)
		}

	}

	if len(errs) > 0 {
		return fmt.Errorf("%d error(s) occurred", len(errs))
	}
	return nil
}

func (a *App) listVersionsCmd(cmd *cobra.Command, args []string) error {
	k, err := NewConfiguration()
	if err != nil {
		return err
	}

	tf, err := NewTerraform(k.BinaryStorePath, true)
	if err != nil {
		return err
	}

	installed := tf.ListInstalled()
	available, err := tf.ListAvailable()
	if err != nil {
		return err
	}

	for _, v := range available {
		isInstalled := false
		for _, i := range installed {
			if v.Equal(i) {
				isInstalled = true
			}
		}
		out := v.String()
		if isInstalled {
			out = fmt.Sprintf("%s [installed]", out)
		}
		fmt.Println(out)
	}
	return nil
}

func (a *App) versionCmd(cmd *cobra.Command, args []string) {
	fmt.Println(VersionInfo())
}
