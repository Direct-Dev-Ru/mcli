package mclihttp

import (
	"context"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	mcli_utils "mcli/packages/mcli-utils"
)

type TemplateEntry struct {
	TmplName            string `yaml:"tmpl-name"`
	TmplType            string `yaml:"tmpl-type"`
	TmplPath            string `yaml:"tmpl-path"`
	TmplPrefix          string `yaml:"tmpl-prefix"`
	TmplDataPath        string `yaml:"tmpl-datapath"`
	TmplRefreshType     string `yaml:"tmpl-refresh-type"`
	TmplRefreshInterval string `yaml:"tmpl-refresh-interval"`
}

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
type MyTemplate struct {
	template     *template.Template
	templatePath string
	timestamp    time.Time
}
type MyTemplateCache struct {
	sync.RWMutex
	cache    map[string]*MyTemplate
	tmplName string
	tmplPath string
}

func processTemplDir(dir, base, part string, cache *map[string]*MyTemplate) error {
	// fmt.Println(dir, base, part)

	pages, err := filepath.Glob(filepath.Join(dir, "*.page.html"))
	if err != nil {
		return err
	}

	for _, page := range pages {
		basename := filepath.Base(page)
		filename := strings.TrimSuffix(basename, filepath.Ext(basename))

		pageFileStats, err := os.Stat(page)
		if err != nil {
			return err
		}

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

		(*cache)[filepath.Join(dir, filename)] = &MyTemplate{template: ts, templatePath: page, timestamp: pageFileStats.ModTime()}
	}
	return nil
}

