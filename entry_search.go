package zim

import (
	"bytes"
	"hash/fnv"
)

const defaultLimitEntries = 100

// EntryWithURL searches for the Directory Entry with the exact URL.
// If the Directory Entry was found, found is set to true and
// the returned position will be the position in the URL pointerlist.
// This can be used to iterate over the next n Directory Entries using
// z.EntryAtURLPosition(position+n).
func (z *File) EntryWithURL(namespace Namespace, url []byte) (
	entry DirectoryEntry, urlPosition uint32, found bool) {
	// more optimized version of entryWithPrefix
	var firstURLPosition int64
	var currentURLPos int64
	var lastURLPosition = int64(z.header.articleCount - 1)
	for firstURLPosition <= lastURLPosition {
		currentURLPos = (firstURLPosition + lastURLPosition) >> 1
		entry = z.readDirectoryEntry(z.urlPointerAtPos(uint32(currentURLPos)), 0)
		var c = cmpNs(entry.namespace, namespace)
		if c == 0 {
			c = bytes.Compare(entry.url, url)
			if c == 0 {
				found = true
				break
			}
		}
		if c < 0 {
			firstURLPosition = currentURLPos + 1
		} else {
			lastURLPosition = currentURLPos - 1
		}
	}
	urlPosition = uint32(currentURLPos)
	return
}

// EntryWithURLPrefix searches the first Directory Entry in the namespace
// having the given URL prefix. If it was found, found is set to true and
// the returned position will be the position in the URL pointerlist.
// This can be used to iterate over the next n Directory Entries using
// z.EntryAtURLPosition(position+n).
func (z *File) EntryWithURLPrefix(namespace Namespace, prefix []byte) (
	entry DirectoryEntry, position uint32, found bool) {
	return z.entryWithPrefix(z.urlPointerAtPos, chooseURL, namespace, prefix)
}

// EntryWithNamespace searches the first Directory Entry in the namespace.
// If it was found, found is set to true and the returned position will
// be the position in the URL pointerlist.
// This can be used to iterate over the next n Directory Entries using
// z.EntryAtURLPosition(position+n).
func (z *File) EntryWithNamespace(namespace Namespace) (
	entry DirectoryEntry, position uint32, found bool) {
	return z.EntryWithURLPrefix(namespace, nil)
}

// EntryWithTitlePrefix searches the first Directory Entry in the namespace
// having the given title prefix. If it was found, found is set to true and
// the returned position will be the position in the title pointerlist.
// This can be used to iterate over the next n Directory Entries using
// z.EntryAtTitlePosition(position+n).
func (z *File) EntryWithTitlePrefix(namespace Namespace, prefix []byte) (
	entry DirectoryEntry, position uint32, found bool) {
	return z.entryWithPrefix(z.titlePointerAtPos, chooseTitle, namespace, prefix)
}

// EntriesWithURLPrefix returns all Directory Entries in the Namespace
// that have the same URL prefix like the given.
// When the Limit is set to <= 0 it gets the default value 100.
func (z *File) EntriesWithURLPrefix(namespace Namespace, prefix []byte, limit int) []DirectoryEntry {
	return z.entriesWithPrefix(chooseURL, z.urlPointerAtPos, namespace, prefix, limit)
}

// EntriesWithNamespace returns the first n Directory Entries in the Namespace
// where n <= limit.
// When the Limit is set to <= 0 it gets the default value 100.
func (z *File) EntriesWithNamespace(namespace Namespace, limit int) []DirectoryEntry {
	return z.EntriesWithURLPrefix(namespace, nil, limit)
}

// EntriesWithTitlePrefix returns all Directory Entries in the Namespace
// that have the same Title prefix like the given.
// When the Limit is set to <= 0 it gets the default value 100.
func (z *File) EntriesWithTitlePrefix(namespace Namespace, prefix []byte, limit int) []DirectoryEntry {
	return z.entriesWithPrefix(chooseTitle, z.titlePointerAtPos, namespace, prefix, limit)
}

