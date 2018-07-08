// See LICENSE file for copyright and license details.

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
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
	Text string
	Link string
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

	indir, err := filepath.Abs(flag.Arg(0))
	if err != nil {
		fatalf("input dir error: %v\n", indir)
	}
	outdir, err := filepath.Abs(flag.Arg(1))
	if err != nil {
		fatalf("output dir error: %v\n", outdir)
	}

	site, err := readcfg(*cfgfile, *stlfile)
	if err != nil {
		fatalf("configuration error: %v\n", err)
	}

	if err := filepath.Walk(indir, buildpage(site, indir, outdir)); err != nil {
		fatalf("walk error: %v\n", err)
	}
}

func readcfg(cfgfile, stlfile string) (site, error) {
	stldata, err := ioutil.ReadFile(stlfile)
	if err != nil {
		return site{}, err
	}
	cfgdata, err := ioutil.ReadFile(cfgfile)
	if err != nil {
		return site{}, err
	}

	var cfg config
	if err := json.Unmarshal(cfgdata, &cfg); err != nil {
		return site{}, err
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
			return err
		}

		sitepath := strings.TrimPrefix(path, indir)

		// Create directory
		if info.IsDir() {
			dst := filepath.Join(outdir, sitepath)
			return os.MkdirAll(dst, 0755)
		}

		// Copy non-md file
		if filepath.Ext(path) != ".md" {
			dst := filepath.Join(outdir, sitepath)
			return copyFile(dst, path)
		}

		// Render md file into html
		page, err := parsepage(site, indir, sitepath)
		if err != nil {
			return err
		}

		htmlpath := strings.TrimSuffix(sitepath, filepath.Ext(sitepath)) + ".html"
		dst := filepath.Join(outdir, htmlpath)

		f, err := os.Create(dst)
		if err != nil {
			return err
		}
		defer f.Close()

		return tmpl.Execute(f, page)
	}
}

func copyFile(dst, src string) error {
	fsrc, err := os.Open(src)
	if err != err {
		return err
	}
	defer fsrc.Close()

	fdst, err := os.Create(dst)
	if err != err {
		return err
	}
	defer fdst.Close()

	_, err = io.Copy(fdst, fsrc)
	return err
}

func parsepage(site site, indir, sitepath string) (page, error) {
	b, err := ioutil.ReadFile(filepath.Join(indir, sitepath))
	if err != nil {
		return page{}, err
	}
	body := blackfriday.Run(b)

	sitedir := filepath.Dir(sitepath)

	indexPath := "index.html"
	if sitedir != string(filepath.Separator) {
		n := strings.Count(sitedir, string(filepath.Separator))
		indexPath = path.Join(strings.Repeat("../", n), "index.html")
	}

	nav, err := buildNav(site, indir, sitepath)
	if err != nil {
		return page{}, err
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
	if filepath.Dir(sitepath) != string(filepath.Separator) {
		nav = append(nav, item{"..", "../index.html", false})
	}

	d := filepath.Dir(filepath.Join(indir, sitepath))
	files, err := ioutil.ReadDir(d)
	if err != nil {
		return nil, err
	}

	for _, f := range files {
		filename := f.Name()
		name := strings.TrimSuffix(filename, filepath.Ext(filename))

		if isBlacklisted(name, site.Cfg.Blacklist) {
			continue
		}

		var link string
		if f.IsDir() {
			link = path.Join(name, "index.html")
		} else {
			link = name + ".html"
		}
		text := strings.Replace(name, "_", " ", -1)
		this := filepath.Base(sitepath) == filename

		nav = append(nav, item{text, link, this})
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
	fmt.Fprintln(os.Stderr, "usage: gosw indir outdir")
	flag.PrintDefaults()
}
