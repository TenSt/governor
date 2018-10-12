package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/smtp"
	"os"
	"os/exec"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

//MongoDb const is a Mongo DB address
const MongoDb string = "35.232.89.65"

type task struct {
	ID     bson.ObjectId `json:"id" bson:"_id"`
	Number int64         `json:"number" bson:"number"`
	User   string        `json:"user" bson:"user"`
	Action string        `json:"action" bson:"action"`
	State  string        `json:"state" bson:"state"`
	Email  string        `json:"email" bson:"email"`
}

func run(cmd *exec.Cmd, Ticket task) {
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
	} else {
		send("Account was successfully disabled.", "Disable account: "+Ticket.User, Ticket.Email)
		changeStatus(Ticket, "done")
		println("Action: account was disabled")
		fmt.Println("Done")
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

func disableUser(Ticket task) {

	println("Login: ", Ticket.User)

	run(exec.Command("PowerShell", "-Command", "Disable-ADAccount", "-Identity "+Ticket.User), Ticket)
}
