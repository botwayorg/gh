package shared

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/MakeNowJust/heredoc"
	"github.com/botwayorg/gh/api"
	"github.com/botwayorg/gh/core/authflow"
	"github.com/botwayorg/gh/core/ghinstance"
	"github.com/botwayorg/gh/pkg/cmd/auth/shared/auth_gh_api"
	"github.com/botwayorg/gh/pkg/cmd/auth/shared/auth_token"
	"github.com/botwayorg/gh/pkg/cmd/auth/shared/git_protocol"
	"github.com/botwayorg/gh/pkg/cmd/auth/shared/upload_ssh"
	"github.com/botwayorg/gh/pkg/iostreams"
)

type iconfig interface {
	Get(string, string) (string, error)
	Set(string, string, string) error
	Write() error
}

type LoginOptions struct {
	IO          *iostreams.IOStreams
	Config      iconfig
	HTTPClient  *http.Client
	Hostname    string
	Interactive bool
	Web         bool
	Scopes      []string
	Executable  string

	sshContext sshContext
}

func Login(opts *LoginOptions) error {
	cfg := opts.Config
	hostname := opts.Hostname
	httpClient := opts.HTTPClient
	cs := opts.IO.ColorScheme()

	var gitProtocol string

	if opts.Interactive {
		var proto string
		var err error

		proto, err = git_protocol.GitProtocol()

		if err != nil {
			return fmt.Errorf("could not prompt: %w", err)
		}

		gitProtocol = strings.ToLower(proto)
	}

	var additionalScopes []string

	credentialFlow := &GitCredentialFlow{Executable: opts.Executable}

	if opts.Interactive && gitProtocol == "https" {
		if err := credentialFlow.Prompt(hostname); err != nil {
			return err
		}

		additionalScopes = append(additionalScopes, credentialFlow.Scopes()...)
	}

	var keyToUpload string

	if opts.Interactive && gitProtocol == "ssh" {
		pubKeys, err := opts.sshContext.localPublicKeys()
		if err != nil {
			return err
		}

		if len(pubKeys) > 0 {
			var keyChoice int
			var err error

			keyChoice, err = upload_ssh.UploadSSH(pubKeys)

			if err != nil {
				return fmt.Errorf("could not prompt: %w", err)
			}

			if keyChoice < len(pubKeys) {
				keyToUpload = pubKeys[keyChoice]
			}
		} else {
			var err error
			keyToUpload, err = opts.sshContext.generateSSHKey()
			if err != nil {
				return err
			}
		}
	}

	if keyToUpload != "" {
		additionalScopes = append(additionalScopes, "admin:public_key")
	}

	var authMode int
	if opts.Web {
		authMode = 0
	} else {
		var err error

		authMode, err = auth_gh_api.AuthGitHubAPI()

		if err != nil {
			return fmt.Errorf("could not prompt: %w", err)
		}
	}

	var authToken string
	userValidated := false

	if authMode == 0 {
		var err error
		authToken, err = authflow.AuthFlowWithConfig(cfg, opts.IO, hostname, "", append(opts.Scopes, additionalScopes...))
		if err != nil {
			return fmt.Errorf("failed to authenticate via web browser: %w", err)
		}

		userValidated = true
	} else {
		var err error

		minimumScopes := append([]string{"repo", "read:org"}, additionalScopes...)
		fmt.Fprint(opts.IO.ErrOut, heredoc.Docf(`
			Tip: you can generate a Personal Access Token here https://%s/settings/tokens
			The minimum required scopes are %s.
		`, hostname, scopesSentence(minimumScopes, ghinstance.IsEnterprise(hostname))))

		authToken, err = auth_token.AuthToken()

		if err != nil {
			return fmt.Errorf("could not prompt: %w", err)
		}

		if err := HasMinimumScopes(httpClient, hostname, authToken); err != nil {
			return fmt.Errorf("error validating token: %w", err)
		}

		if err := cfg.Set(hostname, "oauth_token", authToken); err != nil {
			return err
		}
	}

	var username string
	if userValidated {
		username, _ = cfg.Get(hostname, "user")
	} else {
		apiClient := api.NewClientFromHTTP(httpClient)
		var err error
		username, err = api.CurrentLoginName(apiClient, hostname)

		if err != nil {
			return fmt.Errorf("error using api: %w", err)
		}

		err = cfg.Set(hostname, "user", username)
		if err != nil {
			return err
		}
	}

	if gitProtocol != "" {
		fmt.Fprintf(opts.IO.ErrOut, "- botway gh-config set --host %s git_protocol %s\n", hostname, gitProtocol)
		err := cfg.Set(hostname, "git_protocol", gitProtocol)
		if err != nil {
			return err
		}

		fmt.Fprintf(opts.IO.ErrOut, "%s Configured git protocol\n", cs.SuccessIcon())
	}

	err := cfg.Write()
	if err != nil {
		return err
	}

	if credentialFlow.ShouldSetup() {
		err := credentialFlow.Setup(hostname, username, authToken)
		if err != nil {
			return err
		}
	}

	if keyToUpload != "" {
		err := sshKeyUpload(httpClient, hostname, keyToUpload)
		if err != nil {
			return err
		}

		fmt.Fprintf(opts.IO.ErrOut, "%s Uploaded the SSH key to your GitHub account: %s\n", cs.SuccessIcon(), cs.Bold(keyToUpload))
	}

	fmt.Fprintf(opts.IO.ErrOut, "%s Logged in as %s\n", cs.SuccessIcon(), cs.Bold(username))
	return nil
}

func scopesSentence(scopes []string, isEnterprise bool) string {
	quoted := make([]string, len(scopes))
	for i, s := range scopes {
		quoted[i] = fmt.Sprintf("'%s'", s)
		if s == "workflow" && isEnterprise {
			// remove when GHE 2.x reaches EOL
			quoted[i] += " (GHE 3.0+)"
		}
	}

	return strings.Join(quoted, ", ")
}
