package login

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/MakeNowJust/heredoc"
	"github.com/botwayorg/gh/core/config"
	"github.com/botwayorg/gh/core/ghinstance"
	"github.com/botwayorg/gh/pkg/cmd/auth/login/host"
	re_auth "github.com/botwayorg/gh/pkg/cmd/auth/login/re-auth"
	"github.com/botwayorg/gh/pkg/cmd/auth/shared"
	"github.com/botwayorg/gh/pkg/cmdutil"
	"github.com/botwayorg/gh/pkg/iostreams"
	"github.com/spf13/cobra"
)

type LoginOptions struct {
	IO         *iostreams.IOStreams
	Config     func() (config.Config, error)
	HttpClient func() (*http.Client, error)

	MainExecutable string

	Interactive bool

	Hostname string
	Scopes   []string
	Token    string
	Web      bool
}

func NewCmdLogin(f *cmdutil.Factory, runF func(*LoginOptions) error) *cobra.Command {
	opts := &LoginOptions{
		IO:         f.IOStreams,
		Config:     f.Config,
		HttpClient: f.HttpClient,

		MainExecutable: f.Executable,
	}

	var tokenStdin bool

	cmd := &cobra.Command{
		Use:   "login",
		Args:  cobra.ExactArgs(0),
		Short: "Authenticate with a GitHub host.",
		Long: heredoc.Docf(`
			Authenticate with a GitHub host.

			The default authentication mode is a web-based browser flow.

			Alternatively, pass in a token on standard input by using %[1]s--with-token%[1]s.
			The minimum required scopes for the token are: "repo", "read:org".

			The --scopes flag accepts a comma separated list of scopes you want your botway credentials to have. If
			absent, this command ensures that botway has access to a minimum set of scopes.
		`, "`"),
		Example: heredoc.Doc(`
			# start interactive setup
			botway github login

			# authenticate against github.com by reading the token from a file
			botway github login --with-token < mytoken.txt

			# authenticate with a specific GitHub Enterprise Server instance
			botway github login --hostname enterprise.internal
		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !opts.IO.CanPrompt() && !(tokenStdin || opts.Web) {
				return &cmdutil.FlagError{Err: errors.New("--web or --with-token required when not running interactively")}
			}

			if tokenStdin && opts.Web {
				return &cmdutil.FlagError{Err: errors.New("specify only one of --web or --with-token")}
			}

			if tokenStdin {
				defer opts.IO.In.Close()

				token, err := ioutil.ReadAll(opts.IO.In)

				if err != nil {
					return fmt.Errorf("failed to read token from STDIN: %w", err)
				}

				opts.Token = strings.TrimSpace(string(token))
			}

			if opts.IO.CanPrompt() && opts.Token == "" && !opts.Web {
				opts.Interactive = true
			}

			if cmd.Flags().Changed("hostname") {
				if err := ghinstance.HostnameValidator(opts.Hostname); err != nil {
					return &cmdutil.FlagError{Err: fmt.Errorf("error parsing --hostname: %w", err)}
				}
			}

			if !opts.Interactive {
				if opts.Hostname == "" {
					opts.Hostname = ghinstance.Default()
				}
			}

			if runF != nil {
				return runF(opts)
			}

			return loginRun(opts)
		},
	}

	cmd.Flags().StringVarP(&opts.Hostname, "hostname", "", "", "The hostname of the GitHub instance to authenticate with")
	cmd.Flags().StringSliceVarP(&opts.Scopes, "scopes", "s", nil, "Additional authentication scopes for botway to have")
	cmd.Flags().BoolVar(&tokenStdin, "with-token", false, "Read token from standard input")
	cmd.Flags().BoolVarP(&opts.Web, "web", "w", false, "Open a browser to authenticate")

	return cmd
}

func loginRun(opts *LoginOptions) error {
	cfg, err := opts.Config()

	if err != nil {
		return err
	}

	hostname := opts.Hostname

	if hostname == "" {
		if opts.Interactive {
			var err error

			hostname, err = host.Host()

			if err != nil {
				return err
			}
		} else {
			return errors.New("must specify --hostname")
		}
	}

	if err := cfg.CheckWriteable(hostname, "oauth_token"); err != nil {
		var roErr *config.ReadOnlyEnvError

		if errors.As(err, &roErr) {
			fmt.Fprintf(opts.IO.ErrOut, "The value of the %s environment variable is being used for authentication.\n", roErr.Variable)
			fmt.Fprint(opts.IO.ErrOut, "To have botway store credentials instead, first clear the value from the environment.\n")
			return cmdutil.SilentError
		}

		return err
	}

	httpClient, err := opts.HttpClient()

	if err != nil {
		return err
	}

	if opts.Token != "" {
		err := cfg.Set(hostname, "oauth_token", opts.Token)

		if err != nil {
			return err
		}

		if err := shared.HasMinimumScopes(httpClient, hostname, opts.Token); err != nil {
			return fmt.Errorf("error validating token: %w", err)
		}

		return cfg.Write()
	}

	existingToken, _ := cfg.Get(hostname, "oauth_token")

	if existingToken != "" && opts.Interactive {
		if err := shared.HasMinimumScopes(httpClient, hostname, existingToken); err == nil {
			var keepGoing bool

			keepGoing, err = re_auth.ReAuth(hostname)

			if err != nil {
				return fmt.Errorf("could not prompt: %w", err)
			}

			if !keepGoing {
				os.Exit(0)
			}
		}
	}

	return shared.Login(&shared.LoginOptions{
		IO:          opts.IO,
		Config:      cfg,
		HTTPClient:  httpClient,
		Hostname:    hostname,
		Interactive: opts.Interactive,
		Web:         opts.Web,
		Scopes:      opts.Scopes,
		Executable:  opts.MainExecutable,
	})
}
