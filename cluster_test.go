package zim

import (
	"bytes"
	"testing"
)

func TestClusterAt(t *testing.T) {
	const expectedNumberHTMLFiles = 4
	var numberHTMLFiles = 0
	for position := uint32(0); position < z.ClusterCount(); position++ {
		var cluster, clusterErr = z.ClusterAt(position)

		if clusterErr != nil {
			t.Errorf("error at cluster position %d/%d: %s\n", position, z.ClusterCount()-1, clusterErr)
		}

		for blobPosition := uint32(0); ; blobPosition++ {
			blob, blobErr := cluster.BlobAt(blobPosition)
			if blobErr != nil {
				if blobPosition == 0 {
					t.Errorf("failed to read first blob at cluster position %d\n", position)
				}
				break
			}

			if bytes.Index(blob, []byte("<html")) >= 0 && bytes.Index(blob, []byte("</html>")) > 0 {
				numberHTMLFiles++
			}
		}

	}

	if numberHTMLFiles != expectedNumberHTMLFiles {
		t.Errorf("Number of HTML files in ZIM file was %d; want %d\n", numberHTMLFiles, expectedNumberHTMLFiles)
	}
}
