package cmd

import (
	"fmt"
	"os"

	"github.com/axllent/gitlabci-composer-update-mr/app"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "gitlabci-composer-update-mr <commit-user> <commit-email> <source-branch>",
	Short: "A brief description of your application",
	Long: `A Gitlab CI utility to create composer update merge requests.

Documentation:
  https://github.com/axllent/gitlabci-composer-update-mr
`,
	Args: cobra.ExactArgs(3),
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		app.Config.GitUser = args[0]
		app.Config.GitEmail = args[1]
		app.Config.GitBranch = args[2]

		app.BuildConfig()

		if err := app.SwitchBranch(app.Config.GitBranch); err != nil {
			fmt.Printf("\n==========\nError switching branch: %s\n==========\n", err.Error())
			os.Exit(1)
		}

		preupdate, err := app.ParseComposerLock()
		if err != nil {
			fmt.Printf("\n==========\nError parsing composer.lock: %s\n==========\n", err.Error())
			os.Exit(1)
		}

		if _, err := app.ComposerUpdate(); err != nil {
			fmt.Printf("\n==========\nError updating with composer: %s\n==========\n", err.Error())
			os.Exit(1)
		}

		postupdate, err := app.ParseComposerLock()
		if err != nil {
			fmt.Printf("\n==========\nError parsing composer.lock: %s\n==========\n", err.Error())
			os.Exit(1)
		}

		// check if composer lock has been modified
		if preupdate.Checksum == postupdate.Checksum {
			fmt.Println("\n==========\nThere are no updated composer modules\n==========")
			os.Exit(0)
		}

		diff := app.CompareDiffs(preupdate, postupdate)

		if app.MRExists(diff.Checksum) {
			fmt.Printf("\n==========\nAn identical merge request already exists with checksum: %s\n==========\n", diff.Checksum)
			os.Exit(0)
		}

		if err := app.RemoveOldMRs(); err != nil {
			fmt.Printf("\n==========\nError removing old merge requests: %s\n==========\n", err.Error())
			os.Exit(1)
		}

		if err := app.CreateMergeBranch(diff); err != nil {
			fmt.Printf("\n==========\nError creating merge request: %s\n==========\n", err.Error())
			os.Exit(1)
		}

		pkgs := "package"
		if len(diff.Packages) > 0 {
			pkgs = "packages"
		}

		mrTitle := fmt.Sprintf("Composer update: %d %s", len(diff.Packages), pkgs)

		if err := app.CreateMergeRequest(mrTitle, diff.Description); err != nil {
			fmt.Printf("\n==========\n%s\n==========\n", err.Error())
			os.Exit(1)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringSliceVarP(&app.Config.ComposerFlags, "composer-flags", "f", []string{}, "Custom composer flags")
	rootCmd.Flags().StringVarP(&app.Config.RepoDir, "repo", "r", ".", "Repository directory")
	rootCmd.Flags().StringVarP(&app.Config.GitCommitTitle, "commit-title", "t", "Update composer dependencies", "The git commit message title")

	if err := rootCmd.Flags().MarkHidden("repo"); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// hide help command
	rootCmd.SetHelpCommand(&cobra.Command{Hidden: true})
	// hide help flag
	rootCmd.PersistentFlags().BoolP("help", "h", false, "This help")
	rootCmd.PersistentFlags().Lookup("help").Hidden = true
}
