package mclihttp

import (
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// var TemplatesCache = make(map[string]*template.Template, 10)

func processDirectory(dir, base, part string, cache *map[string]*template.Template) error {
	// fmt.Println(dir, base, part)
	pages, err := filepath.Glob(filepath.Join(dir, "*.page.html"))
	if err != nil {
		return err
	}

	for _, page := range pages {
		basename := filepath.Base(page)
		filename := strings.TrimSuffix(basename, filepath.Ext(basename))

		ts, err := template.ParseFiles(page)
		if err != nil {
			return err
		}

		// we use method ParseGlob to add all skeleton templates *.layout.html
		ts, err = ts.ParseGlob(filepath.Join(base, "*.layout.html"))
		if err != nil {
			return err
		}

		// we use method ParseGlob to add all addon templates
		// *.partial.html
		ts, err = ts.ParseGlob(filepath.Join(part, "*.partial.html"))
		if err != nil {
			return err
		}

		(*cache)[filepath.Join(dir, filename)] = ts
	}
	return nil
}

func LoadTemplatesCache(rootTmpl string) (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}

	var mainLayoutPath string = ""
	var mainPartialPath string = ""

	if e, _ := exists(filepath.Join(rootTmpl, "__base")); e {
		mainLayoutPath = filepath.Join(rootTmpl, "__base")
	}
	if e, _ := exists(filepath.Join(rootTmpl, "__partial")); e {
		mainPartialPath = filepath.Join(rootTmpl, "__partial")
	}

	filepath.Walk(rootTmpl, func(wPath string, info os.FileInfo, err error) error {
		// if the same path
		if wPath == rootTmpl {
			err := processDirectory(wPath, mainLayoutPath, mainPartialPath, &cache)
			if err != nil {
				return err
			}
		}
		// If current path is Dir - process it
		if info.IsDir() && !(strings.Contains(wPath, "__base") || strings.Contains(wPath, "__partial")) {

			layoutPath := mainLayoutPath
			partialPath := mainPartialPath

			if e, _ := exists(filepath.Join(wPath, "__base")); e {
				layoutPath = filepath.Join(wPath, "__base")
			}
			if e, _ := exists(filepath.Join(wPath, "__partial")); e {
				partialPath = filepath.Join(wPath, "__partial")
			}

			err := processDirectory(wPath, layoutPath, partialPath, &cache)
			if err != nil {
				return err
			}
		}
		// if we got file, we do nothing
		// if wPath != rootTmpl && !info.IsDir() {
		// _ = fmt.Sprintf("[%s]\n", wPath)
		// }
		return nil
	})

	return cache, nil

}

// Method to set up templates routes processing
func (r *Router) SetTmplRoutes(tmplPath, tmplPrefix string) {
	// adding templates to routes
	tmplPrefix = strings.Replace(strings.TrimSpace(tmplPrefix), "/", "/", -1)
	tmplPrefix = strings.Replace(strings.TrimSpace(tmplPrefix), "\\", "\\", -1)
	tmplPath = strings.TrimSpace(tmplPath)
	tmplPath = strings.TrimRight(tmplPath, "/")

	if e, _ := exists(tmplPath); e {
		// cache, err := LoadTemplatesCache("http-data/templates")
		cache, err := LoadTemplatesCache(tmplPath)
		if err != nil {
			r.errorLog.Fatal().Msgf("template caching error: %v", err)
		}
		r.infoLog.Trace().Msg("Templates path:" + tmplPath + " Templates prefix:" + tmplPrefix)
		tmplPrefix = "/" + tmplPrefix + "/"

		// constucting new route
		tmplRoute := NewRoute(tmplPrefix, Prefix)
		// setting handler
		tmplRoute.SetHandler(func(res http.ResponseWriter, req *http.Request) {
			url := req.URL.Path
			tmplName := strings.TrimLeft(url, tmplPrefix)
			tmplKey := tmplPath + "/" + tmplName
			tmpl, ok := cache[tmplKey]
			if !ok {
				http.Error(res, "404 Template Not Found: "+tmplKey, 404)
				return
			}
			// Probably Headers processing
			// for header, vals := range req.Header {
			// 	fmt.Fprintf(writer, "Header: %v: %v\n", header, vals)
			// }

			var queryData string = ""
			if req.Method == "GET" {
				queryData = req.URL.Query().Get("data")
			}
			if req.Method == "POST" {
				defer req.Body.Close()
				byteData, err := io.ReadAll(req.Body)
				if err != nil {
					http.Error(res, "Internal Server Error: cannot read body of Post request", 500)
				}
				queryData = string(byteData)
			}

			err = tmpl.Execute(res, struct {
				Req  *http.Request
				Data interface{}
			}{Req: req, Data: struct{ Dummy string }{"Hello world !!!"}})
			if err != nil {
				http.Error(res, err.Error(), http.StatusInternalServerError)
				return
			}
			r.infoLog.Trace().Msg(url + " : " + queryData)
		})
		r.AddRoute(tmplRoute)
	}
}
