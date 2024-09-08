package main

import (
	"fmt"
	"github.com/hashicorp/packer-plugin-sdk/plugin"
	"github.com/hashicorp/packer-plugin-sdk/version"
	"os"
)

func main() {
	ps := plugin.NewSet()
	ps.RegisterProvisioner(plugin.DEFAULT_NAME, &S3Provisioner{})
	ps.SetVersion(version.NewPluginVersion("0.1.0", "dev", ""))
	err := ps.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
