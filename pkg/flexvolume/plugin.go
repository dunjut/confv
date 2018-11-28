package flexvolume

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/spf13/cobra"

	"github.com/dunjut/confv/pkg/config"
)

func AddCobraCommands(rootCmd *cobra.Command) {
	// <driver executable> init
	rootCmd.AddCommand(&cobra.Command{
		Use:   "init",
		Short: "Initialize the driver",
		Args:  cobra.ExactArgs(0),
		Run:   initCmd,
	})

	// <driver executable> mount <mount dir> <json options>
	rootCmd.AddCommand(&cobra.Command{
		Use:   "mount",
		Short: "Mount the volume at the mount dir",
		Args:  cobra.ExactArgs(2),
		Run:   mountCmd,
	})

	// <driver executable> unmount <mount dir>
	rootCmd.AddCommand(&cobra.Command{
		Use:   "unmount",
		Short: "Unmount the volume",
		Args:  cobra.ExactArgs(1),
		Run:   unmountCmd,
	})
}

func initCmd(cmd *cobra.Command, args []string) {
	msg := `{"status": "Success", "capabilities": {"attach": false}}`
	fmt.Print(msg)
}

func mountCmd(cmd *cobra.Command, args []string) {
	var (
		mountDir   = args[0]
		rawOptions = args[1]
	)
	options, err := config.DecodeOptions(rawOptions)
	if err != nil {
		fail(err)
	}
	cfgBytes, err := config.RenderConfig(options)
	if err != nil {
		fail(err)
	}
	if err = ensureMountDir(mountDir); err != nil {
		fail(err)
	}
	fname := path.Join(mountDir, options.TargetFileName)
	if err = ioutil.WriteFile(fname, cfgBytes, 0666); err != nil {
		fail(err)
	}
	succeed()
}

func unmountCmd(cmd *cobra.Command, args []string) {
	if err := os.RemoveAll(args[0]); err != nil {
		fail(err)
	}
	succeed()
}

// ensureMountDir ensures path (directory) exists
func ensureMountDir(path string) error {
	if err := os.MkdirAll(path, 0777); err != nil {
		if os.IsExist(err) {
			return nil
		}
		return err
	}
	return nil
}

// succeed prints success info and exit
func succeed() {
	fmt.Print(`{"status": "Success"}`)
	os.Exit(0)
}

// fail prints failure info and exit
func fail(err error) {
	fmt.Printf(`{"status": "Failure", "message": "%v"}`, err)
	os.Exit(1)
}
