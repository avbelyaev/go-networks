package main

import (
	"github.com/mgutz/logxi/v1"
	"golang.org/x/net/html"
	"net/http"
	"strings"
	"bytes"
	"io"
)

func getAttr(node *html.Node, key string) string {
	for _, attr := range node.Attr {
		if attr.Key == key {
			return attr.Val
		}
	}
	return ""
}

func getChildren(node *html.Node) []*html.Node {
	var children []*html.Node
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		children = append(children, c)
	}
	return children
}

func isElem(node *html.Node, tag string) bool {
	return node != nil && node.Type == html.ElementNode && node.Data == tag
}

func isElemOfClass(node *html.Node, className string) bool {
	return node != nil && node.Type == html.ElementNode && strings.Contains(getAttr(node, "class"), className)
}

func isText(node *html.Node) bool {
	return node != nil && node.Type == html.TextNode
}

func isDiv(node *html.Node, class string) bool {
	return isElem(node, "div") && strings.Contains(getAttr(node, "class"), class)
}

func isAnchor(node *html.Node, class string) bool {
	return isElem(node, "a") && strings.Contains(getAttr(node, "class"), class)
}

type Item struct {
	Ref, Time, Title string
}


func getElementsByClassName(node *html.Node, className string) []*html.Node {
	var nodes []*html.Node
	if isElemOfClass(node, className) {
		nodes = append(nodes, node)
	}
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		var innerNodes = getElementsByClassName(c, className)
		for _, innerNode := range innerNodes {
			nodes = append(nodes, innerNode)
		}
	}
	return nodes
}

func nodeToString(node *html.Node) string {
	var buf bytes.Buffer
	w := io.Writer(&buf)
	html.Render(w, node)
	return buf.String()
}

func getInnerHTML(stringToSearchWithin string) string {
	leftIndex := strings.Index(stringToSearchWithin, ">")
	rightIndex := strings.LastIndex(stringToSearchWithin, "<")
	return stringToSearchWithin[leftIndex + 1 : rightIndex]
}

func search(node *html.Node) []*Item {
	var movieNodes = getElementsByClassName(node, "fn permalink")
	var movies []*Item
	for _, movieNode := range movieNodes {

		var movie Item
		movie.Ref = getAttr(movieNode, "href")

		nodeAsString := nodeToString(movieNode)
		movie.Title = getInnerHTML(nodeAsString)

		movies = append(movies, &movie)
	}
	return movies
}

const TARGET_URL = "https://www.afisha.ru/msk/cinema/"


func downloadNews() []*Item {
	log.Info("sending request ", "url", TARGET_URL)
	if response, err := http.Get(TARGET_URL); err != nil {
		log.Error("request failed", "error", err)
	} else {
		defer response.Body.Close()
		status := response.StatusCode
		log.Info("got response", "status", status)
		if status == http.StatusOK {
			if doc, err := html.Parse(response.Body); err != nil {
				log.Error("invalid HTML", "error", err)
			} else {
				log.Info("HTML from parsed successfully")
				items := search(doc)
				print(items)
				return items
			}
		}
	}
	return nil
}
