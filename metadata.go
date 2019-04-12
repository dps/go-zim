package zim

func (z *File) readMetadata() {
	const entryLimit = 256 // we don't want to fill the memory too much
	const maxKeySize = 128
	const maxValueSize = 2048
	z.metadata = make(map[string]string)
	for _, entry := range z.EntriesWithNamespace(NamespaceZimMetadata, entryLimit) {
		if len(entry.url) <= maxKeySize {
			var blobReader, blobSize, blobReaderErr = z.BlobReader(&entry)
			if blobReaderErr == nil && blobSize <= maxValueSize {
				var value = make([]byte, blobSize)
				if _, blobReadErr := blobReader.Read(value); blobReadErr == nil {
					z.metadata[string(entry.url)] = string(value)
				}
			}
		}
	}
}

// Metadata returns a copy of the internal metadata map of the ZIM file.
func (z *File) Metadata() map[string]string {
	var result = make(map[string]string, len(z.metadata))
	for k, v := range z.metadata {
		result[k] = v
	}
	return result
}

// MetadataFor returns the metadata value for a given key.
// If the key is not set, an empty string is returned.
func (z *File) MetadataFor(key string) string {
	return z.metadata[key]
}

// Name returns the Name of the ZIM file as found in the Metadata.
func (z *File) Name() string { return z.MetadataFor("Name") }

// Title returns the Title of the ZIM file as found in the Metadata.
func (z *File) Title() string { return z.MetadataFor("Title") }

// Creator returns the Creator of the ZIM file as found in the Metadata.
func (z *File) Creator() string { return z.MetadataFor("Creator") }

// Publisher returns the Publisher of the ZIM file as found in the Metadata.
func (z *File) Publisher() string { return z.MetadataFor("Publisher") }

// Date returns the Date of the ZIM file as found in the Metadata.
func (z *File) Date() string { return z.MetadataFor("Date") }

// Description returns the Description of the ZIM file as found in the Metadata.
func (z *File) Description() string { return z.MetadataFor("Description") }

// LongDescription returns the LongDescription of the ZIM file as found in the Metadata.
func (z *File) LongDescription() string { return z.MetadataFor("LongDescription") }

// Language returns the Language of the ZIM file as found in the Metadata.
func (z *File) Language() string { return z.MetadataFor("Language") }

// License returns the License of the ZIM file as found in the Metadata.
func (z *File) License() string { return z.MetadataFor("License") }

// Tags returns the Tags of the ZIM file as found in the Metadata.
func (z *File) Tags() string { return z.MetadataFor("Tags") }

// Relation returns the Relation of the ZIM file as found in the Metadata.
func (z *File) Relation() string { return z.MetadataFor("Relation") }

// Source returns the Source of the ZIM file as found in the Metadata.
func (z *File) Source() string { return z.MetadataFor("Source") }

// Counter returns a String containing the number of Directory Entries per Mimetype.
func (z *File) Counter() string { return z.MetadataFor("Counter") }
