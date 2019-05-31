package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/dps/go-zim"
)

func main() {

	var filename string
	var port int

	flag.StringVar(&filename, "filename", "", "Filename of the ZIM file to use.")
	flag.IntVar(&port, "port", 8080, "TCP port of the HTTP server.")

	flag.Parse()

	if flag.NFlag() < 1 || len(filename) == 0 {
		flag.PrintDefaults()
		return
	}

	if z, zimOpenErr := zim.Open(filename); zimOpenErr != nil {
		log.Fatal(zimOpenErr)
	} else {
		StartHTTPServer(z, uint16(port))
	}

}

// StartHTTPServer starts a HTTP server at localhost with given TCP port
// for browsing the ZIM file.
func StartHTTPServer(z *zim.File, port uint16) {

	var zimName = z.UUID().String()
	var zimNameLen = len(zimName)
	var urlSuffixMainpage []byte
	{
		var mainPageEntry, mainPageErr = z.MainPage()
		if mainPageErr != nil {
			log.Println("No Mainpage specified in ZIM file.")
		}
		urlSuffixMainpage = mainPageEntry.URL()
	}

	var urlPrefix = fmt.Sprintf("/%s/", zimName)

	var createURLFor = func(namespace zim.Namespace, suffix []byte) string {
		return fmt.Sprintf("%s%s/%s", urlPrefix, namespace, suffix)
	}

	fmt.Println(fmt.Sprintf(
		"Serving ZIM file at http://localhost:%d%s\n",
		port, createURLFor(zim.NamespaceArticles, urlSuffixMainpage)))

	if title := z.Title(); len(title) > 0 {
		fmt.Println(title)
	}

	if desc := z.Description(); len(desc) > 0 {
		fmt.Println(desc)
	}

	if date := z.Date(); len(date) > 0 {
		fmt.Println(date)
	}

	var mutex sync.Mutex

	log.Fatal(http.ListenAndServe(fmt.Sprint("localhost:", port), http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {

			if r.URL.Path == "/favicon.ico" {
				mutex.Lock()
				favicon, faviconErr := z.Favicon()
				mutex.Unlock()
				if faviconErr != nil {
					log.Println(faviconErr)
					http.NotFound(w, r)
				} else {
					http.Redirect(w, r, createURLFor(favicon.Namespace(), favicon.URL()), http.StatusFound)
				}
				return
			}
			if len(r.URL.Path) < zimNameLen+5 || !strings.HasPrefix(r.URL.Path, urlPrefix) ||
				r.URL.Path[zimNameLen+3] != '/' {
				http.Redirect(w, r, createURLFor(zim.NamespaceArticles, urlSuffixMainpage), http.StatusFound)
				return
			}

			// The URL now has a UUID so we can cache the result
			// also if it doesn't exist, since it won't change.
			w.Header().Set("Cache-Control", "max-age=87840, must-revalidate")

			var namespace = zim.Namespace(r.URL.Path[zimNameLen+2])
			switch namespace {
			case zim.NamespaceLayout, zim.NamespaceArticles, zim.NamespaceImagesFiles, zim.NamespaceImagesText:
				var suffix = []byte(r.URL.Path[zimNameLen+4:])
				mutex.Lock()
				var entry, _, found = z.EntryWithURL(namespace, suffix)
				mutex.Unlock()
				if found {
					if entry.IsRedirect() {
						mutex.Lock()
						entry, _ = z.FollowRedirect(&entry)
						mutex.Unlock()
						http.Redirect(w, r, createURLFor(entry.Namespace(), entry.URL()), http.StatusFound)
						return
					}
					mutex.Lock()
					var blobReader, _, blobReaderErr = z.BlobReader(&entry)
					if blobReaderErr != nil {
						mutex.Unlock()
						log.Printf("Entry found but loading blob data failed for URL: %s with error %s\n", r.URL.Path, blobReaderErr)
						http.Error(w, blobReaderErr.Error(), http.StatusFailedDependency)
						return
					}
					var mimetypeList = z.MimetypeList()
					if int(entry.Mimetype()) < len(mimetypeList) {
						w.Header().Set("Content-Type", mimetypeList[entry.Mimetype()])
					}
					io.Copy(w, blobReader)
					mutex.Unlock()
					return
				}

				if namespace == zim.NamespaceArticles {
					mutex.Lock()
					var similarEntries = z.EntriesWithSimilarity(namespace, suffix, 100)
					mutex.Unlock()
					w.WriteHeader(http.StatusMultipleChoices)
					w.Write(htmlSuggestions(zimName, similarEntries))
					return
				}

				log.Printf("Entry not found for URL: %s\n", r.URL.Path)
			}
			http.NotFound(w, r)
		})))
}

func htmlSuggestions(zimUUID string, results []zim.DirectoryEntry) []byte {
	var body = make([]byte, 0, 1<<13) // responses with 100 suggestions mostly have size in range [1<<12, 1<<14]
	body = append(body, []byte(string("<!doctype html><html>"))...)
	for _, result := range results {
		if len(result.URL()) == 0 {
			continue
		}
		if result.IsArticle() || result.IsRedirect() {
			body = append(body, []byte(fmt.Sprintf("<a href=\"/%s/%s/%s\">%s</a><br>\n",
				zimUUID, string(result.Namespace()), result.URL(), result.Title()))...)
		}
	}
	body = append(body, []byte(string("</html>"))...)
	return body
}
