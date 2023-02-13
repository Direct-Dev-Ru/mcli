package mclihttp

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	mcli_utils "mcli/packages/mcli-utils"
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
func (r *Router) SetTmplRoutes(tmplPath, tmplPrefix, tmplDataPath string) {
	// adding templates to routes
	tmplPrefix = strings.Replace(strings.TrimSpace(tmplPrefix), "/", "/", -1)
	tmplPrefix = strings.Replace(strings.TrimSpace(tmplPrefix), "\\", "\\", -1)
	tmplPath = strings.TrimSpace(tmplPath)
	tmplPath = strings.TrimRight(tmplPath, "/")
	tmplDataPath = strings.TrimSpace(tmplDataPath)
	tmplDataPath = strings.TrimRight(tmplDataPath, "/")

	_ = tmplDataPath
	// by default all templates are free and not require authentication
	// but if it contains "protected" subfolder in its path, then it should provided with json or bson data file
	// and there must be auth tokens presented:
	// {
	// 	"Version": "1.0.0",
	// 	"Timestamp": 1257894000000000000 //time.Now().UnixNano()
	// 	"Templates": ["home.page","about.page","some.page"]
	//	"AuthTokens": ["token1", "token2"]
	// 	"TemplateData": {
	// 	  "Title": "Direct-Dev.Ru Blog",
	// 	  "Header": "This blog made for you interest. Look inside",
	// 	  "MainText": "Here you can find a lot of interesting information about development processes"
	// 	}
	// }

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
		tmplRoute.SetHandler(func(res http.ResponseWriter, req *http.Request) {
			url := req.URL.Path
			tmplName := strings.TrimPrefix(url, tmplPrefix)
			tmplKey := tmplPath + "/" + tmplName
			tmpl, ok := cache[tmplKey]
			if !ok {
				http.Error(res, "404 Template Not Found: "+tmplKey, 404)
				return
			}
			var queryStrData, pathToData string = "", ""
			var queryData interface{}
			var err error
			var xAccessTokens []string = make([]string, 0, 3)
			header, ok := req.Header["X-Access-Tokens"]
			if ok {
				xAccessTokens = append(xAccessTokens, header...)
			}
			_ = xAccessTokens

			if req.Method == "GET" {
				queryStrData = req.URL.Query().Get("data")
			}
			if req.Method == "POST" {
				defer req.Body.Close()
				byteData, err := io.ReadAll(req.Body)
				if err != nil {
					http.Error(res, "Internal Server Error: cannot read body of Post request", 500)
				}
				queryStrData = string(byteData)
			}
			// we expect queryData is jsonString with specified fields:
			// optional PathToJson:string - relative path to find json file contains required data
			// optional header "X-Access-Tokens" (X-Access-Tokens:string,string,string) - for authenticate purposes //TODO:
			// if PathToJson is empty or not specified it finds file as relative template path
			// and name as template name without extension,
			// for example  tmpl-path=http-data/templates , tmpl-datapath = http-data/templates-data
			// and our template is http-data/templates/home/home.page.html
			// if where are no JsonPath with request we firstly looking for a path http-data/templates-data/bson/home/home.page.bson
			// if it exists we convert it to json and pass to template renderer
			// if not exists we are looking for http-data/templates-data/home/home.page.bson
			// if not exists we are looking for http-data/templates-data/home/home.page.json
			// if not exists - we pass impty interface{} to template renderer
			queryData, err = mcli_utils.JsonStringToInterface(queryStrData)
			if err == nil {
				typedQueryData, ok := queryData.(map[string]interface{})
				if ok {
					if v, ok := typedQueryData["PathToJson"]; ok {
						pathToData = v.(string)
						// only relative path from RootPath (where process starts) and only down in folder tree
						pathToData = strings.TrimPrefix(strings.TrimSpace(pathToData), "/")
						pathToData = strings.TrimPrefix(pathToData, "./")
						pathToData = strings.TrimPrefix(pathToData, "../")
					}
				} else {
					err = fmt.Errorf("wrong json object from query")
					pathToData = ""
				}
			} else {
				// dir := path.Dir(tmplName)
				// base := path.Base(tmplName)
				// ext := path.Ext(base)
				// name := base[:len(base)-len(ext)]
				// noExt := path.Join(dir, name)
				// tmplName - e.g. "home/home.page"
				// lets try first candidate tmplDataPath + "/bson/"+noExt+".bson"
				candidatePath := tmplDataPath + "/bson/" + tmplName + ".bson"
				isExist, _ := exists(candidatePath)
				if isExist {
					pathToData = candidatePath
				}
				if !isExist {
					candidatePath = tmplDataPath + "/" + tmplName + ".bson"
					isExist, _ = exists(candidatePath)
					if isExist {
						pathToData = candidatePath
					}
				}
				if !isExist {
					// candidatePath = os.Getenv("RootPath") + "/" + tmplDataPath + "/" + tmplName + ".json"
					candidatePath = tmplDataPath + "/" + tmplName + ".json"
					isExist, _ = exists(candidatePath)
					if isExist {
						pathToData = candidatePath
					}
				}
			}
			var bytesDataForTemplate []byte
			var templateData interface{}

			if len(pathToData) > 0 {
				bytesDataForTemplate, err = os.ReadFile(pathToData)
				if err != nil {
					templateData, err = mcli_utils.JsonStringToInterface("{}")
				} else {
					if strings.HasSuffix(pathToData, ".bson") {
						templateData, err = mcli_utils.BsonDataToInterfaceMap(bytesDataForTemplate)
					} else {
						templateData, err = mcli_utils.JsonStringToInterface(string(bytesDataForTemplate))
					}
				}
			}
			if err != nil {
				http.Error(res, err.Error(), http.StatusInternalServerError)
				return
			}

			typedTemplateData, ok := templateData.(map[string]interface{})
			if !ok {
				http.Error(res, "error convert template data from interface", http.StatusInternalServerError)
				return
			}
			if v, ok := typedTemplateData["TemplateData"]; ok {
				templateData = v
			} else {
				http.Error(res, "error find TemplateData member in data map", http.StatusInternalServerError)
				return
			}

			// fmt.Println("request processing " + req.URL.String())

			var bindData interface{} = struct {
				Req  *http.Request
				Data interface{}
			}{
				Req:  req,
				Data: templateData,
			}

			err = tmpl.Execute(res, bindData)

			if err != nil {
				http.Error(res, err.Error(), http.StatusInternalServerError)
				return
			}
			r.infoLog.Trace().Msg(url + " : " + queryStrData)
		})
		r.AddRoute(tmplRoute)
	}
}
