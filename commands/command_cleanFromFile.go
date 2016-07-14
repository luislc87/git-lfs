package commands

import (
	"os"

	// "github.com/github/git-lfs/errutil"
	// "github.com/github/git-lfs/lfs"
	// "github.com/github/git-lfs/progress"
	"github.com/spf13/cobra"

	"bufio"
	// "encoding/binary"
	"fmt"
	"time"
)

var (
	cleanFromFileCmd = &cobra.Command{
		Use: "cleanFromFile",
		Run: cleanFromFileCommand,
	}
)

func cleanFromFileCommand(cmd *cobra.Command, args []string) {
	// requireStdin("This command should be run by the Git 'clean' filter")
	// lfs.InstallHooks(false)

	// os.Exit(1)
	// nBytes, nChunks := int64(0), int64(0)
	r := bufio.NewReader(os.Stdin)
	buf := make([]byte, 0, 1)
	// var pathSize uint32

	for {
		n, err := r.Read(buf)

		if n == 0 {
			if err != nil {
				break
			}
			time.Sleep(100 * time.Millisecond)
			f, _ := os.OpenFile("/Users/lars/Code/git/t/output.txt", os.O_APPEND|os.O_WRONLY, 0600)
			f.WriteString("sleep\n")
			f.Close()
		} else {

			f, _ := os.OpenFile("/Users/lars/Code/git/t/output.txt", os.O_APPEND|os.O_WRONLY, 0600)
			f.WriteString(fmt.Sprintf("hallo %d\n", n))
			f.Close()
			fmt.Fprint(os.Stdout, "lars")
		}
	}

	// for {
	// }

	// fmt.Println("lars")
	// nChunks++
	// nBytes += int64(len(buf))
	// // process buf
	// if err != nil && err != io.EOF {
	//     log.Fatal(err)
	// }
	// }

	// var i int
	// for {
	// 	out, err := fmt.Fscanf(os.Stdin, "%s\n")

	// }

	// scanner := bufio.NewScanner(os.Stdin)
	// for scanner.Scan() {
	// 	line := scanner.Text()
	// 	if line == "SHUTDOWN" {
	// 		break
	// 	}
	// 	bs := make([]byte, 4)
	// 	// binary.BigEndian.PutUint32(bs, 4)
	// 	fmt.Println(bs)
	// 	fmt.Println("lars")
	// 	// fmt.Println(line) // or do something else with line
	// }

	// var fileName string
	// var cb progress.CopyCallback
	// var file *os.File
	// var fileSize int64
	// if len(args) > 0 {
	// 	fileName = args[0]

	// 	stat, err := os.Stat(fileName)
	// 	if err == nil && stat != nil {
	// 		fileSize = stat.Size()

	// 		localCb, localFile, err := lfs.CopyCallbackFile("clean", fileName, 1, 1)
	// 		if err != nil {
	// 			Error(err.Error())
	// 		} else {
	// 			cb = localCb
	// 			file = localFile
	// 		}
	// 	}
	// }

	// cleaned, err := lfs.PointerClean(os.Stdin, fileName, fileSize, cb)
	// if file != nil {
	// 	file.Close()
	// }

	// if cleaned != nil {
	// 	defer cleaned.Teardown()
	// }

	// if errutil.IsCleanPointerError(err) {
	// 	os.Stdout.Write(errutil.ErrorGetContext(err, "bytes").([]byte))
	// 	return
	// }

	// if err != nil {
	// 	Panic(err, "Error cleaning asset.")
	// }

	// tmpfile := cleaned.Filename
	// mediafile, err := lfs.LocalMediaPath(cleaned.Oid)
	// if err != nil {
	// 	Panic(err, "Unable to get local media path.")
	// }

	// if stat, _ := os.Stat(mediafile); stat != nil {
	// 	if stat.Size() != cleaned.Size && len(cleaned.Pointer.Extensions) == 0 {
	// 		Exit("Files don't match:\n%s\n%s", mediafile, tmpfile)
	// 	}
	// 	Debug("%s exists", mediafile)
	// } else {
	// 	if err := os.Rename(tmpfile, mediafile); err != nil {
	// 		Panic(err, "Unable to move %s to %s\n", tmpfile, mediafile)
	// 	}

	// 	Debug("Writing %s", mediafile)
	// }

	// lfs.EncodePointer(os.Stdout, cleaned.Pointer)
}

func init() {
	RootCmd.AddCommand(cleanFromFileCmd)
}
