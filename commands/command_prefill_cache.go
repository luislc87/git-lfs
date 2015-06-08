package commands

import (
	"io/ioutil"
	"os"

	"github.com/garyburd/redigo/redis"
	"github.com/github/git-lfs/vendor/_nuts/github.com/spf13/cobra"
)

var (
	prefillCacheCmd = &cobra.Command{
		Use:   "prefillCache",
		Short: "TODO",
		Run:   prefillCacheCommand,
	}
)

func prefillCacheCommand(cmd *cobra.Command, args []string) {

	if len(args) == 0 {
		Print("Specify a remote and a remote branch name (`git lfs push origin master`)")
		os.Exit(1)
	}

	var oid = args[0]
	var filename = args[1]

	redisConnection, redisConnectionErr := redis.Dial("tcp", "10.146.248.76:6379")
	if redisConnectionErr != nil {
		panic(redisConnectionErr)
	}
	defer redisConnection.Close()

	file_content, _ := ioutil.ReadFile(filename)
	redisConnection.Do("SET", oid, file_content)
}

func init() {
	RootCmd.AddCommand(prefillCacheCmd)
}
