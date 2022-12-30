package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

func GetStringParam(param string, cmd *cobra.Command, fromConfig string) (string, error) {

	paramValue, err := cmd.Flags().GetString(param)
	isParamSet := cmd.Flags().Lookup(param).Changed
	if !isParamSet && len(fromConfig) > 0 {
		paramValue = fromConfig
	}
	return paramValue, err
}

func IsPathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
