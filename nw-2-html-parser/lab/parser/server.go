package main

import (
	"github.com/mgutz/logxi/v1"
	"html/template"
	"net/http"
)


const INDEX_HTML = `
    <!doctype html>
    <html lang="ru">
        <head>
            <meta charset="utf-8">
            <title>Рамблер новости</title>
			<style>
				body {
					text-align: center;
				}
				h {
					font-size: 24px;
				}
				a {
					text-decoration: none;
					font-family: "Comic Sans MS", "Comic Sans", cursive;
					font-size: 18px;
				}
  			</style>
        </head>
        <body>
			<h>Рамблер новости</h>
			<br/>
            {{if .}}
                {{range .}}
                    <a href="{{.Ref}}">{{.Title}}</a>
                    <br/>
                {{end}}
            {{else}}
                Не удалось загрузить новости!
            {{end}}
        </body>
    </html>
    `

/*
Var 16
Формирование списка статей с
news.rambler.ru/articles
 */


var indexHtml = template.Must(template.New("index").Parse(INDEX_HTML))

func main() {
	http.HandleFunc("/", func (rsp http.ResponseWriter, rq *http.Request) {
		method, path := rq.Method, rq.URL.Path
		log.Info("got rq", "Method", method, "Path", path)

		if "/" == path && "GET" == method {
			if err := indexHtml.Execute(rsp, downloadNews()); err != nil {
				log.Error("HTML creation failed", "error", err)

			} else {
				log.Info("rsp sent to client successfully")
			}

		} else {
			log.Error("invalid path or method", "Path", path)
			rsp.WriteHeader(http.StatusBadRequest)
		}
	})
	log.Info("starting listener")
	log.Error("listener failed", "error", http.ListenAndServe("127.0.0.1:6002", nil))
}
