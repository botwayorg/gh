package repo

import (
	"github.com/MakeNowJust/heredoc"
	repoBrowseCmd "github.com/botwayorg/gh/pkg/cmd/gh-repo/browse"
	repoCloneCmd "github.com/botwayorg/gh/pkg/cmd/gh-repo/clone"
	repoCreateCmd "github.com/botwayorg/gh/pkg/cmd/gh-repo/create"
	creditsCmd "github.com/botwayorg/gh/pkg/cmd/gh-repo/credits"
	repoForkCmd "github.com/botwayorg/gh/pkg/cmd/gh-repo/fork"
	gardenCmd "github.com/botwayorg/gh/pkg/cmd/gh-repo/garden"
	repoListCmd "github.com/botwayorg/gh/pkg/cmd/gh-repo/list"
	repoSyncCmd "github.com/botwayorg/gh/pkg/cmd/gh-repo/sync"
	repoViewCmd "github.com/botwayorg/gh/pkg/cmd/gh-repo/view"
	"github.com/botwayorg/gh/pkg/cmdutil"
	"github.com/botwayorg/gh/utils"
	// git_config "github.com/botwayorg/git"
	"github.com/spf13/cobra"

	"github.com/botwayorg/gh/pkg/cmd/factory"
)

func NewCmdRepo(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gh-repo <command>",
		Short: "Create, clone, fork, and view repositories.",
		Long:  `Work with GitHub repositories`,
		Example: heredoc.Doc(`
			botway gh-repo create
			botway gh-repo clone moby/moby
		`),
		Annotations: map[string]string{
			"help:arguments": heredoc.Doc(`
				A repository can be supplied as an argument in any of the following formats:
				- "OWNER/REPO"
				- by URL, e.g. "https://github.com/OWNER/REPO"
			`),
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			// username := git_config.GitConfig()

			// if username == ":username" {
			// 	utils.AuthMessage()
			// } else {
			// 	cmd.Help()
			// }

			cmd.Help()

			return nil
		},
	}

	repoResolvingCmdFactory := *f
	repoResolvingCmdFactory.BaseRepo = factory.SmartBaseRepoFunc(f)

	cmd.AddCommand(repoViewCmd.NewCmdView(f, nil))
	cmd.AddCommand(repoForkCmd.NewCmdFork(f, nil))
	cmd.AddCommand(repoCloneCmd.NewCmdClone(f, nil))
	cmd.AddCommand(repoCreateCmd.NewCmdCreate(f, nil))
	cmd.AddCommand(repoListCmd.NewCmdList(f, nil))
	cmd.AddCommand(repoSyncCmd.NewCmdSync(f, nil))
	cmd.AddCommand(creditsCmd.NewCmdRepoCredits(f, nil))
	cmd.AddCommand(gardenCmd.NewCmdGarden(f, nil))
	cmd.AddCommand(repoBrowseCmd.NewCmdBrowse(&repoResolvingCmdFactory, nil))

	return cmd
}
