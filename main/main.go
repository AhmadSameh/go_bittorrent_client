package main

import (
	"bittorrent_client/internal/torrent"
	"log"
	"os"
)

func main() {
	inPath := os.Args[1]
	outPath := os.Args[2]

	tf, err := torrent.OpenTorrent(inPath)
	if err != nil {
		log.Fatal(err)
	}

	err = tf.DownloadTorrent(outPath)
	if err != nil {
		log.Fatal(err)
	}
}