func LoadMyTemplatesCache(rootTmpl string) (map[string]*MyTemplate, error) {
	cache := make(map[string]*MyTemplate)

	var mainLayoutPath string = ""
	var mainPartialPath string = ""

	if e, _ := exists(filepath.Join(rootTmpl, "__base")); e {
		mainLayoutPath = filepath.Join(rootTmpl, "__base")
	}
	if e, _ := exists(filepath.Join(rootTmpl, "__partial")); e {
		mainPartialPath = filepath.Join(rootTmpl, "__partial")
	}

	e := filepath.Walk(rootTmpl, func(wPath string, info os.FileInfo, err error) error {
		// if the same path
		if wPath == rootTmpl {
			err := processTemplDir(wPath, mainLayoutPath, mainPartialPath, &cache)
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

			err := processTemplDir(wPath, layoutPath, partialPath, &cache)
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

	return cache, e

}

// Method to set multiply templates routes

func (r *Router) SetTemplatesRoutes(ctx context.Context, templates []TemplateEntry) {
	for _, t := range templates {
		err := r.setTmplRoutes(ctx, t)
		if err != nil {
			r.infoLog.Error().Msgf("error load templates: %v", err)
		} else {
			r.infoLog.Trace().Msgf("setting up template cache: %v is successfull", t.TmplName)
		}
	}
}

func (r *Router) refreshTemplateCache(ctx context.Context, d time.Duration, myTmplCache *MyTemplateCache) {
	ticker := time.NewTicker(d)
	defer func() {
		r.infoLog.Trace().Msg("Ticker stopped.")
		ticker.Stop()
	}()

	for {
		select {
		case <-ctx.Done():
			r.infoLog.Trace().Msg("Refresh Task stopped.")
			ticker.Stop()
			return
		case <-ticker.C:
			// r.infoLog.Trace().Msg("Refresh Task running ... ")
			myTmplCache.Lock()
			cache, err := LoadMyTemplatesCache(myTmplCache.tmplPath)
			myTmplCache.Unlock()
			if err != nil {
				r.errorLog.Error().Msgf("refreshing of template caching %v got error: %v", myTmplCache.tmplName, err)
				ticker.Stop()
				return
			} else {
				myTmplCache.cache = cache
				r.infoLog.Trace().Msgf("refreshing of template cache: %v is successfull", myTmplCache.tmplName)
			}
			// default:
			// fmt.Println("Task undefined ...")
			// time.Sleep(5 * time.Second)
		}
	}
}

// Method to set up templates routes processing
func (r *Router) setTmplRoutes(ctx context.Context, t TemplateEntry) error {
	// adding templates to routes
	tmplPrefix := strings.Replace(strings.TrimSpace(t.TmplPrefix), "/", "/", -1)
	tmplPrefix = strings.Replace(strings.TrimSpace(tmplPrefix), "\\", "\\", -1)
	tmplPath := strings.TrimSpace(t.TmplPath)
	tmplPath = strings.TrimRight(tmplPath, "/")
	tmplDataPath := strings.TrimSpace(t.TmplDataPath)
	tmplDataPath = strings.TrimRight(tmplDataPath, "/")

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

		// ctx, cancel := context.WithCancel(context.Background())
		// defer cancel()

		cache, err := LoadMyTemplatesCache(tmplPath)
		if err != nil {
			r.errorLog.Error().Msgf("template caching error: %v", err)
			return err
		}
		myTemplateCache := &MyTemplateCache{cache: cache, tmplName: t.TmplName, tmplPath: tmplPath}
		if t.TmplRefreshType == "on-interval" {
			interval, err := strconv.Atoi(t.TmplRefreshInterval)
			if err != nil || interval == 0 {
				r.errorLog.Info().Msgf("Error with refresh interval in config: %v", err)
				r.infoLog.Info().Msg("Refresh interval sets to 30 sec")
				interval = 30
			}
			duration := time.Duration(int64(interval) * int64(time.Second))
			go r.refreshTemplateCache(ctx, duration, myTemplateCache)
		}

		r.infoLog.Trace().Msg("Templates path:" + tmplPath + " Templates prefix:" + tmplPrefix)
		tmplPrefix = "/" + tmplPrefix + "/"

		// constucting new route
		tmplRoute := NewRoute(tmplPrefix, Prefix)
		tmplRoute.SetHandler(func(res http.ResponseWriter, req *http.Request) {
			url := req.URL.Path
			tmplName := strings.TrimPrefix(url, tmplPrefix)
			tmplKey := tmplPath + "/" + tmplName
			tmpl, ok := myTemplateCache.cache[tmplKey]
			if !ok && !(t.TmplRefreshType == "on-change") {
				http.Error(res, "404 Template Not Found: "+tmplKey, 404)
				return
			}

			if t.TmplRefreshType == "on-change" && ok {
				tmplFileStats, err := os.Stat(tmpl.templatePath)
				if err != nil {
					http.Error(res, err.Error(), http.StatusInternalServerError)
					return
				}

				if modtime := tmplFileStats.ModTime(); modtime.UnixNano() > tmpl.timestamp.UnixNano() {
					myTemplateCache.cache, err = LoadMyTemplatesCache(tmplPath)
					if err != nil {
						http.Error(res, fmt.Sprintf("template caching error: %v", err), http.StatusInternalServerError)
					}
				}
				tmpl, ok = myTemplateCache.cache[tmplKey]
				if !ok {
					http.Error(res, "404 Template Not Found: "+tmplKey, 404)
					return
				}
			}
			if t.TmplRefreshType == "on-change" && !ok {
				myTemplateCache.cache, err = LoadMyTemplatesCache(tmplPath)
				if err != nil {
					http.Error(res, fmt.Sprintf("template caching error: %v", err), http.StatusInternalServerError)
				}
				tmpl, ok = myTemplateCache.cache[tmplKey]
				if !ok {
					http.Error(res, "404 Template Not Found: "+tmplKey, 404)
					return
				}
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
					err = fmt.Errorf("incorrect json object from query")
					pathToData = ""
					if err != nil {
						http.Error(res, "error converting json object, received from query: "+err.Error(), http.StatusInternalServerError)
						return
					}
				}
			} else {
				err = nil
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
			var bindData interface{}

			bindData = struct {
				Req  *http.Request
				Data interface{}
			}{
				Req:  req,
				Data: struct{}{},
			}
			if t.TmplType == "markdowm" {
				bindData = struct {
					Req      *http.Request
					Data     interface{}
					Contents map[string]template.HTML
				}{
					Req:      req,
					Data:     struct{}{},
					Contents: make(map[string]template.HTML),
				}
			}

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
				if err != nil {
					http.Error(res, "error converting json to interface: "+err.Error(), http.StatusInternalServerError)
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

				bindData = struct {
					Req  *http.Request
					Data interface{}
				}{
					Req:  req,
					Data: templateData,
				}
				if t.TmplType == "markdowm" {
					bindData = struct {
						Req      *http.Request
						Data     interface{}
						Contents map[string]template.HTML
					}{
						Req:      req,
						Data:     templateData,
						Contents: make(map[string]template.HTML),
					}
					if mdValue, ok := typedTemplateData["MarkdownContents"]; ok {
						markdownSources, ok := mdValue.(map[string]interface{})
						if !ok {
							http.Error(res, "error convert md sources to map[string]interface", http.StatusInternalServerError)
							return
						}
						mdMap := make(map[string]template.HTML)
						for key, source := range markdownSources {
							pathToMd, ok := source.(string)
							if !ok {
								http.Error(res, "error convert md source to string", http.StatusInternalServerError)
								return
							}
							htmlContent, err := ConvertMdToHtml(t.TmplDataPath + "/" + pathToMd)
							if err != nil {
								http.Error(res, "error convert md to html: "+err.Error(), http.StatusInternalServerError)
								return
							}
							mdMap[key] = template.HTML(string(htmlContent))
						}
						bindData = struct {
							Req      *http.Request
							Data     interface{}
							Contents map[string]template.HTML
						}{
							Req:      req,
							Data:     templateData,
							Contents: mdMap,
						}

					} else {
						http.Error(res, "error find MarkdownContents member in template map", http.StatusInternalServerError)
						return
					}

				}
			}
			// fmt.Println("request processing " + req.URL.String())

			err = tmpl.template.Execute(res, bindData)

			if err != nil {
				r.infoLog.Trace().Msgf("url : %v", err)
				http.Error(res, err.Error(), http.StatusInternalServerError)
				return
			}
			r.infoLog.Trace().Msg(url + " : " + queryStrData)
		})
		return r.AddRoute(tmplRoute)
	}
	return fmt.Errorf("path not exists %s", tmplPath)
}

// -----------------------------------------------------------------------------
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
