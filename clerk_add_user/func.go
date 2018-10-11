package main

import (
	"crypto/rand"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/smtp"
	"os"
	"os/exec"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const MongoDb string = "35.232.89.65" //Mongo DB address

//const MongoDb string = "localhost" //Mongo DB address

var StdChars = []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789")

type task struct {
	ID     bson.ObjectId `json:"id" bson:"_id"`
	Number int64         `json:"number" bson:"number"`
	User   string        `json:"user" bson:"user"`
	Action string        `json:"action" bson:"action"`
	State  string        `json:"state" bson:"state"`
	Email  string        `json:"email" bson:"email"`
}

func run(cmd *exec.Cmd, Ticket task, password string, action string) {
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
		changeStatus(Ticket, "error")
		send(err.Error(), "Error detected "+"For support", "governorandclerk@gmail.com")
		//log.Fatal(err)
	} else if action == "create" {
		send("User has been created\n"+"Password is :"+password+"\n Verify your connection: RDP 35.231.245.199", "Account name is  "+Ticket.User, Ticket.Email)
		changeStatus(Ticket, "done")
		println("User created")
		fmt.Println("Done")

	}

}

func NewPassword(length int) string {
	return rand_char(length, StdChars)
}

func rand_char(length int, chars []byte) string {
	new_pword := make([]byte, length)
	random_data := make([]byte, length+(length/4)) // storage for random bytes.
	clen := byte(len(chars))
	maxrb := byte(256 - (256 % len(chars)))
	i := 0
	for {
		if _, err := io.ReadFull(rand.Reader, random_data); err != nil {
			panic(err)
		}
		for _, c := range random_data {
			if c >= maxrb {
				continue
			}
			new_pword[i] = chars[c%clen]
			i++
			if i == length {
				return string(new_pword)
			}
		}
	}
	panic("unreachable")
}

func putdataMongo(status task) {

	session, err := mgo.Dial(MongoDb)

	if err != nil {
		println("Error: Could not connect on MongoDB ")
	}
	defer session.Close()

	resultsCollection := session.DB("governor").C("results")
	err = resultsCollection.Insert(&status)

	if err != nil {
		println("Error: Could not update DB ")
	}

}

func getdataMongo(status string, action string) []task {

	session, err := mgo.Dial(MongoDb)

	if err != nil {
		println("Error: Could not connect  DB ")
	}
	var tasks []task

	c := session.DB("governor").C("tasks")

	err = c.Find(bson.M{"state": status, "action": action}).Sort("-timestamp").All(&tasks)

	if err != nil {
		println("Error: Could not update DB ")
	}
	fmt.Println("Results All: ", tasks)

	defer session.Close()

	return tasks

}

func addUser(Ticket task) {

	println("Login: ", Ticket.User)
	// cmd := exec.Command("PowerShell", "-Command", "Get-ADUser", "-LDAPFilter \"(SAMAccountName="+Ticket.User+")\"", "| select DistinguishedName ", "| ft -hide")
	// out, _ := cmd.CombinedOutput()

	// println("Output is : ", "\""+string(out)+"\"")

	password := NewPassword(20)
	println("Password :", password)
	run(exec.Command("PowerShell", "-Command", "New-ADUser", "-Name "+Ticket.User, "-UserPrincipalName "+Ticket.User, "-ChangePasswordAtLogon $false", "-AccountPassword (ConvertTo-SecureString -AsPlainText "+password+" -Force) ", "-Enabled $true "), Ticket, password, "create")
	run(exec.Command("PowerShell", "-Command", "Add-ADGroupMember", "-Identity \"Domain Admins\"", "-Members "+Ticket.User), Ticket, password, "add to group")

}

func changeStatus(Ticket task, state string) {

	println("changeStatus :", state)

	session, err := mgo.Dial(MongoDb)

	if err != nil {
		println("Error: Could not connect on MongoDB ")
	}

	c := session.DB("governor").C("tasks")

	// Update
	colQuerier := bson.M{"_id": Ticket.ID}
	change := bson.M{"$set": bson.M{"state": state}}
	err = c.Update(colQuerier, change)
	if err != nil {
		println("Error: Could not update DB ")
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
