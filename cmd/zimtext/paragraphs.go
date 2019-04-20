package main

import (
	"bufio"
	"io"
	"log"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/net/html"
)

// Link stores the title and href of an <a>...</a> element that was
// parsed in a Paragraph.
// Note: the Href doesn't always start with "http".
type Link struct {
	Title      string
	Href       string
	IsExternal bool
}

// IsRedLink finds out, if the Link points to a non-existent page.
// When the link is a redlink the link title may have additional information.
func (l Link) IsRedLink() bool {
	return len(l.Href) == 0 || strings.Index(l.Href, "redlink=") > 0 && strings.Index(l.Href, "action=edit") > 0
}

// Paragraph represents the parsed <p>...</p> block of a MediaWiki HTML source
// with text, parsed links, and information about whether specific
// elements appear in the text.
type Paragraph struct {
	Text              string
	Links             []Link
	HasBoldText       bool
	HasItalicText     bool
	HasUnderlinedText bool
	HasSubscript      bool
	HasSuperscript    bool
	HasMath           bool
	HasCode           bool
	HasQuotation      bool
	HasBlockquote     bool
	HasPreTag         bool
	HasNowikiTag      bool
	HasSpanTag        bool
	HasCiteTag        bool
	HasOtherTags      bool
}

func isValidFirstRune(r rune) bool {
	return unicode.IsNumber(r) || (unicode.IsLetter(r) && !unicode.IsLower(r))
}

func isValidLastRune(r rune) bool {
	switch r {
	case '.', '?', '!', '"', '…', '«', '»', '›', '‹', '‘', '“', ')', '’', '”', ']':
		return true
	default:
		return false
	}
}

// HasCleanText checks whether the paragraph likely has clean, readable text.
func (p *Paragraph) HasCleanText() bool {
	if len(p.Text) > 0 && !p.HasMath && !p.HasCode &&
		!p.HasBlockquote && !p.HasQuotation && !p.HasNowikiTag &&
		!p.HasPreTag && !p.HasOtherTags && !p.HasCiteTag && !p.HasSpanTag {
		var lastRune, _ = utf8.DecodeLastRuneInString(p.Text)
		if isValidLastRune(lastRune) {
			var firstRune, _ = utf8.DecodeRuneInString(p.Text)
			return isValidFirstRune(firstRune)
		}
	}
	return false
}

// "golang.org/x/net/html" doesn't include it?
func skipStartTag(tokenizer *html.Tokenizer, tagName string) (finished bool, lastError error) {
	var tokenType html.TokenType
	var numberNestedSameName = 0
	for {
		tokenType = tokenizer.Next()
		if tokenType == html.StartTagToken {
			if tokenizer.Token().Data == tagName {
				numberNestedSameName++
			}
		} else if tokenType == html.EndTagToken {
			if tokenizer.Token().Data == tagName {
				if numberNestedSameName > 0 {
					numberNestedSameName--
				} else {
					// we skipped the tag that was originally opened
					finished = true
					break
				}
			}
		} else if tokenType == html.ErrorToken {
			lastError = tokenizer.Err()
			if lastError == io.EOF {
				break
			}
			log.Println(lastError)
		}
	}
	return
}

