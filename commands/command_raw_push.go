package commands

import (
	"os"

	"github.com/github/git-lfs/lfs"
	"github.com/github/git-lfs/vendor/_nuts/github.com/spf13/cobra"
)

var (
	rawPushCmd = &cobra.Command{
		Use:   "raw-push",
		Short: "raw push files to the Git LFS server",
		Run:   rawPushCommand,
	}
	// shares some global vars and functions with commmands_pre_push.go
)

// rawPushCommand pushes a local object to a Git LFS server.  It takes three
// arguments:
//
//   `<remote> <remote ref> <oid>`
//
// Both a remote name ("origin") or a remote URL are accepted.
//
func rawPushCommand(cmd *cobra.Command, args []string) {

	if len(args) != 3 {
		Print("Specify a remote and a remote branch name and a LFS object hash (`git lfs push origin master a123`)")
		os.Exit(1)
	}

	lfs.Config.CurrentRemote = args[0]
	var oid = args[2]

	cb := func(total int64, written int64, current int) error {
		return nil
	}

	oidPath, _ := lfs.LocalMediaPath(oid)

	obj, err := lfs.UploadCheck(oidPath)
	if err != nil {
		Panic(err, "Upload API failed...")
		os.Exit(1)
	}

	if obj != nil {
		// This is a little weird. Why is Oid not set?
		obj.Oid = oid

		err = lfs.UploadObject(obj, cb)
		if err != nil {
			Panic(err, "Upload failed...")
			os.Exit(1)
		}
	}
}

func init() {
	RootCmd.AddCommand(rawPushCmd)
}
