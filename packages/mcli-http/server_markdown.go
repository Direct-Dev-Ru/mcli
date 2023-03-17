package mclihttp

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/html"
	"github.com/microcosm-cc/bluemonday"
)

type MyRenderer struct {
	html.Renderer
}

// an actual rendering of Paragraph is more complicated
func renderParagraph(w io.Writer, p *ast.Paragraph, entering bool) {
	if entering {
		io.WriteString(w, `<div class="md-paragraph">`)
	} else {
		io.WriteString(w, "</div>")
	}
}

func myRenderHook(w io.Writer, node ast.Node, entering bool) (ast.WalkStatus, bool) {
	if para, ok := node.(*ast.Paragraph); ok {
		renderParagraph(w, para, entering)
		return ast.GoToNext, true
	}
	return ast.GoToNext, false
}

func newCustomizedRender() *html.Renderer {
	opts := html.RendererOptions{
		Flags:          html.CommonFlags,
		RenderNodeHook: myRenderHook,
	}
	return html.NewRenderer(opts)
}

func ConvertMdToHtml(source string) ([]byte, error) {
	if len(source) == 0 {
		return nil, fmt.Errorf("%v", "zero length source")
	}
	if strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://") {
		return nil, fmt.Errorf("%v", "uri source don't support")
		// _, e := url.ParseRequestURI(source)
		// if e != nil {
		// 	return nil, fmt.Errorf("%v", e)
		// }
		//  ....
	}

	if isExist, e := exists(source); !isExist {
		return nil, e
	}
	// Read the Markdown file
	mddata, e := os.ReadFile(source)
	if e != nil {
		return nil, e
	}
	mddata = markdown.NormalizeNewlines(mddata)

	renderer := newCustomizedRender()
	maybeUnsafeHTML := markdown.ToHTML(mddata, nil, renderer)
	html := bluemonday.UGCPolicy().SanitizeBytes(maybeUnsafeHTML)

	return html, nil

}
