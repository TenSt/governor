package main

import (
	"context"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/buger/jsonparser"

	//"github.com/rs/xid"

	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/bson/objectid"
	"github.com/mongodb/mongo-go-driver/mongo"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

// Task is task
type task struct {
	//ID     bson.ObjectID `bson:"_id,omitempty"`
	ID     objectid.ObjectID `json:"id" bson:"_id"`
	Number int64             `json:"number" bson:"number"`
	User   string            `json:"user" bson:"user"`
	Action string            `json:"action" bson:"action"`
	State  string            `json:"state" bson:"state"`
	Email  string            `json:"email" bson:"email"`
}

func sortTasks(s []task) []task {

	// fmt.Println(s)

	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}

	// fmt.Println(s)

	return s
}

func mongoWrite(user string, action string, email string) {
	client, err := mongo.NewClient("mongodb://mongo:27017")
	if err != nil {
		log.Fatal(err)
	}
	err = client.Connect(context.TODO())
	if err != nil {
		log.Fatal(err)
	}

	collection := client.Database("governor").Collection("tasks")

	id, _ := collection.Count(context.Background(), nil)

	newItemDoc := bson.NewDocument(bson.EC.Int64("number", id+1), bson.EC.String("user", user), bson.EC.String("action", action), bson.EC.String("state", "active"), bson.EC.String("email", email))
	_, err = collection.InsertOne(context.Background(), newItemDoc)

	if err != nil {
		log.Fatal(err)
	}

	defer client.Disconnect(context.Background())

	// id := res.InsertedID
	// log.Printf(id.(string))

}

func dropMongo() {
	client, err := mongo.NewClient("mongodb://mongo:27017")
	err = client.Connect(context.TODO())
	if err != nil {
		log.Fatal(err)
	}

	collection := client.Database("governor").Collection("tasks")

	err = collection.Drop(context.Background(), nil)

	//defer client.Disconnect(nil)

}

func dropHandler(w http.ResponseWriter, r *http.Request) {

	dropMongo()
	parseTasks()

}

func tasksHandler(w http.ResponseWriter, r *http.Request) {

	parseTasks()
	http.ServeFile(w, r, "tasks.html")

}

func jiraHandler(w http.ResponseWriter, r *http.Request) {

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}
	log.Println(string(body))

	d, _ := jsonparser.GetUnsafeString(body, "issue", "description")

	//description, _ := data["description"].(string)
	log.Println(d)

	u, a, e := parseDescription(d)
	log.Println(u + " " + a + " " + e)

	_, err = http.PostForm("http://governor.verf.io/index.html",
		url.Values{"user": {u}, "action": {a}, "email": {e}})
	if err != nil {
		log.Fatal(err)
	}
	log.Println("new task sent, user: " + u + ", action: " + a + ", email: " + e)
}

func readMongo() []task {
	client, err := mongo.NewClient("mongodb://mongo:27017")
	err = client.Connect(context.TODO())
	if err != nil {
		log.Fatal(err)
	}

	collection := client.Database("governor").Collection("tasks")

	cur, err := collection.Find(context.Background(), nil)

	if err != nil {
		log.Fatal(err)
	}

	defer cur.Close(context.Background())
	var tasks []task

	//err = ioutil.WriteFile("tasks.html", []byte(r))

	for cur.Next(context.Background()) {
		t := task{}
		err := cur.Decode(&t)
		if err != nil {
			log.Fatal(err)
		}

		tasks = append(tasks, t)
	}
	if err := cur.Err(); err != nil {
		log.Fatal(err)
	}

	defer client.Disconnect(context.Background())
	//	log.Print(tasks)
	return tasks

}

func parseDescription(s string) (string, string, string) {
	s = strings.ToLower(s)

	var u, a, e string

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
			e = strings.TrimRight(f[i], ".,!:?")
		}

	}
	return u, a, e
}

func parseTasks() {
	// parse template
	tpl, err := template.ParseFiles("tasks.gohtml")
	if err != nil {
		log.Fatalln(err)
	}

	// execute template
	//tasks := readMongo()
	tasks := readMongo()
	tasks = sortTasks(tasks)

	f, err := os.Create("tasks.html")
	if err != nil {
		log.Println("create file: ", err)
		return
	}

	err = tpl.Execute(f, tasks)
	if err != nil {
		log.Fatalln(err)
	}
}

func wasmHandler(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case "GET":
		parseTasks()
		http.ServeFile(w, r, "index.html")
	case "POST":

		// Call ParseForm() to parse the raw query and update r.PostForm and r.Form.
		if err := r.ParseForm(); err != nil {
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			return
		}
		//fmt.Fprintf(w, "Post from website! r.PostFrom = %v\n", r.PostForm)
		u := r.FormValue("user")
		a := r.FormValue("action")
		e := r.FormValue("email")

		mongoWrite(u, a, e)

		//readMongo()

		parseTasks()

	default:
		fmt.Fprintf(w, "Sorry, only GET and POST methods are supported.")
	}

}

func main() {

	files, err := filepath.Glob("*")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(files) // contains a list of all files in the current directory

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir(".")))
	mux.HandleFunc("/index.html", wasmHandler)
	mux.HandleFunc("/drop", dropHandler)
	mux.HandleFunc("/webhooks/jira", jiraHandler)
	mux.HandleFunc("/tasks.html", tasksHandler)
	log.Printf("server started")
	log.Fatal(http.ListenAndServe(":3000", mux))

}
