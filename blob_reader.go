package zim

import (
	"encoding/binary"
	"errors"
	"io"
	"io/ioutil"
	"os"
)

const defaultOffsetSize = 4
const extendedOffsetSize = 8

func clusterOffsetSize(clusterInformation uint8) uint8 {
	return 4 << ((clusterInformation & 16) >> 4)
}

func clusterCompression(clusterInformation uint8) uint8 {
	return clusterInformation & 15
}

func fileAtPosition(f *os.File, offsetSize int64, blobPosition uint32) (
	reader io.Reader, blobSize int64, err error) {
	// TODO: include in readerAtPosition without performance loss
	var thisBlobIndex = int64(blobPosition) * offsetSize
	newOffset, _ := f.Seek(thisBlobIndex, 1)
	filePosBegin := newOffset - thisBlobIndex

	var thisBlobPointer int64
	var nextBlobPointer int64

	switch offsetSize {
	case extendedOffsetSize:
		thisBlobPointer = int64(readUint64(f))
		nextBlobPointer = int64(readUint64(f))
	default:
		thisBlobPointer = int64(readUint32(f))
		nextBlobPointer = int64(readUint32(f))
	}

	if nextBlobPointer < thisBlobPointer {
		err = errors.New("zim: invalid blob index")
		return
	}

	seek(f, filePosBegin+thisBlobPointer)
	blobSize = nextBlobPointer - thisBlobPointer
	return io.LimitReader(f, blobSize), blobSize, nil
}

func readerAtPosition(srcReader io.Reader, offsetSize int64, blobPosition uint32) (
	reader io.Reader, blobSize int64, err error) {

	var thisBlobIndex = int64(blobPosition) * offsetSize

	// seek to position where we get the relevant start and end positions of the blob
	if _, err = io.CopyN(ioutil.Discard, srcReader, thisBlobIndex); err != nil {
		return
	}

	// read the start and end positions
	var thisBlobPointer int64
	var nextBlobPointer int64
	{
		var uintBuf [extendedOffsetSize]byte
		switch offsetSize {
		case extendedOffsetSize:
			srcReader.Read(uintBuf[:extendedOffsetSize])
			thisBlobPointer = int64(binary.LittleEndian.Uint64(uintBuf[:extendedOffsetSize]))
			srcReader.Read(uintBuf[:extendedOffsetSize])
			nextBlobPointer = int64(binary.LittleEndian.Uint64(uintBuf[:extendedOffsetSize]))
		default:
			srcReader.Read(uintBuf[:defaultOffsetSize])
			thisBlobPointer = int64(binary.LittleEndian.Uint32(uintBuf[:defaultOffsetSize]))
			srcReader.Read(uintBuf[:defaultOffsetSize])
			nextBlobPointer = int64(binary.LittleEndian.Uint32(uintBuf[:defaultOffsetSize]))
		}
	}

	if nextBlobPointer < thisBlobPointer {
		err = errors.New("zim: invalid blob index")
		return
	}

	var alreadyRead = thisBlobIndex + 2*offsetSize

	// seek to the position of blob data start
	if _, err = io.CopyN(ioutil.Discard, srcReader, thisBlobPointer-alreadyRead); err != nil {
		return
	}

	blobSize = nextBlobPointer - thisBlobPointer
	reader = io.LimitReader(srcReader, blobSize)
	return
}

// BlobReaderAt returns a LimitedReader for the blob data at the given positions.
func (z *File) BlobReaderAt(clusterPosition, blobPosition uint32) (
	reader io.Reader, blobSize int64, err error) {
	var clusterPointer = z.clusterPointerAtPos(clusterPosition)
	seek(z.f, int64(clusterPointer))
	var clusterInformation = readUint8(z.f)
	var offsetSize = int64(clusterOffsetSize(clusterInformation))
	var compression = clusterCompression(clusterInformation)
	switch compression {
	case 0, 1:
		// uncompressed
		reader, blobSize, err = fileAtPosition(z.f, offsetSize, blobPosition)
		// takes 0ms - 2ms for creating a prepared LimitedReader from os.File
	case 4:
		// xz compressed
		if err = z.xzReader.Reset(z.f); err == nil {
			z.xzReader.Multistream(false)
			reader, blobSize, err = readerAtPosition(z.xzReader, offsetSize, blobPosition)
			// takes 1ms - 130ms for creating a prepared LimitedReader from a generic io.Reader
		}
	default:
		// 2: zlib compressed (not used anymore)
		// 3: bzip2 compressed (not used anymore)
		err = errors.New("zim: unsupported cluster compression")
	}
	return
}

// BlobReader returns a LimitedReader for the blob data of the given
// Directory Entry.
func (z *File) BlobReader(e *DirectoryEntry) (
	reader io.Reader, blobSize int64, err error) {
	return z.BlobReaderAt(e.clusterNumber, e.blobNumber)
}