// WriteParagraphs writes the text within HTML paragraphs of the source Reader to the target Writer
// if the takeParagraph function returns true for the given paragraph.
// The number of written paragraphs is returned.
func WriteParagraphs(htmlSrc io.Reader, target *bufio.Writer, takeParagraph func(*Paragraph) bool, limit int) int {
	if limit <= 0 {
		limit = int(^uint(0) >> 1)
	}
	var tokenizer = html.NewTokenizer(htmlSrc)
	var currentParagraph Paragraph
	var paragraphsWritten = 0
	for {
		var tokenType = tokenizer.Next()
		if tokenType == html.ErrorToken {
			var err = tokenizer.Err()
			if err == io.EOF {
				return paragraphsWritten
			}
			log.Println(err)
		} else if tokenType == html.StartTagToken {
			var token = tokenizer.Token()
			if token.Data == "p" {
				for {
					tokenType = tokenizer.Next()
					if tokenType == html.TextToken {
						currentParagraph.Text += tokenizer.Token().Data
					} else if tokenType == html.StartTagToken {
						var token = tokenizer.Token()
						var skipTag = false
						for _, a := range token.Attr {
							if a.Key == "class" {
								if a.Val == "reference" {
									skipTag = true
									break
								}
								if strings.HasPrefix(a.Val, "noprint") || strings.HasPrefix(a.Val, "no print") {
									skipTag = true
									break
								}
							}
						}
						if skipTag {
							var _, lastError = skipStartTag(tokenizer, token.Data)
							if lastError != nil {
								if lastError == io.EOF {
									return paragraphsWritten
								}
								log.Println(lastError)
							}
						} else {
							switch token.Data {
							case "a":
								// we have read an url inside a paragraph
								var link Link
								for _, a := range token.Attr {
									switch a.Key {
									case "href":
										link.Href = a.Val
									case "title":
										link.Title = strings.TrimSpace(a.Val)
									case "class":
										link.IsExternal = strings.HasPrefix(a.Val, "external")
									}
								}
								//if len(link.Title) > 0 {
								currentParagraph.Links = append(currentParagraph.Links, link)
								//}
							case "b":
								currentParagraph.HasBoldText = true
							case "blockquote":
								currentParagraph.HasBlockquote = true
							case "cite":
								currentParagraph.HasCiteTag = true
							case "code":
								currentParagraph.HasCode = true
							case "i":
								currentParagraph.HasItalicText = true
							case "math":
								currentParagraph.HasMath = true
							case "nowiki":
								currentParagraph.HasNowikiTag = true
							case "pre":
								currentParagraph.HasPreTag = true
							case "q":
								currentParagraph.HasQuotation = true
							case "span":
								currentParagraph.HasSpanTag = true
							case "sub":
								currentParagraph.HasSubscript = true
							case "sup":
								currentParagraph.HasSuperscript = true
							case "u", "ins":
								currentParagraph.HasUnderlinedText = true
							default:
								currentParagraph.HasOtherTags = true
							}
						}
					} else if tokenType == html.EndTagToken {
						var token = tokenizer.Token()
						if token.Data == "p" {
							currentParagraph.Text = strings.TrimSpace(currentParagraph.Text)
							if len(currentParagraph.Text) > 0 {
								if takeParagraph(&currentParagraph) {
									if _, err := target.WriteString(currentParagraph.Text); err != nil {
										log.Fatal(err)
									}
									target.WriteByte('\n')
									paragraphsWritten += strings.Count(currentParagraph.Text, "\n") + 1
									if paragraphsWritten >= limit {
										return paragraphsWritten
									}
								}
								currentParagraph = Paragraph{}
							}
							break
						}
					} else if tokenType == html.ErrorToken {
						var err = tokenizer.Err()
						if err == io.EOF {
							return paragraphsWritten
						}
						log.Println(err)
					}
				}
			} else if token.Data == "table" || token.Data == "blockquote" || token.Data == "q" {
				// skip relevant tags if they start in the outer scope
				var _, lastError = skipStartTag(tokenizer, token.Data)
				if lastError != nil {
					if lastError == io.EOF {
						return paragraphsWritten
					}
					log.Println(lastError)
				}
			}
		}
	}
}

// WriteCleanText analyzes the paragraphs from the HTML source reader
// and writes the text to the writer if it's likely clean readable text.
func WriteCleanText(htmlSrc io.Reader, target *bufio.Writer, limit int) int {
	return WriteParagraphs(htmlSrc, target, func(p *Paragraph) bool {
		return p.HasCleanText()
	}, limit)
}

// WriteCleanSentences analyzes the paragraphs from the HTML source reader
// and writes it to the writer if the text is likely a single sentence.
func WriteCleanSentences(htmlSrc io.Reader, target *bufio.Writer, limit int) int {
	return WriteParagraphs(htmlSrc, target, func(p *Paragraph) bool {
		return len(p.Text) >= 12 && strings.HasSuffix(p.Text, ".") && p.HasCleanText() &&
			strings.Count(p.Text, ".") == 1 && strings.Count(p.Text, "\n") == 0
	}, limit)
}
