package main

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"os/exec"
	"time"
)

const urlUsers string = "http://governor.verf.io/api/users/"

var stdChars = []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789")

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

		tickets := getTickets("active", "create")

		for _, ticket := range *tickets {
			addUser(&ticket)
		}

		time.Sleep(5 * time.Second)
	}
}

func getTickets(status string, action string) *[]task {

	resp, err := myClient.Get(urlUsers)
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

func addUser(ticket *task) {

	println("Login: ", ticket.User)
	// cmd := exec.Command("PowerShell", "-Command", "Get-ADUser", "-LDAPFilter \"(SAMAccountName="+Ticket.User+")\"", "| select DistinguishedName ", "| ft -hide")
	// out, _ := cmd.CombinedOutput()

	// println("Output is : ", "\""+string(out)+"\"")

	password := newPassword(20)
	println("Password :", password)
	run(exec.Command("PowerShell", "-Command", "New-ADUser", "-Name "+ticket.User, "-UserPrincipalName "+ticket.User, "-ChangePasswordAtLogon $false", "-AccountPassword (ConvertTo-SecureString -AsPlainText "+password+" -Force) ", "-Enabled $true "), ticket, password, "create")
	run(exec.Command("PowerShell", "-Command", "Add-ADGroupMember", "-Identity \"Domain Admins\"", "-Members "+ticket.User), ticket, password, "add to group")
}

func newPassword(length int) string {
	return randChar(length, stdChars)
}

func randChar(length int, chars []byte) string {
	newPword := make([]byte, length)
	randomData := make([]byte, length+(length/4)) // storage for random bytes.
	clen := byte(len(chars))
	maxrb := byte(256 - (256 % len(chars)))
	i := 0
	for {
		if _, err := io.ReadFull(rand.Reader, randomData); err != nil {
			panic(err)
		}
		for _, c := range randomData {
			if c >= maxrb {
				continue
			}
			newPword[i] = chars[c%clen]
			i++
			if i == length {
				return string(newPword)
			}
		}
	}
}

func run(cmd *exec.Cmd, ticket *task, password string, action string) {
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
	} else if action == "create" {
		send("User has been created\n"+"Password is :"+password+"\n Verify your connection: RDP 35.231.245.199", "Account name is  "+ticket.User, ticket.Email)
		changeStatus(ticket, "done")
		println("User created")
		fmt.Println("Done")

	}
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
	resp, err := myClient.Post(urlUser, "application/json", t)
	if err != nil {
		fmt.Println("Error with POST request")
	}
	defer resp.Body.Close()
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

func readFile(filename string) string {

	bs, err := ioutil.ReadFile(filename)

	if err != nil {
		fmt.Println("Error:", err)
		//os.Exit(1)
	}

	pass := string(bs)

	return pass
}
