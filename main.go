package main

import (
	"bytes"
	"encoding/json"
	"fmt"
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

func parse(s string) (string, string, string, string, string, string, string) {
	s = strings.ToLower(s)

	var u, a, e string
	var r, t, z, v string

	f := strings.Fields(s)
	for i, word := range f {
		if word == "user" || word == "username" || word == "username:" || word == "user:" || word == "login:" || word == "login" || word == "name:" || word == "name" || word == "account" || word == "account:" {
			if f[i+1] != "for" {
				u = strings.TrimRight(f[i+1], ".,!:?")
			}
		}

		if word == "reset" || word == "add" || word == "delete" || word == "create" || word == "disable" || word == "remove" {
			a = strings.TrimRight(f[i], ".,!:?")
		}

		if strings.Contains(word, "@") {
			if strings.Contains(word, "|") {
				e = (strings.SplitAfter(f[i], "|"))[0]
				e = strings.Trim(e, "[].,!:?]|")
			} else {
				e = strings.Trim(f[i], "[].,!:?]")
			}

			//			e = strings.TrimRight(f[i], ".,!:?")
		}

		if word == "a" || word == "ns" || word == "cname" {
			if f[i-1] == "type" || f[i+1] == "record" {
				t = f[i]
			}
		}
		if word == "hostname" || word == "record" {
			r = strings.TrimRight(f[i+1], ".,!:?")
		}
		if word == "zone" || word == "domain" {
			z = strings.TrimRight(f[i+1], ".,!:?")
		}
		if word == "ip" || word == "address" {
			if f[i+1] != "address" {
				v = f[i+1]
			}
		}
	}

	// users := strings.SplitAfter(s, "user ")
	// actions := strings.SplitAfter(s, "please ")
	// emails := strings.SplitAfter(s, "email is")

	// u := users[0]
	// a := actions[0]
	// e := emails[0]

	fmt.Println(u, a, e, t, r, z, v)
	return u, a, e, t, r, z, v
}

func main() {

	t := dom.GetDocument().GetElementById("tasks")
	logger := log.New((*writer)(t), "", log.LstdFlags)
	//i := dom.GetDocument().GetElementById("tasks")

	// u := js.Global().Get("document").Call("getElementById", "user").Get("value").String()
	// a := js.Global().Get("document").Call("getElementById", "action").Get("value").String()
	// e := js.Global().Get("document").Call("getElementById", "email").Get("value").String()
	c := js.Global().Get("document").Call("getElementById", "chat").Get("value").String()

	u, a, e, ty, r, z, v := parse(c)

	if u == "" || a == "" || e == "" {
		//logger.Print("task is invalid please specify all the values, user: " + u + ", action: " + a + ", email: " + e)

		values := map[string]string{
			"source":     "governor",
			"sourceid":   "-",
			"record":     r,
			"recordtype": strings.ToUpper(ty),
			"zone":       z,
			"target":     v,
			"action":     a,
			"email":      e,
		}

		json, _ := json.Marshal(values)
		req, err := http.NewRequest("POST", "/api/dns/", bytes.NewBuffer(json))
		req.Header.Set("X-Custom-Header", "myvalue")
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()

		// _, err := http.PostForm("./index.html",
		// 	url.Values{"user": {u}, "action": {a}, "email": {e}, "source": {"governor"}, "sourceid": {"-"}})
		// if err != nil {
		// 	log.Fatal(err)
		// }
		logger.Print("new task is ready, user: " + u + ", action: " + a + ", email: " + e)
	} else {
		_, err := http.PostForm("./index.html",
			url.Values{"user": {u}, "action": {a}, "email": {e}, "source": {"governor"}, "sourceid": {"-"}})
		if err != nil {
			log.Fatal(err)
		}
		logger.Print("new task is ready, user: " + u + ", action: " + a + ", email: " + e)
	}

	//t.SetAttribute("target", "main.go")
	//t.SetAttribute("target", "tasks.html")

}
