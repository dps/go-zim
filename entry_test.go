package zim

import (
	"bytes"
	"log"
	"testing"
)

func TestMainPage(t *testing.T) {
	var mainPageEntry, mainPageErr = z.MainPage()
	if mainPageErr != nil {
		t.Error(mainPageErr)
	}
	blobReader, blobSize, blobReaderErr := z.BlobReader(&mainPageEntry)
	if blobReaderErr != nil {
		log.Fatal(blobReaderErr)
	}
	mainPageData := make([]byte, blobSize)
	blobReader.Read(mainPageData)
	const expectedDataLen = 1982
	if blobLen := len(mainPageData); blobLen != expectedDataLen {
		t.Errorf("Directory Entry of z.MainPage() has length %d; want %d\n", blobLen, expectedDataLen)
	}
	var expectedMainPageURL = []byte("index.htm")
	if bytes.Compare(mainPageEntry.url, expectedMainPageURL) != 0 {
		t.Errorf("URL of Directory Entry from z.MainPage() was `%s`; want `%s`\n", mainPageEntry.url, expectedMainPageURL)
	}
	var expectedMainPageTitle = []byte("Summary")
	if bytes.Compare(mainPageEntry.title, expectedMainPageTitle) != 0 {
		t.Errorf("Title of Directory Entry from z.MainPage() was `%s`; want `%s`\n", mainPageEntry.title, expectedMainPageTitle)
	}
	if mainPageEntry.namespace != NamespaceArticles {
		t.Errorf("Namespace of Directory Entry from z.MainPage() was `%s`; want `%s`\n",
			string(mainPageEntry.namespace), string(NamespaceArticles))
	}
	if !mainPageEntry.IsArticle() {
		t.Error("Directoy Entry from z.MainPage() was not detected as an Article.")
	}
	if mainPageEntry.IsRedirect() {
		t.Error("Directory Entry from z.MainPage() was wrongly detected as an Article.")
	}
	const expectedMimetype = Mimetype(7)
	if mainPageEntry.mimetype != expectedMimetype {
		t.Errorf("Mimetype from z.MainPage() was %d; want %d\n", mainPageEntry.mimetype, expectedMimetype)
	}
}

func TestFavicon(t *testing.T) {
	var faviconEntry, faviconErr = z.Favicon()
	if faviconErr != nil {
		t.Error(faviconErr)
	}
	if faviconEntry.IsRedirect() {
		faviconEntry, faviconErr = z.FollowRedirect(&faviconEntry)
	}
	if faviconErr != nil {
		t.Error(faviconErr)
	}
	blobReader, blobSize, blobReaderErr := z.BlobReader(&faviconEntry)
	if blobReaderErr != nil {
		log.Fatal(blobReaderErr)
	}
	faviconData := make([]byte, blobSize)
	blobReader.Read(faviconData)
	const expectedDataLen = 1376
	if blobLen := len(faviconData); blobLen != expectedDataLen {
		t.Errorf("Directory Entry of z.Favicon() has length %d; want %d\n", blobLen, expectedDataLen)
	}
}

func TestRedirects(t *testing.T) {
	var redirectEntryURL = []byte("Orbite_heliosynchrone.html")
	var redirectEntry, _, found = z.EntryWithURL(NamespaceArticles, redirectEntryURL)
	if !found {
		t.Errorf("Directory Entry with URL `%s` not found.", redirectEntryURL)
	}
	if !redirectEntry.IsRedirect() {
		t.Errorf("Directory Entry with URL `%s` not detected as Redirect Entry.", redirectEntryURL)
	}
	var targetEntry, redirectError = z.FollowRedirect(&redirectEntry)
	if redirectError != nil {
		t.Error(redirectError)
	}
	var expectedTargetURL = []byte("Orbite_héliosynchrone.html")
	if bytes.Compare(targetEntry.url, expectedTargetURL) != 0 {
		t.Errorf("Target Directory Entry has URL `%s`; want `%s`", targetEntry.url, expectedTargetURL)
	}
}

func TestDirectoryEntryAtURLPosition(t *testing.T) {
	var prevNamespace Namespace
	var prevURL []byte
	for position := uint32(0); ; position++ {
		var entry, entryErr = z.EntryAtURLPosition(position)

		if entryErr != nil {
			const expectedEndPosition = 189
			if z.ArticleCount() != position {
				t.Errorf("z.ArticleCount() != %d (end position)\n", expectedEndPosition)
			}
			break
		}

		if position != 0 {
			if int(prevNamespace) > int(entry.Namespace()) {
				t.Errorf("previous namespace > current namespace: %s > %s\n", prevNamespace, entry.Namespace())
			} else if prevNamespace == entry.Namespace() && bytes.Compare(prevURL, entry.URL()) == 1 {
				t.Errorf("previous  > current URL: %s > %s\n", string(prevURL), string(entry.URL()))
			}
		}

		prevNamespace = entry.Namespace()
		prevURL = entry.URL()

		if len(entry.URL()) == 0 {
			t.Errorf("entry.URL() has length 0")
		}
	}
}

