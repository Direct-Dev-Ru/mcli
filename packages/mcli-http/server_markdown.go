package mclihttp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/html"
	"github.com/microcosm-cc/bluemonday"
	"gopkg.in/yaml.v3"
)

type MyRenderer struct {
	html.Renderer
}
type MdMetaData struct {
	IsTemplate       bool   `yaml:"is-template"`
	TemplateDataPath string `yaml:"template-data-path"`
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
	var mddata []byte
	var e error

	// if source is internet resource
	if strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://") {
		// return nil, fmt.Errorf("%v", "uri source don't support")

		// Make the HTTP GET request
		response, err := http.Get(source)
		if err != nil {

			return nil, err
		}
		defer response.Body.Close()

		// Check the response status code
		if response.StatusCode != http.StatusOK {
			err = fmt.Errorf("unexpected status code: %v", response.StatusCode)
			return nil, err
		}

		// Read the response body into a byte slice
		mddata, err = io.ReadAll(response.Body)
		if err != nil {
			return nil, err
		}
		// os.WriteFile("temp.md", bodyBytes, 0644)
		// return bodyBytes, nil
	} else {
		// is source is file
		if isExist, e := exists(source); !isExist {
			return nil, e
		}
		// Read the Markdown file
		mddata, e = os.ReadFile(source)
		if e != nil {
			return nil, e
		}
	}
	startMarker := []byte("<!--")
	endMarker := []byte("-->")

	startIndex := bytes.Index(mddata, startMarker)
	endIndex := bytes.Index(mddata, endMarker)
	var metaData MdMetaData = MdMetaData{IsTemplate: false}

	if startIndex != -1 && endIndex != -1 && startIndex < endIndex {
		commentContent := mddata[startIndex+len(startMarker) : endIndex]

		err := yaml.Unmarshal(commentContent, &metaData)
		if err != nil {
			metaData = MdMetaData{IsTemplate: false}
		}
	}

	if metaData.IsTemplate {
		var data interface{} = struct{}{}
		var err error
		dataSource := metaData.TemplateDataPath
		// lets get data
		var rawDataForMd []byte
		if strings.HasPrefix(dataSource, "http://") || strings.HasPrefix(dataSource, "https://") {
			rawDataForMd, err = getHTTP(dataSource)
			if err != nil {
				rawDataForMd = make([]byte, 0)
			}
		} else {
			basePath := filepath.Dir(source)
			rawDataForMd, err = os.ReadFile(basePath + "/" + dataSource)
			if err != nil {
				rawDataForMd = make([]byte, 0)
			}
		}

		err = json.Unmarshal([]byte(rawDataForMd), &data)
		if err != nil {
			err = yaml.Unmarshal([]byte(rawDataForMd), &data)
		}
		if err != nil {
			data = struct{}{}
		}
		fmt.Println("dataformd:", data)
		tmpl, err := template.New("mdtmpl").Parse(string(mddata))
		if err != nil {
			return nil, err
		}
		var result bytes.Buffer
		err = tmpl.Execute(&result, data)
		if err != nil {
			fmt.Println("error md templates:", err)
			return nil, err
		}
		mddata = result.Bytes()
		fmt.Println(string(mddata))
	}

	mddata = markdown.NormalizeNewlines(mddata)

	renderer := newCustomizedRender()
	maybeUnsafeHTML := markdown.ToHTML(mddata, nil, renderer)
	html := bluemonday.UGCPolicy().SanitizeBytes(maybeUnsafeHTML)

	return html, nil

}

func getHTTP(url string) ([]byte, error) {

	// Make the HTTP GET request
	response, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer response.Body.Close()

	// Check the response status code
	if response.StatusCode != http.StatusOK {
		err = fmt.Errorf("unexpected status code: %v", response.StatusCode)
		return nil, err
	}

	// Read the response body into a byte slice
	content, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	return content, nil
}
