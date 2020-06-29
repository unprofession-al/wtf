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
		Use:   "wtf",
		Short: "Wrapper for Terraform: Transparently work with multiple terraform versions",
	}
	a.Execute = rootCmd.Execute

	// exec
	execCmd := &cobra.Command{
		Use:                "exec",
		Short:              "Run correct version of terraform",
		DisableFlagParsing: true,
		Run:                a.execCmd,
	}
	rootCmd.AddCommand(execCmd)

	// configure
	configureCmd := &cobra.Command{
		Use:   "configure",
		Short: "Configure interactively",
		Run:   a.configureCmd,
	}
	rootCmd.AddCommand(configureCmd)

	// config
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Print configuration",
		Run:   a.configCmd,
	}
	rootCmd.AddCommand(configCmd)

	// install
	installCmd := &cobra.Command{
		Use:   "install",
		Short: "install a version of terraform",
		Run:   a.installCmd,
	}
	rootCmd.AddCommand(installCmd)

	// list-versions
	listVersionsCmd := &cobra.Command{
		Use:   "list-versions",
		Short: "list versions of terraform",
		Run:   a.listVersionsCmd,
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

func (a *App) configureCmd(cmd *cobra.Command, args []string) {
	k, err := NewConfiguration()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	k.Interactive()
}

func (a *App) configCmd(cmd *cobra.Command, args []string) {
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
}

func (a *App) installCmd(cmd *cobra.Command, args []string) {
	k, err := NewConfiguration()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	tf, err := NewTerraform(k.BinaryStorePath, true)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
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
		fmt.Printf("%d error(s) occurred\n", len(errs))
		os.Exit(1)
	}
}

func (a *App) listVersionsCmd(cmd *cobra.Command, args []string) {
	k, err := NewConfiguration()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	tf, err := NewTerraform(k.BinaryStorePath, true)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	installed := tf.ListInstalled()
	available, err := tf.ListAvailable()

	for _, v := range available {
		isInstalled := false
		for _, i := range installed {
			if v.Equal(i) {
				isInstalled = true
			}
		}
		out := fmt.Sprintf("%s", v)
		if isInstalled {
			out = fmt.Sprintf("%s [installed]", out)
		}
		fmt.Println(out)
	}

}

func (a *App) versionCmd(cmd *cobra.Command, args []string) {
	fmt.Println(versionInfo())
}
