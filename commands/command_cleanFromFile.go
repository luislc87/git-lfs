package commands

import (
	"bufio"
	"encoding/binary"
	"os"

	"github.com/github/git-lfs/errutil"
	"github.com/github/git-lfs/lfs"
	"github.com/spf13/cobra"
)

var (
	cleanFromFileCmd = &cobra.Command{
		Use: "cleanFromFile",
		Run: cleanFromFileCommand,
	}
)

func cleanFile(reader *bufio.Reader) {

	f, _ := os.OpenFile("/Users/lars/Code/git/t/output.txt", os.O_APPEND|os.O_WRONLY, 0600)
	f.WriteString("clean\n")
	f.Close()

	// Read fileName length
	buf := make([]byte, 4)
	_, err := reader.Read(buf)
	if err != nil {
		Panic(err, "Error reading asset filename length for cleaning.")
	}
	fileNameLen := binary.LittleEndian.Uint32(buf)

	// Read fileName
	buf = make([]byte, fileNameLen)
	_, err = reader.Read(buf)
	if err != nil {
		Panic(err, "Error reading asset filename for cleaning.")
	}
	fileName := string(buf)

	// Generate LFS pointer
	// TODO: ProgressCallback?!
	fileToClean, _ := os.OpenFile(fileName, os.O_RDONLY, 0600)
	cleaned, err := lfs.PointerClean(fileToClean, fileName, 0, nil)
	fileToClean.Close()

	if cleaned != nil {
		defer cleaned.Teardown()
	}

	if errutil.IsCleanPointerError(err) {
		os.Stdout.Write(errutil.ErrorGetContext(err, "bytes").([]byte))
		return
	}

	if err != nil {
		Panic(err, "Error cleaning asset.")
	}

	// Write LFS pointer
	encodedPointer := cleaned.Pointer.Encoded()
	binary.Write(os.Stdout, binary.LittleEndian, uint32(len(encodedPointer)))
	os.Stdout.Write([]byte(encodedPointer))
}

func cleanFromFileCommand(cmd *cobra.Command, args []string) {
	requireStdin("This command should be run by the Git 'clean' filter")
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
			cleanFile(reader)
		case 9:
			return
		default:
			panic("Unrecognized cleanFromFile command")
		}
	}
}

func init() {
	RootCmd.AddCommand(cleanFromFileCmd)
}
