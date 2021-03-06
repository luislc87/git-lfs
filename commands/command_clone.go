package commands

import (
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/github/git-lfs/subprocess"

	"github.com/github/git-lfs/config"
	"github.com/github/git-lfs/git"
	"github.com/github/git-lfs/localstorage"
	"github.com/github/git-lfs/tools"
	"github.com/spf13/cobra"
)

var (
	cloneCmd = &cobra.Command{
		Use: "clone",
		Run: cloneCommand,
	}

	cloneFlags      git.CloneFlags
	cloneIncludeArg string
	cloneExcludeArg string
)

func cloneCommand(cmd *cobra.Command, args []string) {

	// We pass all args to git clone
	err := git.CloneWithoutFilters(cloneFlags, args)
	if err != nil {
		Exit("Error(s) during clone:\n%v", err)
	}

	// now execute pull (need to be inside dir)
	cwd, err := os.Getwd()
	if err != nil {
		Exit("Unable to derive current working dir: %v", err)
	}

	// Either the last argument was a relative or local dir, or we have to
	// derive it from the clone URL
	clonedir, err := filepath.Abs(args[len(args)-1])
	if err != nil || !tools.DirExists(clonedir) {
		// Derive from clone URL instead
		base := path.Base(args[len(args)-1])
		if strings.HasSuffix(base, ".git") {
			base = base[:len(base)-4]
		}
		clonedir, _ = filepath.Abs(base)
		if !tools.DirExists(clonedir) {
			Exit("Unable to find clone dir at %q", clonedir)
		}
	}

	err = os.Chdir(clonedir)
	if err != nil {
		Exit("Unable to change directory to clone dir %q: %v", clonedir, err)
	}

	// Make sure we pop back to dir we started in at the end
	defer os.Chdir(cwd)

	// Also need to derive dirs now
	localstorage.ResolveDirs()
	requireInRepo()

	// Now just call pull with default args
	// Support --origin option to clone
	if len(cloneFlags.Origin) > 0 {
		config.Config.CurrentRemote = cloneFlags.Origin
	} else {
		config.Config.CurrentRemote = "origin"
	}

	include, exclude := determineIncludeExcludePaths(config.Config, cloneIncludeArg, cloneExcludeArg)
	if cloneFlags.NoCheckout || cloneFlags.Bare {
		// If --no-checkout or --bare then we shouldn't check out, just fetch instead
		fetchRef("HEAD", include, exclude)
	} else {
		pull(include, exclude)

		err := postCloneSubmodules(args)
		if err != nil {
			Exit("Error performing 'git lfs pull' for submodules: %v", err)
		}
	}

}

func postCloneSubmodules(args []string) error {
	// In git 2.9+ the filter option will have been passed through to submodules
	// So we need to lfs pull inside each
	if !git.Config.IsGitVersionAtLeast("2.9.0") {
		// In earlier versions submodules would have used smudge filter
		return nil
	}
	// Also we only do this if --recursive or --recurse-submodules was provided
	if !cloneFlags.Recursive && !cloneFlags.RecurseSubmodules {
		return nil
	}

	// Use `git submodule foreach --recursive` to cascade into nested submodules
	// Also good to call a new instance of git-lfs rather than do things
	// inside this instance, since that way we get a clean env in that subrepo
	cmd := subprocess.ExecCommand("git", "submodule", "foreach", "--recursive",
		"git lfs pull")
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	return cmd.Run()
}

func init() {
	// Mirror all git clone flags
	cloneCmd.Flags().StringVarP(&cloneFlags.TemplateDirectory, "template", "", "", "See 'git clone --help'")
	cloneCmd.Flags().BoolVarP(&cloneFlags.Local, "local", "l", false, "See 'git clone --help'")
	cloneCmd.Flags().BoolVarP(&cloneFlags.Shared, "shared", "s", false, "See 'git clone --help'")
	cloneCmd.Flags().BoolVarP(&cloneFlags.NoHardlinks, "no-hardlinks", "", false, "See 'git clone --help'")
	cloneCmd.Flags().BoolVarP(&cloneFlags.Quiet, "quiet", "q", false, "See 'git clone --help'")
	cloneCmd.Flags().BoolVarP(&cloneFlags.NoCheckout, "no-checkout", "n", false, "See 'git clone --help'")
	cloneCmd.Flags().BoolVarP(&cloneFlags.Progress, "progress", "", false, "See 'git clone --help'")
	cloneCmd.Flags().BoolVarP(&cloneFlags.Bare, "bare", "", false, "See 'git clone --help'")
	cloneCmd.Flags().BoolVarP(&cloneFlags.Mirror, "mirror", "", false, "See 'git clone --help'")
	cloneCmd.Flags().StringVarP(&cloneFlags.Origin, "origin", "o", "", "See 'git clone --help'")
	cloneCmd.Flags().StringVarP(&cloneFlags.Branch, "branch", "b", "", "See 'git clone --help'")
	cloneCmd.Flags().StringVarP(&cloneFlags.Upload, "upload-pack", "u", "", "See 'git clone --help'")
	cloneCmd.Flags().StringVarP(&cloneFlags.Reference, "reference", "", "", "See 'git clone --help'")
	cloneCmd.Flags().BoolVarP(&cloneFlags.Dissociate, "dissociate", "", false, "See 'git clone --help'")
	cloneCmd.Flags().StringVarP(&cloneFlags.SeparateGit, "separate-git-dir", "", "", "See 'git clone --help'")
	cloneCmd.Flags().StringVarP(&cloneFlags.Depth, "depth", "", "", "See 'git clone --help'")
	cloneCmd.Flags().BoolVarP(&cloneFlags.Recursive, "recursive", "", false, "See 'git clone --help'")
	cloneCmd.Flags().BoolVarP(&cloneFlags.RecurseSubmodules, "recurse-submodules", "", false, "See 'git clone --help'")
	cloneCmd.Flags().StringVarP(&cloneFlags.Config, "config", "c", "", "See 'git clone --help'")
	cloneCmd.Flags().BoolVarP(&cloneFlags.SingleBranch, "single-branch", "", false, "See 'git clone --help'")
	cloneCmd.Flags().BoolVarP(&cloneFlags.NoSingleBranch, "no-single-branch", "", false, "See 'git clone --help'")
	cloneCmd.Flags().BoolVarP(&cloneFlags.Verbose, "verbose", "", false, "See 'git clone --help'")
	cloneCmd.Flags().BoolVarP(&cloneFlags.Ipv4, "ipv4", "", false, "See 'git clone --help'")
	cloneCmd.Flags().BoolVarP(&cloneFlags.Ipv6, "ipv6", "", false, "See 'git clone --help'")

	cloneCmd.Flags().StringVarP(&cloneIncludeArg, "include", "I", "", "Include a list of paths")
	cloneCmd.Flags().StringVarP(&cloneExcludeArg, "exclude", "X", "", "Exclude a list of paths")

	RootCmd.AddCommand(cloneCmd)
}
