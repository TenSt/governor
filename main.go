package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"syscall/js"

	"github.com/dennwc/dom"
)

type writer dom.Element

// Write implements io.Writer.

type predictions struct {
	Predictions [][]float64 `json:"predictions"`
}

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

	c := js.Global().Get("document").Call("getElementById", "chat").Get("value").String()

	a := predict(c)
	_, err := http.PostForm("./index.html",
		url.Values{
			"user":     {"governor"},
			"action":   {a},
			"email":    "stepan.maks@gmail.com",
			"source":   {"governor"},
			"sourceid": {c}})
	if err != nil {
		log.Fatal(err)
	}
	logger.Print("new task is ready, user: " + u + ", action: " + a + ", email: " + e)
}

func predict(str string) string {
	body := getBody(str)

	URL := "http://governor-tf:8501/v1/models/governor:predict"
	req, err := http.NewRequest("POST", URL, bytes.NewBuffer(body))
	if err != nil {
		fmt.Println(err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	var p predictions
	err = json.Unmarshal(bodyBytes, &p)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(p.Predictions[0][0])
	if p.Predictions[0][0] > 0.5 {
		return "create"
	}
	return "reset"
}

func getBody(str string) []byte {
	wordIndexJSON := `{"<OOV>": 1, "new": 2, "my": 3, "password": 4, "please": 5, "reset": 6, "client": 7, "for": 8, "create": 9, "customer": 10, "could": 11, "you": 12, "a": 13, "account": 14, "the": 15, "need": 16, "hi": 17, "hello": 18, "i": 19, "ventus": 20, "add": 21, "it": 22, "user": 23, "to": 24, "system": 25, "me": 26, "acccount": 27, "creation": 28, "resete": 29, "username": 30, "our": 31, "signed": 32, "contract": 33, "we": 34, "added": 35, "process": 36, "onboard": 37, "with": 38, "company": 39}`
	wordIndex := make(map[string]int)
	err := json.Unmarshal([]byte(wordIndexJSON), &wordIndex)
	if err != nil {
		fmt.Println(err)
	}
	sequence := textToSequences(str, wordIndex)
	if err != nil {
		fmt.Println(err)
	}
	paddedSequence := padSequence(sequence, 10, 0)
	var body [][]int
	body = append(body, paddedSequence)
	sJSON, err := json.Marshal(body)
	if err != nil {
		fmt.Println(err)
	}
	reqBody := `{"instances" : ` + string(sJSON) + ` }`
	return []byte(reqBody)
}

func textToSequences(s string, wordIndex map[string]int) []int {
	var sequence []int
	fields := strings.Fields(s)
	if len(fields) == 0 {
		return sequence
	}
	for _, f := range fields {
		if val, ok := wordIndex[f]; ok {
			sequence = append(sequence, val)
		} else {
			sequence = append(sequence, 1)
		}
	}
	return sequence
}

func padSequence(s []int, length int, pad int) []int {
	diff := length - len(s)
	for i := 0; i < diff; i++ {
		s = append(s, pad)
	}
	return s
}
