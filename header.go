package zim

import (
	"encoding/hex"
	"errors"
)

const uuidLen = 16

// UUID is the unique ID of a ZIM file
type UUID []byte

func (u UUID) String() string {
	return hex.EncodeToString(u)
}

// Header is the header of a ZIM file
type Header struct {
	magicNumber   uint32 // Magic number to recognize the file format, must be 72173914
	majorVersion  uint16 // Major version of the ZIM file format (5 or 6)
	minorVersion  uint16 // Minor version of the ZIM file format
	uuid          UUID   // unique ID of the ZIM file
	articleCount  uint32 // total number of articles
	clusterCount  uint32 // total number of clusters
	urlPtrPos     uint64 // position of the directory pointerlist ordered by URL
	titlePtrPos   uint64 // position of the directory pointerlist ordered by Title
	clusterPtrPos uint64 // position of the cluster pointer list
	mimeListPos   uint64 // position of the MIME type list (also header size)
	mainPage      uint32 // main page or 0xffffffff if no main page
	layoutPage    uint32 // layout page or 0xffffffff if no layout page
	checksumPos   uint64 // pointer to the MD5 checksum of the ZIM file without the checksum itself. This points always 16 bytes before the end of the file.
}

func (z *File) readHeader() error {
	seek(z.f, 0)
	z.header.magicNumber = readUint32(z.f)
	if z.header.magicNumber != MagicNumber {
		return errors.New("zim: file has no ZIM header")
	}
	z.header.majorVersion = readUint16(z.f)
	z.header.minorVersion = readUint16(z.f)
	var uuidReadErr error
	z.header.uuid, uuidReadErr = readSlice(z.f, uuidLen)
	if uuidReadErr != nil {
		return uuidReadErr
	}
	z.header.articleCount = readUint32(z.f)
	z.header.clusterCount = readUint32(z.f)
	z.header.urlPtrPos = readUint64(z.f)
	z.header.titlePtrPos = readUint64(z.f)
	z.header.clusterPtrPos = readUint64(z.f)
	z.header.mimeListPos = readUint64(z.f)
	z.header.mainPage = readUint32(z.f)
	z.header.layoutPage = readUint32(z.f)
	z.header.checksumPos = readUint64(z.f)

	switch z.header.majorVersion {
	case 5, 6:
		return nil
	default:
		return errors.New("zim: version currently not supported")
	}
}
