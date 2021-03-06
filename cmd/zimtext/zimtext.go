package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/tim-st/go-zim"
)

func main() {

	var filenameZim string
	var filenameText string
	var limit int
	var singleSentences bool

	flag.StringVar(&filenameZim, "zim", "", "Path to the ZIM file to read from.")
	flag.StringVar(&filenameText, "txt", "", "Path to the text file, that is created or truncated if exists.")
	flag.IntVar(&limit, "limit", -1, "Stop after N lines were written (where N >= limit).")
	flag.BoolVar(&singleSentences, "sentences", false, "Only write paragraphs which are likely a single sentence.")
	flag.Parse()

	if flag.NFlag() < 2 || len(filenameZim) == 0 || len(filenameText) == 0 {
		flag.PrintDefaults()
		return
	}

	z, zimOpenErr := zim.Open(filenameZim)

	if zimOpenErr != nil {
		log.Fatal(zimOpenErr)
	}

	var txtFile, txtFileErr = os.Create(filenameText)

	if txtFileErr != nil {
		log.Fatal(txtFileErr)
	}

	var bufWriter = bufio.NewWriterSize(txtFile, 1<<22) // 4mb buffer

	var paragraphsWritten = 0

	var sliceReader = bytes.NewReader(nil)

	var printProgress func(clusterPosition uint32)

	if limit > 0 {
		printProgress = func(clusterPosition uint32) {
			if clusterPosition%4 == 0 {
				fmt.Printf("\r%.1f%%", (float32(paragraphsWritten)/float32(limit))*100)
			}
		}
	} else {
		limit = int(^uint(0) >> 1)
		printProgress = func(clusterPosition uint32) {
			if clusterPosition%16 == 0 {
				fmt.Printf("\r%.1f%%", (float32(clusterPosition)/float32(z.ClusterCount()))*100)
			}
		}
	}

	defer func() {
		fmt.Print("\r100.0%")
		bufWriter.Flush()
		txtFile.Close()
	}()

	for clusterPosition := uint32(0); clusterPosition < z.ClusterCount(); clusterPosition++ {

		printProgress(clusterPosition)

		var cluster, clusterErr = z.ClusterAt(clusterPosition)

		if clusterErr != nil {
			continue
		}

		if !cluster.WasCompressed() {
			continue
		}

		for blobPosition := uint32(0); ; blobPosition++ {

			var requiredParagraphs = limit - paragraphsWritten

			if requiredParagraphs <= 0 || paragraphsWritten >= limit {
				return
			}

			var blob, blobErr = cluster.BlobAt(blobPosition)

			if blobErr != nil {
				break
			}

			if bytes.Index(blob, []byte("<html")) >= 0 && bytes.Index(blob, []byte("</html>")) > 5 {
				sliceReader.Reset(blob)
				if singleSentences {
					paragraphsWritten += WriteCleanSentences(sliceReader, bufWriter, requiredParagraphs)
				} else {
					paragraphsWritten += WriteCleanText(sliceReader, bufWriter, requiredParagraphs)
				}
			}
		}

	}
}
