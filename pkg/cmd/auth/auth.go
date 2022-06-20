package auth

import (
	authGetUsernameCmd "github.com/botwayorg/gh/pkg/cmd/auth/get-username"
	gitCredentialCmd "github.com/botwayorg/gh/pkg/cmd/auth/gitcredential"
	authLoginCmd "github.com/botwayorg/gh/pkg/cmd/auth/login"
	authLogoutCmd "github.com/botwayorg/gh/pkg/cmd/auth/logout"
	authRefreshCmd "github.com/botwayorg/gh/pkg/cmd/auth/refresh"
	authStatusCmd "github.com/botwayorg/gh/pkg/cmd/auth/status"
	"github.com/botwayorg/gh/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func NewCmdAuth(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth <command>",
		Short: "Login, logout, and refresh your authentication with github.",
		Long:  `Manage botway's authentication state.`,
	}

	cmdutil.DisableAuthCheck(cmd)

	cmd.AddCommand(authGetUsernameCmd.GetUsername())
	cmd.AddCommand(authLoginCmd.NewCmdLogin(f, nil))
	cmd.AddCommand(authLogoutCmd.NewCmdLogout(f, nil))
	cmd.AddCommand(authStatusCmd.NewCmdStatus(f, nil))
	cmd.AddCommand(authRefreshCmd.NewCmdRefresh(f, nil))
	cmd.AddCommand(gitCredentialCmd.NewCmdCredential(f, nil))

	return cmd
}
