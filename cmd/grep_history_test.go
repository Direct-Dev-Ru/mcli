package cmd

// https://gianarb.it/blog/golang-mockmania-cli-command-with-cobra
// https://www.digitalocean.com/community/tutorials/how-to-write-unit-tests-in-go-using-go-test-and-the-testing-package

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/spf13/cobra"
)

var test_filter string

func NewHistoryRunCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "grep history",
		Short: "Writes out history of shell commands with filtering",
		Run:   historyRun,
	}
	cmd.Flags().StringVar(&test_filter, "filter", "", "This is filter string")
	return cmd
}

func Test_HistoryRun(t *testing.T) {

	// got := Add(4, 6)
	// want := 10

	// if got != want {
	// 	t.Errorf("got %q, wanted %q", got, want)
	// }

	cmd := NewHistoryRunCmd()
	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	cmd.SetArgs([]string{"--filter", "merge"})
	cmd.Execute()
	out, err := ioutil.ReadAll(b)
	if err != nil {
		t.Fatal(err)
	}
	if len(string(out)) == 0 {
		t.Fatalf("expected \"%s\" got \"%s\"", "", string(out))
	}
}
