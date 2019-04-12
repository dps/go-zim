package zim

import (
	"fmt"
	"testing"
)

func TestMimetypeList(t *testing.T) {
	var expectedMimetypeList = []string{
		"application/javascript",
		"application/octet-stream+xapian",
		"image/gif",
		"image/jpeg",
		"image/png",
		"image/svg+xml",
		"text/css",
		"text/html",
		"text/plain",
		"webm",
	}
	s1 := fmt.Sprint(z.MimetypeList())
	s2 := fmt.Sprint(expectedMimetypeList)
	if len(z.MimetypeList()) != len(expectedMimetypeList) || s1 != s2 {
		t.Errorf("z.MimetypeList() was `%s`; want `%s`", s1, s2)
	}
}
