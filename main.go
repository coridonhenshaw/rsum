package main

import (
	"fmt"
	"log"

	"github.com/alecthomas/kong"
)

const RemoveWorkingDatabase = false
const WorkerThreads = 4
const LargeFileThreshold = 1024 * 1024 * 16

var cli struct {
	Recurse  bool     `help:"Recurse into subdirectories." short:"r" name:"recurse"`
	HashType string   `help:"Select hash algorithm. Valid options are blake2b-512 and sha512." short:"t" name:"hashtype" enum:"blake2b-512,sha512" default:"blake2b-512"`
	Base64   bool     `help:"Write sums in RFC 4648 base 64 rather than hexadecimal." short:"" name:"base64"`
	Verbose  bool     `help:"Show progress." short:"v" name:"verbose"`
	UseCWD   bool     `help:"Store resume database in the current working directory rather than in the user cache directory." short:"c" name:"usecwd"`
	Output   string   `required:"" help:"Output filename" short:"o" name:"output"`
	Paths    []string `arg:"" help:"Input files" name:"paths"`
}

var DB SqliteStruct

var WorkingDatabaseInCWD bool
var UseBase64 bool
var Verbose bool

func main() {

	var err error

	fmt.Print("rsum release 0. THIS IS EXPERIMENTAL SOFTWARE. DO NOT USE IT FOR ANYTHING IMPORTANT.\n\n")

	kong.Parse(&cli,
		kong.Name("rsum"),
		kong.Description("Recursive Sum Tool"),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
			Summary: true,
		}))

	DB, err = DBConnect()
	if err != nil {
		log.Panic(err)
	}
	defer DB.Close()

	UseBase64 = cli.Base64
	Verbose = cli.Verbose
	WorkingDatabaseInCWD = cli.UseCWD

	switch cli.HashType {
	case "blake2b-512":
		HashType = HashTypeBlake2B
	case "sha512":
		HashType = HashTypeSHA512
	case "sha3-512":
		HashType = HashTypeSHA3_512
	}

	// w, err := os.OpenFile("rsum.prof", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	// defer w.Close()

	// runtime.StartCPUProfile(w)

	CreateSumList(cli.Paths, cli.Recurse, cli.Output)

	// runtime.StopCPUProfile()

}
