/*
Copyright © 2022 DIRECT-DEV.RU <INFO@DIRECT-DEV.RU>
*/
package cmd

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	mcli_crypto "mcli/packages/mcli-crypto"
	mcli_http "mcli/packages/mcli-http"

	mcli_redis "mcli/packages/mcli-redis"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// testCmd represents the test command
var testCmd = &cobra.Command{
	Use:   "test",
	Short: "just test command for debuging",
	Long:  `useful only when DEBUG variable set to true`,
	Run:   TestRunCommand,
}

func init() {
	rootCmd.AddCommand(testCmd)
	// testCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	testCmd.Flags().StringP("function", "f", "", "function name to run")
}

func TestRunCommand(cmd *cobra.Command, args []string) {
	funcMap := map[string]func(cmd *cobra.Command, args []string){
		"RedisTestCommand": RedisTestCommand,
		"RsaReadFromFile":  RsaReadFromFile,
	}
	functionName, _ := cmd.Flags().GetString("function")

	if len(functionName) == 0 {
		functionName = "RedisTestCommand"
	}
	funcMap[functionName](cmd, args)
}

// testing functions

func RedisTestCommand(cmd *cobra.Command, args []string) {
	redisStore, err := mcli_redis.NewRedisStore("127.0.0.1:6379", "!mySuperPwd0", "userlist")
	if err != nil {
		log.Fatalln("RedisInit:", err)
	}
	var memoryUsers map[string]interface{}

	hashPassword1, _ := mcli_crypto.HashPassword("adminK@")
	hashPassword2, _ := mcli_crypto.HashPassword("user0k")

	var user1 *mcli_http.Credential = mcli_http.NewCredential("admin", hashPassword1, false, nil)
	user1.Email = "info@direct-dev.ru"
	user1.Phone = "+79059400071"
	user1.Description = "this is admin user"
	user1.Roles = []string{"admin", "user-rw"}

	var user2 *mcli_http.Credential = mcli_http.NewCredential("user@yandex.ru", hashPassword2, false, nil)
	user2.Description = "this is regular user"
	user2.Phone = "+79608755599"
	user2.Roles = []string{"user-rw"}

	memoryUsers = map[string]interface{}{user1.Username: user1, user2.Username: user2}

	// err = redisStore.SetRecordEx(user1.Username, memoryUsers[user1.Username], 20000, "userlist")
	// if err != nil {
	// 	log.Fatalln("error SetRecord User1:", err)
	// }

	// err = redisStore.SetRecord(user2.Username, memoryUsers[user2.Username])
	// if err != nil {
	// 	log.Fatalln("error SetRecord User2:", err)
	// }

	err = redisStore.SetRecordsEx(memoryUsers, 360000, "userlist")
	if err != nil {
		log.Fatalln("error SetRecords failed:", err)
	}

	getKey := "admin"
	// value := struct{ Key1 string }{Key1: "value1"}
	// err = redisStore.SetRecord(getKey, value)
	// if err != nil {
	// 	log.Fatalln("error SetRecord ", getKey, err)
	// }

	// rawRecord1, err, ok := redisStore.GetRecord(getKey)
	// if err != nil {
	// 	log.Fatalln("error GetRecord admin :", err)
	// }
	// log.Println(getKey, ok, rawRecord1)

	rawRecord1, ttl, err := redisStore.GetRecordEx(getKey)
	if err != nil {
		log.Fatalln("error GetRecord admin :", err)
	}
	log.Println(getKey, ttl, rawRecord1)

	getKey = "user"
	rawRecord2, err, ok := redisStore.GetRecord("user")
	if err != nil {
		log.Fatalln("error GetRecord user :", err)
	}
	log.Println(ok, rawRecord2)

	// var user1FromDb mcli_http.Credential = mcli_http.Credential{}
	// err = json.Unmarshal([]byte(rawRecord1), &user1FromDb)
	// if err != nil {
	// 	log.Fatalln("error Unmarshal GetRecord user :", err)
	// }
	// log.Println(user1FromDb.Username)
	// log.Println(mcli_crypto.CheckHashedPassword(user1FromDb.Password, "admin"))

	var user2FromDb mcli_http.Credential = mcli_http.Credential{}
	err = json.Unmarshal([]byte(rawRecord2), &user2FromDb)
	if err != nil {
		log.Fatalln("error Unmarshal GetRecord user :", err)
	}
	log.Println(user2FromDb.Username)
	log.Println(mcli_crypto.CheckHashedPassword(user2FromDb.Password, "userOk"))

	// fmt.Println(redisStore.RemoveRecords([]string{"admin"}, "users", "userlist"))

	mapa, err := redisStore.GetRecords("*", "user*")
	if err != nil {
		log.Fatalln("error GetRecords :", err)
	}
	fmt.Println(mapa)
}