func TestDirectoryEntryAtTitlePosition(t *testing.T) {
	var prevNamespace Namespace
	var prevTitle []byte
	for position := uint32(0); ; position++ {
		var entry, entryErr = z.EntryAtTitlePosition(position)

		if entryErr != nil {
			const expectedEndPosition = 189
			if z.ArticleCount() != position {
				t.Errorf("z.ArticleCount() != %d (end position)\n", expectedEndPosition)
			}
			break
		}

		if position != 0 {
			if int(prevNamespace) > int(entry.Namespace()) {
				t.Errorf("previous namespace > current namespace: %s > %s\n", prevNamespace, entry.Namespace())
			} else if prevNamespace == entry.Namespace() && bytes.Compare(prevTitle, entry.Title()) == 1 {
				t.Errorf("previous  > current Title: %s > %s\n", string(prevTitle), string(entry.Title()))
			}
		}

		prevNamespace = entry.Namespace()
		prevTitle = entry.Title()

		if len(entry.Title()) == 0 {
			t.Errorf("entry.Title() has length 0")
		}
	}
}

func TestEntryLookupAndPositions(t *testing.T) {
	var entries []DirectoryEntry

	for position := uint32(0); ; position++ {
		var entry, entryErr = z.EntryAtURLPosition(position)

		if entryErr != nil {
			break
		}
		entries = append(entries, entry)
	}

	for _, entry := range entries {
		entry1, urlPosition1, found1 := z.EntryWithURL(entry.Namespace(), entry.URL())
		if !found1 {
			t.Errorf("Entry not found by URL lookup: %s\n", entry1.String())
		}

		if entry.Namespace() != entry1.Namespace() {
			t.Errorf("Wrong entry found by URL lookup for %s: %s\n", entry.String(), entry1.String())
		}

		if bytes.Compare(entry.URL(), entry1.URL()) != 0 {
			t.Errorf("Wrong entry found by URL lookup for %s: %s\n", entry.String(), entry1.String())
		}

		entry2, urlPosition2, found2 := z.EntryWithURLPrefix(entry.Namespace(), entry.URL())
		if !found2 {
			t.Errorf("Entry not found by URL-Prefix lookup: %s\n", entry2.String())
		}

		if entry.Namespace() != entry2.Namespace() {
			t.Errorf("Wrong entry found by URL lookup for %s: %s\n", entry.String(), entry2.String())
		}

		if bytes.Compare(entry.URL(), entry2.URL()) != 0 {
			t.Errorf("Wrong entry found by URL lookup for %s: %s\n", entry.String(), entry2.String())
		}

		if urlPosition1 != urlPosition2 {
			t.Errorf("Lookup of z.EntryWithURL() and z.EntryWithURLPrefix() yielded different results for same entry: %s\n", entry.String())
		}

		entry3, err3 := z.EntryAtURLPosition(urlPosition1)

		if err3 != nil {
			t.Error(err3)
		}

		if entry.Namespace() != entry3.Namespace() {
			t.Errorf("Wrong entry found by URL position lookup for %s: %s\n", entry.String(), entry3.String())
		}

		if bytes.Compare(entry.URL(), entry3.URL()) != 0 {
			t.Errorf("Wrong entry found by URL position lookup for %s: %s\n", entry.String(), entry3.String())
		}

	}

}

func TestEntryInNamespace(t *testing.T) {
	for _, namespace := range []Namespace{
		NamespaceLayout,
		NamespaceArticles,
		NamespaceImagesFiles,
		NamespaceZimMetadata,
	} {
		entry, urlPosition, found := z.EntryWithNamespace(namespace)
		if !found {
			t.Errorf("z.EntryWithNamespace() couldn't find namespace %s\n", namespace)
		}
		if entry.Namespace() != namespace {
			t.Errorf("z.EntryWithNamespace() couldn't find correct entry\n")
		}
		if urlPosition != 0 {
			entry2, err := z.EntryAtURLPosition(urlPosition - 1)
			if err != nil {
				t.Error(err)
			}
			if entry2.Namespace() >= namespace {
				t.Error("z.EntryWithNamespace() couldn't find first entry in namespace")
			}
		}
	}
}
