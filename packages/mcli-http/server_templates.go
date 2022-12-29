package mclihttp

import (
	"fmt"
	"html/template"
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
	fmt.Println(dir, base, part)
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
		if wPath != rootTmpl && !info.IsDir() {
			_ = fmt.Sprintf("[%s]\n", wPath)
		}
		return nil
	})

	return cache, nil

}
