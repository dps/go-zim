package zim

// Namespace is an Ascii-Character representing the category of a Directory Entry.
type Namespace byte

// Possible values for a Namespace.
const (
	NamespaceLayout                           = Namespace('-') // layout, eg. the LayoutPage, CSS, favicon.png (48x48), JavaScript and images not related to the articles
	NamespaceArticles                         = Namespace('A')
	NamespaceArticleMetadata                  = Namespace('B')
	NamespaceImagesFiles                      = Namespace('I')
	NamespaceImagesText                       = Namespace('J')
	NamespaceZimMetadata                      = Namespace('M')
	NamespaceCategoriesText                   = Namespace('U')
	NamespaceCategoriesArticleList            = Namespace('V')
	NamespaceCategoriesPerArticleCategoryList = Namespace('W')
	NamespaceFulltextIndex                    = Namespace('X') // fulltext index in "ZIM Index Format"
)

func (n Namespace) String() string { return string(n) }
