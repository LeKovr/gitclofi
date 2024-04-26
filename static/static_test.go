package static_test

import (
	"testing"

	"github.com/LeKovr/gitclofi/static"
	ass "github.com/alecthomas/assert/v2"
)

func TestNewEmbed(t *testing.T) {
	got, err := static.New("tmpl/")
	ass.NotZero(t, got, "FS not nil")
	ass.NoError(t, err, "New success")
	f, err := got.Open("header.gohtml")
	errC := f.Close()
	ass.NoError(t, err, "Open success")
	ass.NoError(t, errC, "Close success")
}