// EntriesWithSimilarity returns Directory Entries in the Namespace
// that have a similar URL prefix or Title prefix to the given one.
// When the Limit is set to <= 0 it takes the default value 100.
func (z *File) EntriesWithSimilarity(namespace Namespace, prefix []byte, limit int) []DirectoryEntry {
	const maxLengthDifference = 15
	type wasSuggested = struct{}
	var alreadySuggested = make(map[uint32]wasSuggested, limit)
	var suggestions = make([]DirectoryEntry, 0, limit)
	for i := 0; i < maxLengthDifference; i++ {

		for _, prefixFunc := range [2]func(Namespace, []byte, int) []DirectoryEntry{
			z.EntriesWithURLPrefix, z.EntriesWithTitlePrefix} {
			var nextSuggestions = prefixFunc(namespace, prefix, limit)
			for _, suggestion := range nextSuggestions {
				var key = hash(suggestion.url)
				var _, suggestedBefore = alreadySuggested[key]
				if !suggestedBefore {
					suggestions = append(suggestions, suggestion)
					alreadySuggested[key] = wasSuggested{}
					if len(suggestions) >= limit {
						return suggestions
					}
				}
			}

		}

		if len(prefix) == 0 {
			return suggestions
		}

		prefix = prefix[:len(prefix)-1]
	}

	return suggestions
}

func chooseTitle(entry *DirectoryEntry) []byte { return entry.title }

func chooseURL(entry *DirectoryEntry) []byte { return entry.url }

func hash(data []byte) uint32 {
	h := fnv.New32a()
	h.Write(data)
	return h.Sum32()
}

func cmpNs(ns1, ns2 Namespace) int {
	if ns1 > ns2 {
		return 1
	}
	if ns2 > ns1 {
		return -1
	}
	return 0
}

func cmpPrefix(s, prefix []byte) int {
	if bytes.HasPrefix(s, prefix) {
		return 0
	}
	return bytes.Compare(s, prefix)
}

func (z *File) entryWithPrefix(
	pointerAtPosition func(uint32) uint64,
	chooseField func(entry *DirectoryEntry) []byte,
	namespace Namespace,
	prefix []byte) (
	entry DirectoryEntry, position uint32, found bool) {
	var firstPosition int64
	var currentPosition int64
	var lastPosition = int64(z.header.articleCount - 1)
	for firstPosition <= lastPosition {
		currentPosition = (firstPosition + lastPosition) >> 1
		entry = z.readDirectoryEntry(pointerAtPosition(uint32(currentPosition)), 0)
		var c = cmpNs(entry.namespace, namespace)
		if c == 0 {
			c = cmpPrefix(chooseField(&entry), prefix)
			if c == 0 {
				// we found an entry with the given prefix
				if currentPosition == 0 {
					// already lowest position
					found = true
					break
				}
				var prevEntry = z.readDirectoryEntry(pointerAtPosition(uint32(currentPosition-1)), 0)
				if prevEntry.namespace != namespace || !bytes.HasPrefix(chooseField(&prevEntry), prefix) {
					// we found the lowest position
					found = true
					break
				}
				// the entry below also has the prefix, but maybe much more entries have it too...
				c = 1 // so the current entry is greater
			}
		}
		if c < 0 {
			firstPosition = currentPosition + 1
		} else {
			lastPosition = currentPosition - 1
		}
	}
	position = uint32(currentPosition)
	return
}

func (z *File) entriesWithPrefix(
	chooseField func(*DirectoryEntry) []byte,
	pointerAtPosition func(uint32) uint64,
	namespace Namespace,
	prefix []byte,
	limit int) []DirectoryEntry {

	if limit <= 0 {
		limit = defaultLimitEntries
	}
	var cap int
	if limit <= defaultLimitEntries {
		cap = limit
	} else {
		cap = defaultLimitEntries
	}
	var entry, position, found = z.entryWithPrefix(pointerAtPosition, chooseField, namespace, prefix)
	var result []DirectoryEntry
	if found {
		result = make([]DirectoryEntry, 0, cap)
		result = append(result, entry)
		var entriesAdded = 1
		var lastPosition = z.header.articleCount - 1
		for entriesAdded < limit && position < lastPosition {
			position++
			var nextEntry = z.readDirectoryEntry(pointerAtPosition(position), 0)
			if !bytes.HasPrefix(chooseField(&nextEntry), prefix) {
				break
			}
			result = append(result, nextEntry)
			entriesAdded++
		}
	}
	return result
}
