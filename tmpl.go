// See LICENSE file for copyright and license details.

package main

const pageTmpl = `<!doctype html>
<html>
	<head>
		<meta charset="utf-8">
		<title>{{.Site.Cfg.Title}}</title>
		<style type="text/css">
			{{.Site.Style}}
		</style>
	</head>
	<body>
		<div id="header">
			<a id="headerLink" href="{{.IndexPath}}">{{.Site.Cfg.Title}}</a>
			<span id="headerSubtitle">{{.Site.Cfg.Subtitle}}</span>
		</div>
		<div id="content">
			<div id="nav">
				<ul>
					{{- range .Nav}}
					{{- if .This}}
					<li><a class="thisPage" href="{{.Path}}">{{.Name}}</a></li>
					{{- else}}
					<li><a href="{{.Path}}">{{.Name}}</a></li>
					{{- end}}
					{{- end}}
				</ul>
			</div>
			<div id="main">
				{{.Body}}
			</div>
		</div>
		<div id="footer">
			<span class="right"><a href="https://github.com/jroimartin/gosw">Powered by gosw</a></span>
		</div>
	</body>
</html>`
