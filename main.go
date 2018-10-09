package main

import (
	"log"
	"net/http"
	"net/url"
	"strings"
	"syscall/js"

	"github.com/dennwc/dom"
)

type writer dom.Element

// Write implements io.Writer.

func (d writer) Write(p []byte) (n int, err error) {
	node := dom.GetDocument().CreateElement("div")
	node.SetTextContent(string(p))
	(*dom.Element)(&d).AppendChild(node)
	return len(p), nil
}

func parse(s string) (string, string, string) {
	s = strings.ToLower(s)

	var u, a, e string

	f := strings.Fields(s)
	for i, word := range f {
		if word == "user" || word == "username" || word == "username:" || word == "user:" || word == "login:" || word == "login" || word == "name:" || word == "name" || word == "account" || word == "account:" {
			u = strings.TrimRight(f[i+1], ".,!:?")
		}

		if word == "reset" || word == "add" || word == "delete" || word == "create" || word == "disable" || word == "remove" {
			a = strings.TrimRight(f[i], ".,!:?")
		}

		if strings.Contains(word, "@") {
			e = strings.TrimRight(f[i], ".,!:?")
		}

	}

	// users := strings.SplitAfter(s, "user ")
	// actions := strings.SplitAfter(s, "please ")
	// emails := strings.SplitAfter(s, "email is")

	// u := users[0]
	// a := actions[0]
	// e := emails[0]

	return u, a, e
}

func main() {

	t := dom.GetDocument().GetElementById("tasks")
	logger := log.New((*writer)(t), "", log.LstdFlags)
	//i := dom.GetDocument().GetElementById("tasks")

	// u := js.Global().Get("document").Call("getElementById", "user").Get("value").String()
	// a := js.Global().Get("document").Call("getElementById", "action").Get("value").String()
	// e := js.Global().Get("document").Call("getElementById", "email").Get("value").String()
	c := js.Global().Get("document").Call("getElementById", "chat").Get("value").String()

	u, a, e := parse(c)

	if u == "" || a == "" || e == "" {
		logger.Print("task is invalid please specify all the values, user: " + u + ", action: " + a + ", email: " + e)
	} else {
		_, err := http.PostForm("./index.html",
			url.Values{"user": {u}, "action": {a}, "email": {e}})
		if err != nil {
			log.Fatal(err)
		}
		logger.Print("new task is ready, user: " + u + ", action: " + a + ", email: " + e)
	}

	//t.SetAttribute("target", "main.go")
	//t.SetAttribute("target", "tasks.html")

}
