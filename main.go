package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	blackfriday "gopkg.in/russross/blackfriday.v2"
)

type config struct {
	Title     string
	Subtitle  string
	Blacklist []string
}

type site struct {
	Cfg   config
	Style string
}

type page struct {
	Site      site
	IndexPath string
	Nav       []item
	Body      string
}

type item struct {
	Name string
	Path string
	This bool
}

var tmpl = template.Must(template.New("page").Parse(pageTmpl))

func main() {
	cfgfile := flag.String("config", "config.json", "config file")
	stlfile := flag.String("style", "style.css", "CSS file")
	flag.Usage = usage
	flag.Parse()

	if flag.NArg() != 2 {
		usage()
		os.Exit(2)
	}
	indir := filepath.Clean(flag.Arg(0))
	outdir := filepath.Clean(flag.Arg(1))

	site, err := readcfg(*cfgfile, *stlfile)
	if err != nil {
		fatalf("cannot get configuration: %v\n", err)
	}

	err = filepath.Walk(indir, buildpage(site, indir, outdir))
	if err != nil {
		fatalf("cannot walk tree: %v\n", err)
	}
}

func readcfg(cfgfile, stlfile string) (site, error) {
	stldata, err := ioutil.ReadFile(stlfile)
	if err != nil {
		return site{}, fmt.Errorf("cannot read style file: %v", err)
	}
	cfgdata, err := ioutil.ReadFile(cfgfile)
	if err != nil {
		return site{}, fmt.Errorf("cannot read config file: %v", err)
	}

	var cfg config
	if err := json.Unmarshal(cfgdata, &cfg); err != nil {
		return site{}, fmt.Errorf("cannot parse config: %v", err)
	}

	s := site{
		Cfg:   cfg,
		Style: string(stldata),
	}

	return s, nil
}

func buildpage(site site, indir, outdir string) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("cannot walk tree: %v", err)
		}

		sitepath := strings.TrimPrefix(path, indir)

		if info.IsDir() {
			newdir := filepath.Join(outdir, sitepath)

			if err := os.Mkdir(newdir, 0755); err != nil {
				return fmt.Errorf("cannot create directory: %v", err)
			}
			return nil
		}

		if filepath.Ext(path) != ".md" {
			return nil
		}

		page, err := parsepage(site, indir, sitepath)
		if err != nil {
			return fmt.Errorf("cannot parse page: %v", err)
		}

		newpath := filepath.Join(outdir, strings.TrimSuffix(sitepath, filepath.Ext(sitepath))+".html")

		f, err := os.Create(newpath)
		if err != nil {
			return fmt.Errorf("cannot open output file: %v", err)
		}
		defer f.Close()

		if err := tmpl.Execute(f, page); err != nil {
			return fmt.Errorf("cannot execute template: %v", err)
		}

		return nil
	}
}

func parsepage(site site, indir, sitepath string) (page, error) {
	b, err := ioutil.ReadFile(filepath.Join(indir, sitepath))
	if err != nil {
		return page{}, fmt.Errorf("cannot read file: %v", err)
	}
	body := blackfriday.Run(b)

	sitedir := filepath.Dir(sitepath)

	indexPath := "index.html"
	if sitedir != "/" {
		indexPath = strings.Repeat("../", strings.Count(sitedir, "/")) + "index.html"
	}

	nav, err := buildNav(site, indir, sitepath)
	if err != nil {
		return page{}, fmt.Errorf("cannot build nav: %v", err)
	}

	page := page{
		Site:      site,
		IndexPath: indexPath,
		Nav:       nav,
		Body:      string(body),
	}

	return page, nil
}

func buildNav(site site, indir, sitepath string) ([]item, error) {
	var nav []item

	if filepath.Base(sitepath) != "index.md" {
		nav = append(nav, item{".", "index.html", false})
	}
	if filepath.Dir(sitepath) != "/" {
		nav = append(nav, item{"..", "../index.html", false})
	}

	d := filepath.Dir(filepath.Join(indir, sitepath))
	files, err := ioutil.ReadDir(d)
	if err != nil {
		return nil, fmt.Errorf("cannot read directory: %v", err)
	}

	for _, f := range files {
		filename := f.Name()
		name := strings.TrimSuffix(filename, filepath.Ext(filename))

		if isBlacklisted(name, site.Cfg.Blacklist) {
			continue
		}

		var path string
		if f.IsDir() {
			path = filepath.Join(name, "index.html")
		} else {
			path = name + ".html"
		}
		repr := strings.Replace(name, "_", " ", -1)
		this := filepath.Base(sitepath) == filename

		nav = append(nav, item{repr, path, this})
	}

	return nav, nil
}

func isBlacklisted(name string, blacklist []string) bool {
	if name == "index" {
		return true
	}

	for _, bl := range blacklist {
		if name == bl {
			return true
		}
	}

	return false
}

func fatalf(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, format, a...)
	os.Exit(1)
}

func usage() {
	fmt.Println("usage: gosw indir outdir")
	flag.PrintDefaults()
}