func RsaReadFromFile(cmd *cobra.Command, args []string) {
	ENV_DEBUG := strings.ToLower(os.Getenv("DEBUG"))
	if ENV_DEBUG != "true" {
		os.Exit(1)
	}
	mcli_crypto.GetPublicKeyFromFile("./cert/mcli-cert.pem")
	fmt.Println("----------private---------------")
	mcli_crypto.GetPrivateKeyFromFile("./cert/mcli-key.pem")
}

func TestHttptemplCache(cmd *cobra.Command, args []string) {
	ENV_DEBUG := strings.ToLower(os.Getenv("DEBUG"))
	if ENV_DEBUG != "true" {
		os.Exit(1)
	}
	cache, err := mcli_http.LoadTemplatesCache("http-data/templates")
	fmt.Printf("|%+v| <%+v>\n", cache, err)
}

func RSAEncDecTest(cmd *cobra.Command, args []string) {
	ENV_DEBUG := strings.ToLower(os.Getenv("DEBUG"))
	if ENV_DEBUG != "true" {
		os.Exit(1)
	}

	// mcli_crypto.GenerateRsaCert("localhost", "./cert/localhost/", []string{"localhost", "mcli.dev"})
	ret, _ := mcli_crypto.GenRsa()
	first, second := ret[0], ret[1]
	firstPubKey := &first.PublicKey
	secondPubKey := &second.PublicKey
	fmt.Println("Private Key : ", first)
	fmt.Println("Public key ", firstPubKey)
	fmt.Println("Private Key :", second)
	fmt.Println("Public key ", secondPubKey)

	// first participant crypt message with public key of second participant
	message := []byte("Это все, что останется после меня, Это все, что возьму я с собой!")
	label := []byte("")
	hash := sha256.New()
	ciphertext, err := rsa.EncryptOAEP(
		hash,
		rand.Reader,
		secondPubKey,
		message,
		label,
	)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Printf("OAEP encrypted [%s] to \n[%x]\n", string(message), ciphertext)

	// next step - first user signs message with private key
	var opts rsa.PSSOptions
	opts.SaltLength = rsa.PSSSaltLengthAuto
	PSSmessage := message
	newhash := crypto.SHA256
	pssh := newhash.New()
	pssh.Write(PSSmessage)
	hashed := pssh.Sum(nil)
	signature, err := rsa.SignPSS(
		rand.Reader,
		first,
		newhash,
		hashed,
		&opts,
	)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Printf("PSS Signature : %x\n", signature)

	// sending to second participant [ciphertext, signature] ====>>>>>==>>>>>===>
	// second participant receives that data and first thing he'll do will be decrypt the message
	plainText, err := rsa.DecryptOAEP(
		hash,
		rand.Reader,
		second,
		ciphertext,
		label,
	)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Printf("OAEP decrypted [%x] to \n[%s]\n", ciphertext, plainText)

	// Last thing, second participant should verify the message origin to determine it is first participant:
	err = rsa.VerifyPSS(
		firstPubKey,
		newhash,
		hashed,
		signature,
		&opts,
	)
	if err != nil {
		fmt.Println("Who are U? Verify Signature failed")
		os.Exit(1)
	} else {
		fmt.Println("Verify Signature successful")
	}
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
// mcli_fs.SetFile.TarToFile("cmd", ".test-data/cmd.tar")
// mcli_fs.TarToFile("cmd", ".test-data/cmd.tar")
// mcli_fs.UntarFromFile(".test-data/cmd.tar", ".test-data")

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
