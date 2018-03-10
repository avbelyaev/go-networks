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
            <title>Афиша</title>
        </head>
        <body>
			<h style="font-size:24px;" >Афиша</h>
			<br/>
            {{if .}}
                {{range .}}
                    <a style="color:red; font-size:18px;" href="{{.Ref}}">{{.Title}}</a>
                    <br/>
                {{end}}
            {{else}}
                Не удалось загрузить новости!
            {{end}}
        </body>
    </html>
    `

/*
Var 11
Формирование списка фильмов с
www.afisha.ru/msk/cinema
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
