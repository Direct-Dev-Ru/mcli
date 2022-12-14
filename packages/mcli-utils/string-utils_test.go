package mcliutils

// https://itnan.ru/post.php?c=1&p=591725&ysclid=lbez5l0est293792622

// https://gianarb.it/blog/golang-mockmania-cli-command-with-cobra
// https://www.digitalocean.com/community/tutorials/how-to-write-unit-tests-in-go-using-go-test-and-the-testing-package

import (
	"testing"
)

var string_one string = "testing !string#one"

func Test_SubString(t *testing.T) {

	got := SubString(string_one, 0, 6)
	want := "testin"

	if got != want {
		t.Errorf("got %q, wanted %q", got, want)
	}

}

func Test_SubStringFind(t *testing.T) {

	got := SubStringFind(string_one, "!", "#")
	want := "string"

	if got != want {
		t.Errorf("got %q, wanted %q", got, want)
	}

}
