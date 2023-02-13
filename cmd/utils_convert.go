/*
Copyright © 2023 DIRECT-DEV.RU <INFO@DIRECT-DEV.RU>
*/
package cmd

import (
	mcli_utils "mcli/packages/mcli-utils"
	"path"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

func convertRunFunc(cmd *cobra.Command, args []string) {

	sourcePath, _ := cmd.Flags().GetString("source-path")
	// isSourcePathSet := cmd.Flags().Lookup("source-path").Changed
	destPath, _ := cmd.Flags().GetString("dest-path")
	// isDestPathSet := cmd.Flags().Lookup("dest-path").Changed
	convertType, _ := cmd.Flags().GetString("convert-type")
	// isConvertTypeSet := cmd.Flags().Lookup("convert-type").Changed
	convertType = strings.ToUpper(convertType)

	_, sourceType, err := IsPathExistsAndCreate(sourcePath, false)
	if err != nil {
		Elogger.Fatal().Msgf("someting goes wrong %v", err)
	}

	_, destType, err := IsPathExistsAndCreate(destPath, true)

	if err != nil {
		Elogger.Fatal().Msgf("something goes wrong while creating dest directory: %v", err)
	}

	switch convertType {
	case "TO_BSON_FROM_JSON":
		if sourceType == "directory" {
			Elogger.Fatal().Msg("converter TO_BSON_FROM_JSON don't support directory source - specify path to file")
		}
		if destType == "directory" {
			fileName := filepath.Base(sourcePath)
			ext := filepath.Ext(fileName)
			fileNameWithoutExt := fileName[:len(fileName)-len(ext)]
			destPath = path.Join(destPath, fileNameWithoutExt+".bson")
		}
		err := mcli_utils.ConvertJsonToBson(sourcePath, destPath)
		if err != nil {
			Elogger.Fatal().Msgf("error occured in converter TO_BSON_FROM_JSON: %v", err)
		}
	case "TO_JSON_FROM_BSON":
		if sourceType == "directory" {
			Elogger.Fatal().Msg("converter TO_JSON_FROM_BSON don't support directory source - specify path to file")
		}
		if destType == "directory" {
			fileName := filepath.Base(sourcePath)
			ext := filepath.Ext(fileName)
			fileNameWithoutExt := fileName[:len(fileName)-len(ext)]
			destPath = path.Join(destPath, fileNameWithoutExt+".json")
		}
		err := mcli_utils.ConvertBsonToJson(sourcePath, destPath)
		if err != nil {
			Elogger.Fatal().Msgf("error occured in converter TO_JSON_FROM_BSON: %v", err)
		}

	default:
		Elogger.Fatal().Msg("converter with given type doesn't supported yet")

	}

}

// convertCmd represents the convert command
var convertCmd = &cobra.Command{
	Use:   "convert",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: convertRunFunc,
}

func init() {
	utilsCmd.AddCommand(convertCmd)

	convertCmd.Flags().StringP("convert-type", "t", "toBsonFromJson", "Type of conversion (toBsonFromJson, ...)")
	convertCmd.Flags().StringP("source-path", "s", "", "path to source file or dir")
	convertCmd.Flags().StringP("dest-path", "d", "", "path to destination file or dir")

}
