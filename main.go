package main

import (
	"log"
	"net/http"
	"net/url"
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

func main() {

	t := dom.GetDocument().GetElementById("tasks")
	//i := dom.GetDocument().GetElementById("tasks")

	u := js.Global().Get("document").Call("getElementById", "user").Get("value").String()
	a := js.Global().Get("document").Call("getElementById", "action").Get("value").String()
	e := js.Global().Get("document").Call("getElementById", "email").Get("value").String()

	_, err := http.PostForm("./index.html",
		url.Values{"user": {u}, "action": {a}, "email": {e}})
	if err != nil {
		log.Fatal(err)
	}

	//t.SetAttribute("target", "main.go")
	//t.SetAttribute("target", "tasks.html")

	logger := log.New((*writer)(t), "", log.LstdFlags)
	logger.Print("new task is ready, user: " + u + ", action: " + a + ", email: " + e)

}
