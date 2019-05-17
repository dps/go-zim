package zim

import (
	"errors"
	"io"
	"io/ioutil"
	"os"
)

func blobReader(clusterReader io.Reader, offsetSize int64, blobPosition uint32) (
	reader io.Reader, blobSize int64, err error) {

	var file, clusterReaderIsFile = clusterReader.(*os.File)

	var thisBlobIndex = int64(blobPosition) * offsetSize

	// seek to position where we get the relevant start and end positions of the blob
	if clusterReaderIsFile {
		var newOffset int64
		newOffset, err = file.Seek(thisBlobIndex, 1)
		blobSize = newOffset - thisBlobIndex
	} else {
		_, err = io.CopyN(ioutil.Discard, clusterReader, thisBlobIndex)
	}
	if err != nil {
		err = errors.New("zim: invalid blob position")
		return
	}

	// read the start and end positions
	var thisBlobPointer int64
	var nextBlobPointer int64
	switch offsetSize {
	case extendedOffsetSize:
		thisBlobPointer = int64(readUint64R(clusterReader))
		nextBlobPointer = int64(readUint64R(clusterReader))
	default:
		thisBlobPointer = int64(readUint32R(clusterReader))
		nextBlobPointer = int64(readUint32R(clusterReader))
	}

	if nextBlobPointer < thisBlobPointer {
		err = errors.New("zim: invalid blob index")
		return
	}

	// seek to the position of blob data start
	if clusterReaderIsFile {
		seek(file, blobSize+thisBlobPointer)
	} else {
		var alreadyRead = thisBlobIndex + 2*offsetSize
		_, err = io.CopyN(ioutil.Discard, clusterReader, thisBlobPointer-alreadyRead)
	}

	blobSize = nextBlobPointer - thisBlobPointer
	reader = io.LimitReader(clusterReader, blobSize)
	return
}

// BlobReaderAt returns a LimitedReader for the blob data at the given positions.
func (z *File) BlobReaderAt(clusterPosition, blobPosition uint32) (
	reader io.Reader, blobSize int64, err error) {

	var clusterInformation uint8
	reader, clusterInformation, err = z.clusterReader(clusterPosition)

	if err == nil {
		reader, blobSize, err = blobReader(reader, int64(clusterOffsetSize(clusterInformation)), blobPosition)
	}
	return
}

// BlobReader returns a LimitedReader for the blob data of the given Directory Entry.
func (z *File) BlobReader(e *DirectoryEntry) (
	reader io.Reader, blobSize int64, err error) {
	return z.BlobReaderAt(e.ClusterNumber(), e.BlobNumber())
}
