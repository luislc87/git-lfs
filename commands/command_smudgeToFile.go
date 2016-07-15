package commands

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"os"
	"path/filepath"

	"github.com/github/git-lfs/config"
	"github.com/github/git-lfs/errutil"
	"github.com/github/git-lfs/lfs"
	"github.com/spf13/cobra"
)

var (
	smudgeToFileInfo = false
	smudgeToFileSkip = false
	smudgeToFileCmd  = &cobra.Command{
		Use: "smudgeToFile",
		Run: smudgeToFileCommand,
	}
)

func smudgeToFile(reader *bufio.Reader) {

	f, _ := os.OpenFile("/Users/lars/Code/git/t/output.txt", os.O_APPEND|os.O_WRONLY, 0600)
	f.WriteString("smudge\n")
	f.Close()

	// Read fileName length
	buf := make([]byte, 4)
	_, err := reader.Read(buf)
	if err != nil {
		Panic(err, "Error reading asset filename length for smudging.")
	}
	fileNameLen := binary.LittleEndian.Uint32(buf)

	// Read fileName
	buf = make([]byte, fileNameLen)
	_, err = reader.Read(buf)
	if err != nil {
		Panic(err, "Error reading asset filename for smudging.")
	}
	fileName := string(buf)

	// Read asset length
	buf = make([]byte, 4)
	_, err = reader.Read(buf)
	if err != nil {
		Panic(err, "Error reading asset pointer length for smudging.")
	}
	assetPtrLen := binary.LittleEndian.Uint32(buf)

	// Read asset
	buf = make([]byte, assetPtrLen)
	_, err = reader.Read(buf)
	if err != nil {
		Panic(err, "Error reading asset pointer for smudging.")
	}
	ptr, err := lfs.DecodeKV(bytes.TrimSpace(buf))
	if err != nil {
		// Panic(err, "TODO")
	}

	lfs.LinkOrCopyFromReference(ptr.Oid, ptr.Size)

	if smudgeToFileInfo {
		localPath, err := lfs.LocalMediaPath(ptr.Oid)
		if err != nil {
			Exit(err.Error())
		}

		stat, err := os.Stat(localPath)
		if err != nil {
			Print("%d --", ptr.Size)
		} else {
			Print("%d %s", stat.Size(), localPath)
		}
		return
	}

	cb, file, err := lfs.CopyCallbackFile("smudge", fileName, 1, 1)
	if err != nil {
		Error(err.Error())
	}

	cfg := config.Config
	download := lfs.FilenamePassesIncludeExcludeFilter(fileName, cfg.FetchIncludePaths(), cfg.FetchExcludePaths())

	if smudgeToFileSkip || cfg.GetenvBool("GIT_LFS_SKIP_SMUDGE", false) {
		download = false
	}

	err = lfs.PointerSmudgeToFile(fileName, ptr, download, cb)
	if file != nil {
		file.Close()
	}

	if err != nil {

		// ptr.Encode(os.Stdout)
		// Download declined error is ok to skip if we weren't requesting download
		if !(errutil.IsDownloadDeclinedError(err) && !download) {
			LoggedError(err, "Error downloading object: %s (%s)", fileName, ptr.Oid)
			if !cfg.SkipDownloadErrors() {
				os.Exit(2)
			}
		}
	}
}

func smudgeToFileCommand(cmd *cobra.Command, args []string) {
	requireStdin("This command should be run by the Git 'smudge' filter")
	lfs.InstallHooks(false)
	reader := bufio.NewReader(os.Stdin)
	for {
		buf := make([]byte, 1)
		_, err := reader.Read(buf)
		if err != nil {
			continue
		}
		switch buf[0] {
		case 1:
			smudgeToFile(reader)
		case 9:
			return
		default:
			panic("Unrecognized smudgeToFile command")
		}
	}

}

func smudgeToFileFilename(args []string, err error) string {
	if len(args) > 0 {
		return args[0]
	}

	if errutil.IsSmudgeError(err) {
		return filepath.Base(errutil.ErrorGetContext(err, "FileName").(string))
	}

	return "<unknown file>"
}

func init() {
	// update man page
	smudgeToFileCmd.Flags().BoolVarP(&smudgeToFileInfo, "info", "i", false, "")
	smudgeToFileCmd.Flags().BoolVarP(&smudgeToFileSkip, "skip", "s", false, "")
	RootCmd.AddCommand(smudgeToFileCmd)
}
