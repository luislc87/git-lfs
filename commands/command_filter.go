package commands

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/github/git-lfs/config"
	"github.com/github/git-lfs/errutil"
	"github.com/github/git-lfs/lfs"
	"github.com/github/git-lfs/progress"
	"github.com/spf13/cobra"
)

const (
	FilterDriverVersion = 1
)

type InputFileHdr struct {
	FileName string
	FileLen  uint64
}

func (h *InputFileHdr) Read(r *bufio.Reader) error {
	fileName, err := r.ReadString('\n')
	if err == io.EOF {
		return err
	} else if err != nil {
		Panic(err, "Error reading filter command.")
	}
	h.FileName = strings.TrimSuffix(fileName, "\n")

	fileLenStr, err := r.ReadString('\n')
	if err == io.EOF {
		return err
	} else if err != nil {
		Panic(err, "Error reading input data length.")
	}
	fileLenStr = strings.TrimSuffix(fileLenStr, "\n")
	fileLen, err := strconv.ParseUint(fileLenStr, 10, 64)
	if err != nil {
		Panic(err, fmt.Sprintf("Error parsing input data length '%s'.", fileLenStr))
	}
	h.FileLen = fileLen

	// Panic(err, fmt.Sprintf("Errorddddddddddddgth '%d'.", fileLen))

	return nil
}

var (
	filterSmudgeSkip = false
	filterCmd        = &cobra.Command{
		Use: "filter",
		Run: filterCommand,
	}
)

func clean(reader io.Reader, fileName string) ([]byte, error) {
	var cb progress.CopyCallback
	var file *os.File
	var fileSize int64
	if len(fileName) > 0 {
		stat, err := os.Stat(fileName)
		if err == nil && stat != nil {
			fileSize = stat.Size()

			localCb, localFile, err := lfs.CopyCallbackFile("clean", fileName, 1, 1)
			if err != nil {
				Error(err.Error())
			} else {
				cb = localCb
				file = localFile
			}
		}
	}

	cleaned, err := lfs.PointerClean(reader, fileName, fileSize, cb)
	if file != nil {
		file.Close()
	}

	if cleaned != nil {
		defer cleaned.Teardown()
	}

	if errutil.IsCleanPointerError(err) {
		// TODO: What is happening here? (triggered by ./test/test-clone.sh)
		return errutil.ErrorGetContext(err, "bytes").([]byte), nil
	}

	if err != nil {
		Panic(err, "Error cleaning inputData.")
	}

	tmpfile := cleaned.Filename
	mediafile, err := lfs.LocalMediaPath(cleaned.Oid)
	if err != nil {
		Panic(err, "Unable to get local media path.")
	}

	if stat, _ := os.Stat(mediafile); stat != nil {
		if stat.Size() != cleaned.Size && len(cleaned.Pointer.Extensions) == 0 {
			Exit("Files don't match:\n%s\n%s", mediafile, tmpfile)
		}
		Debug("%s exists", mediafile)
	} else {
		if err := os.Rename(tmpfile, mediafile); err != nil {
			Panic(err, "Unable to move %s to %s\n", tmpfile, mediafile)
		}

		Debug("Writing %s", mediafile)
	}

	return []byte(cleaned.Pointer.Encoded()), nil
}

func smudge(reader io.Reader, filename string) ([]byte, error) {
	ptr, err := lfs.DecodePointer(reader)
	if err != nil {
		// TODO: No test seems to trigger this code path
		// mr := io.MultiReader(b, os.Stdin)
		// _, err := io.Copy(os.Stdout, mr)
		// if err != nil {
		// Panic(err, "Error writing data to stdout:")
		// }
		return []byte("TODO"), nil
	}

	lfs.LinkOrCopyFromReference(ptr.Oid, ptr.Size)

	cb, file, err := lfs.CopyCallbackFile("smudge", filename, 1, 1)
	if err != nil {
		Error(err.Error())
	}

	cfg := config.Config
	download := lfs.FilenamePassesIncludeExcludeFilter(filename, cfg.FetchIncludePaths(), cfg.FetchExcludePaths())

	if filterSmudgeSkip || cfg.GetenvBool("GIT_LFS_SKIP_SMUDGE", false) {
		download = false
	}

	buf := new(bytes.Buffer)
	err = ptr.Smudge(buf, filename, download, cb)
	if file != nil {
		file.Close()
	}

	if err != nil {
		// Download declined error is ok to skip if we weren't requesting download
		if !(errutil.IsDownloadDeclinedError(err) && !download) {
			LoggedError(err, "Error downloading object: %s (%s)", filename, ptr.Oid)
			if !cfg.SkipDownloadErrors() {
				// TODO: What to do best here?
				os.Exit(2)
			}
		}

		return []byte(ptr.Encoded()), nil
	}

	return buf.Bytes(), nil
}

func filterCommand(cmd *cobra.Command, args []string) {
	requireStdin("This command should be run by the Git filter")
	lfs.InstallHooks(false)

	reader := bufio.NewReader(os.Stdin)
	writer := bufio.NewWriter(os.Stdout)

	writer.WriteString(fmt.Sprintf("version %d", FilterDriverVersion))
	writer.Flush()

	for {
		var command string
		command, err := reader.ReadString('\n')
		if err == io.EOF {
			return
		} else if err != nil {
			Panic(err, "Error reading filter command.")
		}
		command = strings.TrimSuffix(command, "\n")

		var inputHeader InputFileHdr
		if err := inputHeader.Read(reader); err != nil {
			Panic(err, "Error reading input header.")
		}

		// Read inputData
		var outputData []byte
		if inputHeader.FileLen > 0 {
			inputData := io.LimitReader(reader, int64(inputHeader.FileLen))
			switch command {
			case "clean":
				outputData, _ = clean(inputData, inputHeader.FileName)
			case "smudge":
				outputData, _ = smudge(inputData, inputHeader.FileName)
			default:
				panic(fmt.Sprintf("Unrecognized filter command '%s'.", command))
			}
		}

		writer.WriteString(fmt.Sprintf("%d\n", len(outputData)))
		writer.Flush()
		if len(outputData) > 0 {
			writer.Write(outputData)
		}
		writer.Flush()
	}
}

func init() {
	filterCmd.Flags().BoolVarP(&filterSmudgeSkip, "skip-smudge", "s", false, "")
	RootCmd.AddCommand(filterCmd)
}
