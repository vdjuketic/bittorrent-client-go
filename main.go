package main

import (
	"flag"
	"fmt"
	"os"
)

const DOWNLOAD_COMMAND = "download"

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Expected download command")
		os.Exit(1)
	}

	switch command := os.Args[1]; command {
	case DOWNLOAD_COMMAND:
		handleDownloadCommand()
	default:
		fmt.Printf("Command not supported.")
		os.Exit(1)
	}
}

func handleDownloadCommand() {
	downloadCmd := flag.NewFlagSet(DOWNLOAD_COMMAND, flag.ExitOnError)
	output := downloadCmd.String("output", "", "output location")
	torrentFile := downloadCmd.String("torrent", "", "torrent file location")
	downloadCmd.Parse(os.Args[2:])

	if *output == "" {
		fmt.Println("output not specified")
		os.Exit(1)
	}

	if *torrentFile == "" {
		fmt.Println("torrent file not specified")
		os.Exit(1)
	}

	handleDownload(*output, *torrentFile)
}

func handleDownload(output string, torrentFile string) {
	fmt.Printf("downloading %s to %s\n", torrentFile, output)

	file, err := os.Open(torrentFile)
	if err != nil {
		fmt.Println("Invalid torrent file location.")
		panic(err)
	}

	torrentMeta := fromBencode(file)
	downloadTorrent(torrentMeta)
}
