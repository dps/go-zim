package zim

import (
	"encoding/binary"
	"errors"
	"io"
	"io/ioutil"
)

const (
	defaultOffsetSize  = 4
	extendedOffsetSize = 8
	maxClusterLen      = 1024 * 1024 * 32 // 32MB
)

func clusterOffsetSize(clusterInformation uint8) uint8 {
	return 4 << ((clusterInformation & 16) >> 4)
}

func clusterCompression(clusterInformation uint8) uint8 {
	return clusterInformation & 15
}

func (z *File) clusterReader(clusterPosition uint32) (reader io.Reader, clusterInformation uint8, err error) {
	if clusterPosition >= z.ClusterCount() {
		err = errors.New("zim: invalid cluster position")
		return
	}
	var clusterPointer = z.clusterPointerAtPos(clusterPosition)
	seek(z.f, int64(clusterPointer))
	clusterInformation = readUint8(z.f)
	var compression = clusterCompression(clusterInformation)
	switch compression {
	case 0, 1: // uncompressed
		reader = z.f
	case 4: // xz compressed
		if err = z.xzReader.Reset(z.f); err == nil {
			z.xzReader.Multistream(false)
			reader = z.xzReader
		}
	default:
		// 2: zlib compressed (not used anymore)
		// 3: bzip2 compressed (not used anymore)
		err = errors.New("zim: unsupported cluster compression")
	}
	return
}

func (z *File) lastClusterPosition() uint32 {
	return z.header.clusterCount - 1
}

// Cluster stores the uncompressed cluster data (blob positions followed by a sequence of blobs).
// Each blob belongs to a Directory Entry.
type Cluster struct {
	data        []byte // always uncompressed and len(data) <= 32MB
	position    uint32 // cluster position
	information uint8  // cluster information byte; stores information about compression and offset size
}

// WasCompressed shows if the cluster data was compressed.
// This information can be used as an indicator about the
// cluster contents.
func (c *Cluster) WasCompressed() bool {
	return clusterCompression(c.information) > 1
}

func (z *File) nextClusterPointer(c *Cluster) uint64 {
	if c.position >= z.lastClusterPosition() {
		return z.header.checksumPos - 1
	}
	return z.clusterPointerAtPos(c.position + 1)
}

// clusterLen returns the length of the cluster in bytes.
func (z *File) clusterLen(c *Cluster) int64 {
	var nextClusterPointer = z.nextClusterPointer(c)
	var clusterPointer = z.clusterPointerAtPos(c.position)
	seek(z.f, int64(clusterPointer)+1) // file position was (very likely) changed; seek back.
	// The +1 is because c.information byte was read afterwards too.
	return int64(nextClusterPointer - clusterPointer - 1)
}

// ClusterAt returns the Cluster of the ZIM file at the given cluster position.
// The complete cluster data is stored uncompressed in memory.
// If the size of the cluster data is more than 32MB an error is returned
// and the data is not read into memory.
// Note: Only use this function, when it's needed to read every single blob of a
// ZIM file into memory (for example when iterating over all contents this improves performance).
func (z *File) ClusterAt(clusterPosition uint32) (Cluster, error) {
	var c = Cluster{position: clusterPosition}
	var clusterLen = z.clusterLen(&c)
	if clusterLen <= 0 || clusterLen > maxClusterLen {
		return c, errors.New("zim: invalid cluster size")
	}
	var clusterReader, clusterInformation, clusterReaderErr = z.clusterReader(clusterPosition)
	c.information = clusterInformation
	if clusterReaderErr != nil {
		return c, clusterReaderErr
	}

	var clusterData, clusterDataErr = ioutil.ReadAll(clusterReader)

	if clusterDataErr != nil {
		return c, clusterDataErr
	}

	c.data = clusterData
	return c, nil
}

// BlobAt returns the blob data at blob position of a given Cluster.
// This is only useful when iteration over all blobs in a Cluster is done.
// When only a single blob of a Cluster should be retrieved, it's better
// to use z.BlobReaderAt(clusterPosition, blobPosition) instead.
// The blob position starts at 0 and ends if an error is returned.
func (c *Cluster) BlobAt(blobPosition uint32) ([]byte, error) {
	var offsetSize = uint64(clusterOffsetSize(c.information))
	var thisBlobIndex = uint64(blobPosition) * offsetSize
	var nextBlobIndex = thisBlobIndex + offsetSize
	if nextBlobIndex+offsetSize > uint64(len(c.data)) {
		return nil, errors.New("zim: invalid blob position")
	}
	var thisBlobPointer uint64
	var nextBlobPointer uint64
	if offsetSize == defaultOffsetSize {
		thisBlobPointer = uint64(binary.LittleEndian.Uint32(c.data[thisBlobIndex : thisBlobIndex+defaultOffsetSize]))
		nextBlobPointer = uint64(binary.LittleEndian.Uint32(c.data[nextBlobIndex : nextBlobIndex+defaultOffsetSize]))
	} else {
		thisBlobPointer = binary.LittleEndian.Uint64(c.data[thisBlobIndex : thisBlobIndex+extendedOffsetSize])
		nextBlobPointer = binary.LittleEndian.Uint64(c.data[nextBlobIndex : nextBlobIndex+extendedOffsetSize])
	}
	if nextBlobPointer >= thisBlobPointer && nextBlobPointer <= uint64(len(c.data)) {
		return c.data[thisBlobPointer:nextBlobPointer], nil
	}
	return nil, errors.New("zim: invalid blob index")
}
