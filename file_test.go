package zim

import (
	"path"
	"testing"
)

const filenameTestfile = `wikipedia_fr_test_2018-10.zim`

var z *File
var openErr error

func init() {
	z, openErr = Open(path.Join("testdata", filenameTestfile))
}

func TestOpen(t *testing.T) {
	if openErr != nil {
		t.Error(openErr)
	}
}

func TestArticleCount(t *testing.T) {
	const expectedArticleCount = 189
	if count := z.ArticleCount(); count != expectedArticleCount {
		t.Errorf("z.ArticleCount() = %d; want %d", count, expectedArticleCount)
	}
}

func TestClusterCount(t *testing.T) {
	const expectedClusterCount = 2
	if count := z.ClusterCount(); count != expectedClusterCount {
		t.Errorf("z.ClusterCount() = %d; want %d", count, expectedClusterCount)
	}
}

func TestFilesize(t *testing.T) {
	const expectedFilesize = 792439
	if fs := z.Filesize(); fs != expectedFilesize {
		t.Errorf("z.Filesize() was %d; want %d", fs, expectedFilesize)
	}
}

func TestUUID(t *testing.T) {
	const expectedUUID = "017f96d106e20a91626e2f5cfebb50e2"
	if UUID := z.UUID().String(); UUID != expectedUUID {
		t.Errorf("z.UUID().String() was %s; want %s", UUID, expectedUUID)
	}
}
