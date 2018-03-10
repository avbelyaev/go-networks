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
	elemClass := getAttr(node, "class")
	splitted := strings.Split(elemClass, " ")
	return node != nil && node.Type == html.ElementNode && contains(splitted, className)
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

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
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

func cleanTitle(title string) string {
	rightIndex := strings.Index(title, "<")
	if -1 != rightIndex {
		return title[:rightIndex]
	}
	return title
}

func search(node *html.Node) []*Item {
	var articleNodes = getElementsByClassName(node, "format-card__text")
	var items []*Item
	for _, articleNode := range articleNodes {

		var article Item
		article.Ref = "https://news.rambler.ru" + getAttr(articleNode, "href")

		titles := getElementsByClassName(articleNode, "format-card__title")
		nodeAsString := nodeToString(titles[0])
		innerHTML := getInnerHTML(nodeAsString)
		article.Title = cleanTitle(innerHTML)

		items = append(items, &article)
	}
	return items
}

const TARGET_URL = "https://news.rambler.ru/articles/"


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
