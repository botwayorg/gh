package cmdutil

import (
	"net/http"

	"github.com/botwayorg/gh/context"
	"github.com/botwayorg/gh/core/config"
	"github.com/botwayorg/gh/core/ghrepo"
	"github.com/botwayorg/gh/pkg/iostreams"
)

type Browser interface {
	Browse(string) error
}

type Factory struct {
	IOStreams *iostreams.IOStreams
	Browser   Browser

	HttpClient func() (*http.Client, error)
	BaseRepo   func() (ghrepo.Interface, error)
	Remotes    func() (context.Remotes, error)
	Config     func() (config.Config, error)
	Branch     func() (string, error)

	Executable string
}
