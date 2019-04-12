package zim

import (
	"crypto/md5"
	"errors"
	"io"
)

// InternalChecksum is the MD5 checksum for the ZIM file.
// It's precalculated and saved in the header.
func (z *File) InternalChecksum() ([md5.Size]byte, error) {
	seek(z.f, int64(z.header.checksumPos))
	var buf, readErr = readSlice(z.f, md5.Size)
	var md5sum [md5.Size]byte
	if readErr != nil {
		return md5sum, errors.New("zim: reading internal checksum failed")
	}
	copy(md5sum[:], buf)
	return md5sum, nil
}

// CalculateChecksum calculates the MD5 checksum of the ZIM file.
// This could take some time dependent on the size of the file.
func (z *File) CalculateChecksum() ([md5.Size]byte, error) {
	var digest = md5.New()
	var md5Sum [md5.Size]byte
	seek(z.f, 0)
	if _, copyErr := io.CopyN(digest, z.f, int64(z.header.checksumPos)); copyErr != nil {
		return md5Sum, copyErr
	}
	copy(md5Sum[:], digest.Sum(nil)[:md5.Size])
	return md5Sum, nil
}

// ValidateChecksum compares the internal MD5 checksum
// of the ZIM file with the calculated one.
func (z *File) ValidateChecksum() error {
	if internal, internalChecksumErr := z.InternalChecksum(); internalChecksumErr != nil {
		return internalChecksumErr
	} else if calculated, calculatedChecksumErr := z.CalculateChecksum(); calculatedChecksumErr != nil {
		return calculatedChecksumErr
	} else if internal != calculated {
		return errors.New("zim: checksum mismatched")
	} else {
		return nil
	}
}
