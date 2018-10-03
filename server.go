package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"

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
	User   string            `json:"user" bson:"user"`
	Action string            `json:"action" bson:"action"`
	State  string            `json:"state" bson:"state"`
}

func sortTasks(s []task) []task {

	// fmt.Println(s)

	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}

	// fmt.Println(s)

	return s
}

func mongoWrite(user string, action string) {
	client, err := mongo.NewClient("mongodb://mongo:27017")
	if err != nil {
		log.Fatal(err)
	}
	err = client.Connect(context.TODO())
	if err != nil {
		log.Fatal(err)
	}

	collection := client.Database("governor").Collection("tasks")

	newItemDoc := bson.NewDocument(bson.EC.String("user", user), bson.EC.String("action", action), bson.EC.String("state", "active"))
	_, err = collection.InsertOne(context.Background(), newItemDoc)

	if err != nil {
		log.Fatal(err)
	}
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

}

func dropHandler(w http.ResponseWriter, r *http.Request) {

	dropMongo()
	parseTasks()

}

func tasksHandler(w http.ResponseWriter, r *http.Request) {

	parseTasks()
	http.ServeFile(w, r, "tasks.html")

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

	return tasks
	//	log.Print(tasks)
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

		mongoWrite(u, a)

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
	mux.HandleFunc("/tasks.html", tasksHandler)
	log.Printf("server started")
	log.Fatal(http.ListenAndServe(":3000", mux))
}
