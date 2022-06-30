package main

import (
	"github.com/polygon-io/errands-go/cmd/errands/cmd"
	"github.com/spf13/cobra"
)

func main() {
	ctl, err := cmd.NewCommand()
	cobra.CheckErr(err)

	cobra.CheckErr(ctl.Execute())
}
