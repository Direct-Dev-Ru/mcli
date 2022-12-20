/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	mcli_fs "mcli/packages/mcli-filesystem"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

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

		// var paths string = ".mcli.yaml  http-data/http-params cmd    http-data"
		// content, _ := mcli_fs.GetFile.GetContent("/home/su/projects/golang/cobra-cli-example/run.cmd")

		// gzData, err := mcli_fs.ZipData(content)
		// if err != nil {
		// 	Elogger.Error().Msg(err.Error())
		// }
		// str := base64.StdEncoding.EncodeToString(gzData)
		// str := string(gzData)
		// fmt.Println(len(gzData))
		// fmt.Println(gzData)
		// mcli_fs.SetFile.SetContent("/home/su/projects/golang/cobra-cli-example/.test-data/grep1.go.gz", gzData)
		// mcli_fs.SetFile.SetZipContent("/home/su/projects/golang/cobra-cli-example/.test-data/grep2.go.gz", content)
		// unzipcontent, err := mcli_fs.GetFile.GetUnZipContent("/home/su/projects/golang/cobra-cli-example/.test-data/grep2.go.gz")

		// if err := mcli_fs.SetFile.ZipFile(".mcli.yaml", ".test-data/.mcli.yaml.gz"); err != nil {
		// 	fmt.Println(err)
		// }
		// if err := mcli_fs.GetFile.UnZipFile(".test-data/.mcli.yaml.gz", ""); err != nil {
		// 	fmt.Println("unzip error :", err)
		// }

		// fmt.Println(string(unzipcontent), err)
		mcli_fs.SetFile.TarToFile("cmd", ".test-data/cmd.tar")
		// mcli_fs.TarToFile("cmd", ".test-data/cmd.tar")
		mcli_fs.UntarFromFile(".test-data/cmd.tar", ".test-data")

		// // data, _ := base64.StdEncoding.DecodeString(str)
		// data := []byte(str)
		// unGzData, err := mcli_fs.UnZipData(data)
		// if err != nil {
		// 	Elogger.Error().Msg(err.Error())
		// }
		// fmt.Println(len(string(unGzData)))
		// fmt.Println(string(unGzData))

		// var re = regexp.MustCompile(`(?m)^([A-Z0-9А-Я]+:)(.+)`)
		// var re = regexp.MustCompile(`^?(?P<col>[A-Z0-9А-Я]+):(?P<val>.+)`)
		// var str = `IMAGE:django`
		// fmt.Printf("%#v\n", re.FindStringSubmatch(str))
		// fmt.Printf("%#v\n", re.SubexpNames())
		// fmt.Printf("%#v\n", re.SubexpIndex("col"))
		// fmt.Printf("%#v\n", re.SubexpIndex("val"))
		// fmt.Printf("%#v\n", re.FindStringSubmatchIndex(str))

		// for i, match := range re.FindAllString(str, -1) {

		// 	fmt.Println(match, "found at index", i)
		// }

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
