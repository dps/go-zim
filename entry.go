package zim

import (
	"errors"
	"fmt"
)

// DirectoryEntry holds the information about a specific article, image or other object in a ZIM file.
type DirectoryEntry struct {
	mimetype                  Mimetype
	parameterLen              byte
	namespace                 Namespace
	revision                  uint32
	clusterNumber             uint32
	blobNumberOrRedirectIndex uint32
	url                       []byte
	title                     []byte
	//parameter     []byte    // (not used) extra parameters; see `parameterLen`
}

func (e *DirectoryEntry) String() string {
	return fmt.Sprintf("DirectoryEntry{Namespace: %s, Title: %s, URL: %s}",
		e.namespace, string(e.title), string(e.url))
}

// Mimetype is the Mimetype of the Directory Entry.
func (e *DirectoryEntry) Mimetype() Mimetype {
	return e.mimetype
}

// Namespace defines to which namespace the Directory Entry belongs.
func (e *DirectoryEntry) Namespace() Namespace {
	return e.namespace
}

// Revision identifies a revision of the contents of the Directory Entry,
// needed to identify updates or revisions in the original history.
func (e *DirectoryEntry) Revision() uint32 {
	return e.revision
}

// ClusterNumber in which the data of this Directory Entry is stored.
func (e *DirectoryEntry) ClusterNumber() uint32 {
	return e.clusterNumber
}

// BlobNumber is the blob number inside the uncompressed cluster, where the contents are stored.
func (e *DirectoryEntry) BlobNumber() uint32 {
	return e.blobNumberOrRedirectIndex
}

// RedirectIndex is a pointer to the Directory Entry of the Redirect Target.
func (e *DirectoryEntry) RedirectIndex() uint32 {
	return e.blobNumberOrRedirectIndex
}

// URL is the URL of the Directory Entry, which is unique for the specific Namespace.
func (e *DirectoryEntry) URL() []byte {
	return e.url
}

// Title is the title of the Directory Entry.
func (e *DirectoryEntry) Title() []byte {
	if len(e.title) > 0 {
		return e.title
	}
	return e.url
}

func (z *File) readDirectoryEntry(filePosition uint64, maxRedirects uint8) DirectoryEntry {
	var result = DirectoryEntry{}
	seek(z.f, int64(filePosition))
	result.mimetype = Mimetype(readUint16(z.f))
	result.parameterLen = readUint8(z.f)
	result.namespace = Namespace(readUint8(z.f))
	result.revision = readUint32(z.f)
	switch result.mimetype {
	case MimetypeDeletedEntry, MimetypeLinkTarget:
		// no extra fields here
	case MimetypeRedirectEntry:
		result.blobNumberOrRedirectIndex = readUint32(z.f) // redirectIndex
		if maxRedirects > 0 {
			return z.readDirectoryEntry(z.urlPointerAtPos(result.RedirectIndex()), maxRedirects-1)
		}
	default:
		// Mimetype: ArticleEntry
		result.clusterNumber = readUint32(z.f)
		result.blobNumberOrRedirectIndex = readUint32(z.f) // blobNumber
	}
	result.url = readNullTerminatedSlice(z.f)
	result.title = readNullTerminatedSlice(z.f)
	//if result.parameterLen > 0 {
	//var buf, readErr = readSlice(z.f, int(result.parameterLen))
	//if readErr != nil {
	//	log.Printf("Error reading parameter data of length %d: %s\n", result.parameterLen, readErr)
	//} else {
	//	result.parameter = string(buf)
	//}
	//}
	return result
}

// EntryAtURLPosition returns the Directory Entry
// at the position as defined in the ordered URL pointerlist.
// If 0 >= position < z.ArticleCount() the returned error is nil.
// Redirects are not followed automatically.
func (z *File) EntryAtURLPosition(position uint32) (DirectoryEntry, error) {
	if position >= z.header.articleCount {
		return DirectoryEntry{}, errors.New("zim: position out of range")
	}
	return z.readDirectoryEntry(z.urlPointerAtPos(position), 0), nil
}

// EntryAtTitlePosition returns the Directory Entry
// at the position as defined in the ordered title pointerlist.
// If 0 >= position < z.ArticleCount() the returned error is nil.
// Redirects are not followed automatically.
func (z *File) EntryAtTitlePosition(position uint32) (DirectoryEntry, error) {
	if position >= z.header.articleCount {
		return DirectoryEntry{}, errors.New("zim: position out of range")
	}
	return z.readDirectoryEntry(z.titlePointerAtPos(position), 0), nil
}

// IsArticle checks whether the Directory Entry is an Article
func (e *DirectoryEntry) IsArticle() bool {
	return e.namespace == NamespaceArticles && e.mimetype < MimetypeDeletedEntry
}

// IsRedirect checks whether the Directory Entry is a Redirect
// to another Directory Entry
func (e *DirectoryEntry) IsRedirect() bool {
	return e.mimetype == MimetypeRedirectEntry
}

// IsLinkTarget checks whether the Directory Entry is a LinkTarget
func (e *DirectoryEntry) IsLinkTarget() bool {
	return e.mimetype == MimetypeLinkTarget
}

// IsDeletedEntry checks whether the Directory Entry is a DeletedEntry
func (e *DirectoryEntry) IsDeletedEntry() bool {
	return e.mimetype == MimetypeDeletedEntry
}

// FollowRedirect returns the target Directory Entry of the given Redirect Entry
func (z *File) FollowRedirect(redirectEntry *DirectoryEntry) (DirectoryEntry, error) {
	if !redirectEntry.IsRedirect() {
		return *redirectEntry, errors.New("zim: Directory Entry is not a Redirect Entry")
	}
	return z.readDirectoryEntry(z.urlPointerAtPos(redirectEntry.RedirectIndex()), 4), nil
}

// MainPage returns the Directory Entry for the MainPage of the ZIM file
func (z *File) MainPage() (DirectoryEntry, error) {
	if z.header.mainPage == NoMainPage {
		return DirectoryEntry{
			namespace: NamespaceArticles,
			url:       []byte("index.html"),
		}, errors.New("zim: no main page specified in ZIM file")
	}
	return z.readDirectoryEntry(z.urlPointerAtPos(z.header.mainPage), 4), nil
}

// LayoutPage returns the Directory Entry for the LayoutPage of the ZIM file
func (z *File) LayoutPage() (DirectoryEntry, error) {
	if z.header.layoutPage == NoLayoutPage {
		mainPage, _ := z.MainPage()
		return mainPage, errors.New("zim: no layout page specified in ZIM file")
	}
	return z.readDirectoryEntry(z.urlPointerAtPos(z.header.layoutPage), 4), nil
}

// Favicon returns the Directory Entry for the Favicon of the ZIM file
func (z *File) Favicon() (entry DirectoryEntry, err error) {
	for _, namespace := range [2]Namespace{NamespaceLayout, NamespaceImagesFiles} {
		var found bool
		entry, _, found = z.EntryWithURL(namespace, []byte("favicon"))
		if !found {
			entry, _, found = z.EntryWithURL(namespace, []byte("favicon.png"))
		}
		if found {
			if entry.IsRedirect() {
				entry, err = z.FollowRedirect(&entry)
			}
			return
		}
	}
	err = errors.New("zim: favicon not found")
	return
}
