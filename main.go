package main

import (
	"fmt"
	"os"

	"github.com/hashicorp/packer-plugin-sdk/plugin"
	"github.com/hashicorp/packer-plugin-sdk/version"
)

func main() {
	ps := plugin.NewSet()
	ps.RegisterProvisioner(plugin.DEFAULT_NAME, &S3Provisioner{})
	ps.SetVersion(version.NewPluginVersion("1.0.0", "", ""))
	err := ps.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
