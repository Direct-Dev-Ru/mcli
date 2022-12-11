package cmd

// https://itnan.ru/post.php?c=1&p=591725&ysclid=lbez5l0est293792622

// https://gianarb.it/blog/golang-mockmania-cli-command-with-cobra
// https://www.digitalocean.com/community/tutorials/how-to-write-unit-tests-in-go-using-go-test-and-the-testing-package

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

var test_filter string

func NewHistoryRunCmd(filter string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "grep history",
		Short: "Writes out history of shell commands with filtering",
		Run: func(cmd *cobra.Command, args []string) {
			historyRun(cmd, args)
		},
	}
	cmd.Flags().StringVar(&test_filter, "filter", filter, "This is filter string")
	return cmd
}

func Test_HistoryRun(t *testing.T) {

	// got := Add(4, 6)
	// want := 10

	// if got != want {
	// 	t.Errorf("got %q, wanted %q", got, want)
	// }

	cmd := NewHistoryRunCmd("merge")
	// b := bytes.NewBufferString("")
	var b bytes.Buffer
	cmd.SetOut(&b)
	// cmd.SetArgs([]string{"--filter", "merge"})
	cmd.Execute()
	out, err := io.ReadAll(&b)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("------")
	fmt.Println(len(string(out)), string(out))
	fmt.Println("------")
	if len(string(out)) == 0 {
		t.Fatalf("expected \"%s\" got \"%s\"", "", string(out))
	}
}

func join(elems []string, sep string) string {
	var res string
	for i, s := range elems {
		res += s
		if i < len(elems)-1 {
			res += sep
		}
	}
	return res
}

var strs = []string{
	"string a",
	"string b",
	"string c",
	"string d",
	"string e",
}

// go test -benchmem -run=^$ -bench ^Benchmark  mcli/cmd
func BenchmarkMyJoin(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = join(strs, ";")
	}
}
func BenchmarkStringsJoin(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = strings.Join(strs, ";")
	}
}
