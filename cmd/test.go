/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

func listDirByWalk(path string) {
	filepath.Walk(path, func(wPath string, info os.FileInfo, err error) error {

		// Обход директории без вывода
		if wPath == path {
			return nil
		}

		// Если данный путь является директорией, то останавливаем рекурсивный обход
		// и возвращаем название папки
		if info.IsDir() {
			fmt.Printf("[%s]\n", wPath)
			// return filepath.SkipDir
		}

		// Выводится название файла
		if wPath != path && !info.IsDir() {
			fmt.Println(path + "/" + wPath)
		}
		return nil
	})
}

// testCmd represents the test command
var testCmd = &cobra.Command{
	Use:   "test",
	Short: "just test command for debuging",
	Long:  `useful only when DEBUG variable set to true`,
	Run: func(cmd *cobra.Command, args []string) {
		ENV_DEBUG := strings.ToLower(os.Getenv("DEBUG"))
		if ENV_DEBUG != "true" {
			os.Exit(1)
		}

		// var re = regexp.MustCompile(`(?m)^([A-Z0-9А-Я]+:)(.+)`)
		var re = regexp.MustCompile(`^?(?P<col>[A-Z0-9А-Я]+):(?P<val>.+)`)
		var str = `IMAGE:django`
		fmt.Printf("%#v\n", re.FindStringSubmatch(str))
		fmt.Printf("%#v\n", re.SubexpNames())
		fmt.Printf("%#v\n", re.SubexpIndex("col"))
		fmt.Printf("%#v\n", re.SubexpIndex("val"))
		fmt.Printf("%#v\n", re.FindStringSubmatchIndex(str))

		for i, match := range re.FindAllString(str, -1) {

			fmt.Println(match, "found at index", i)
		}

		// tests := "test str*ing"
		// fmt.Println(mcli_utils.SubString(tests, 5, 22222222), len(mcli_utils.SubString(tests, 5, 22222222)))
		// fmt.Println(mcli_utils.SubStringFind(tests, "t", "t"), len(mcli_utils.SubStringFind(tests, " ", "*")))

		// fmt.Println("List by Walk")
		// listDirByWalk("/home")

		// refString := "{{$RootPath$}}/pwdgen/{{$HOME}}/freqrnc2011.csv"
		// templateRegExp := regexp.MustCompile(`{{\$.+?}}`)

		// all := templateRegExp.FindAllString(refString, -1)
		// fmt.Println("All: ")
		// for _, val := range all {
		// 	fmt.Println(val)
		// }

		// cmdSh := &exec.Cmd{
		// 	Path:   "./script.sh",
		// 	Args:   []string{"./script.sh"},
		// 	Stdout: os.Stdout,
		// 	Stderr: os.Stderr,
		// }

		// if err := cmdSh.Run(); err != nil {
		// 	fmt.Println("Error:", err)
		// }

		// writeContent := strings.Repeat("Blya-ha Muha ", 100)

		// n, err := mcli_fs.SetFile.SetContent("file.tmp", []byte(writeContent))
		// if err != nil {
		// 	Elogger.Fatal().Msg(err.Error())
		// }
		// Ilogger.Info().Msg(fmt.Sprintf("\nwritten:\n%d bytes\n", n))

		// readContent, err := mcli_fs.GetFile.GetContent("file.tmp")

		// if err != nil {
		// 	Elogger.Fatal().Msg(err.Error())
		// }
		// Ilogger.Info().Msg(fmt.Sprintf("\ncontent:\n%s\n", readContent))
	},
}

func init() {
	rootCmd.AddCommand(testCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// testCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// testCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
