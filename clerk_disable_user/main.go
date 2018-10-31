package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"os/exec"
	"strings"
	"time"
)

const urlUsers string = "http://governor.verf.io/api/users/"
const urlAuth string = "https://verfio.auth0.com/oauth/token"

var myClient = &http.Client{Timeout: 10 * time.Second}

type task struct {
	ID       string `json:"id"`
	Number   int64  `json:"number"`
	Source   string `json:"source"`
	SourceID string `json:"sourceid"`
	User     string `json:"user"`
	Action   string `json:"action"`
	State    string `json:"state"`
	Email    string `json:"email"`
}

func main() {
	for {

		tickets := getTickets("active", "disable")

		for _, ticket := range *tickets {
			disableUser(&ticket)
		}

		time.Sleep(5 * time.Second)
	}
}

func run(cmd *exec.Cmd, ticket *task) {
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
	}
	err = cmd.Wait()
	if err != nil {
		println("Error")
		changeStatus(ticket, "error")
		send(err.Error(), "Error detected "+"For support", "governorandclerk@gmail.com")
		//log.Fatal(err)
	} else {
		send("Account was successfully disabled.", "Disable account: "+ticket.User, ticket.Email)
		changeStatus(ticket, "done")
		println("Action: account was disabled")
		fmt.Println("Done")
	}
}

func getTickets(status string, action string) *[]task {

	bearer := auth()
	req, _ := http.NewRequest("GET", urlUsers, nil)
	req.Header.Add("authorization", bearer)
	resp, err := myClient.Do(req)
	if err != nil {
		println("Error:", err)
	}
	defer resp.Body.Close()

	var ticketsTemp []task
	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	respByte := buf.Bytes()
	err = json.Unmarshal(respByte, &ticketsTemp)
	if err != nil {
		println("Error:", err)
	}

	var tickets []task
	for _, t := range ticketsTemp {
		if t.State == status && t.Action == action {
			tickets = append(tickets, t)
		}
	}

	fmt.Println("Results All: ", tickets)

	return &tickets
}

func changeStatus(ticket *task, state string) {
	println("changeStatus :", state)
	ticket.State = state

	var urlUser = urlUsers + ticket.ID
	j, err := json.Marshal(ticket)
	if err != nil {
		fmt.Println("Error marshaling ticket into JSON")
	}

	t := bytes.NewReader(j)

	bearer := auth()
	req, _ := http.NewRequest("POST", urlUser, t)
	req.Header.Add("authorization", bearer)
	req.Header.Add("contentType", "application/json")
	resp, err := myClient.Do(req)
	//resp, err := myClient.Post(urlUser, "application/json", t)
	if err != nil {
		fmt.Println("Error with POST request")
	}
	defer resp.Body.Close()
}

func readFile(filename string) string {

	bs, err := ioutil.ReadFile(filename)

	if err != nil {
		fmt.Println("Error:", err)
		//os.Exit(1)
	}

	pass := string(bs)

	return pass
}

func send(body string, subject string, email string) {

	from := "governorandclerk@gmail.com"
	pass := readFile("pass.txt")
	to := email

	msg := "From: " + from + "\n" +
		"To: " + to + "\n" +
		"Subject:" + subject + "\n\n" +
		body

	err := smtp.SendMail("smtp.gmail.com:587",
		smtp.PlainAuth("", from, pass, "smtp.gmail.com"),
		from, []string{to}, []byte(msg))

	if err != nil {
		log.Printf("smtp error: %s", err)
		return
	}
}

func disableUser(ticket *task) {

	println("Login: ", ticket.User)

	run(exec.Command("PowerShell", "-Command", "Disable-ADAccount", "-Identity "+ticket.User), ticket)
}

func auth() string {

	payload := strings.NewReader("{\"grant_type\":\"client_credentials\",\"client_id\": \"lIJmNudGywMs2JPzhayxCvTvnxb2YnRO\",\"client_secret\": \"IvqFrrtetMVRnj_zahi7nvBkgjolFM5xzTCPVbDyoFW8YmmqLUMB-vw2dHyyy-oG\",\"audience\": \"https://governor.verf.io/api\"}")

	req, err := http.NewRequest("POST", urlAuth, payload)
	if err != nil {
		log.Printf("Error creating POST request: %s", err)
	}

	req.Header.Add("content-type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("Error sending POST request: %s", err)
	}

	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)

	token := (strings.Split(string(body), ":"))[1]
	token = strings.Trim((strings.Split(token, ","))[0], "\"")
	bearer := "Bearer " + token

	return bearer
}
