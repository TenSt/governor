package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Task is task
type task struct {
	//ID     bson.ObjectID `bson:"_id,omitempty"`
	ID       primitive.ObjectID `json:"id" bson:"_id"`
	Number   string             `json:"number" bson:"number"`
	Source   string             `json:"source" bson:"source"`
	SourceID string             `json:"sourceid" bson:"sourceid"`
	User     string             `json:"user" bson:"user"`
	Action   string             `json:"action" bson:"action"`
	State    string             `json:"state" bson:"state"`
	Email    string             `json:"email" bson:"email"`
}

type predictions struct {
	Predictions [][]float64 `json:"predictions"`
}

func main() {
	files, err := filepath.Glob("*")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(files) // contains a list of all files in the current directory

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir(".")))
	mux.HandleFunc("/tasks", tasks)
	// mux.HandleFunc("/tasks.html", tasksHandler)
	log.Printf("server started")
	log.Fatal(http.ListenAndServe(":3000", mux))
}

func mongoWrite(user string, action string, email string, source string, sourceid string) {
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://mongo:27017"))
	if err != nil {
		log.Fatal(err)
	}
	err = client.Connect(context.TODO())
	if err != nil {
		log.Fatal(err)
	}

	collection := client.Database("governor").Collection("tasks")

	id, _ := collection.CountDocuments(context.Background(), bson.D{})

	newItemDoc := bson.D{
		primitive.E{Key: "number", Value: id + 1},
		primitive.E{Key: "user", Value: user},
		primitive.E{Key: "action", Value: action},
		primitive.E{Key: "state", Value: "active"},
		primitive.E{Key: "email", Value: email},
		primitive.E{Key: "source", Value: source},
		primitive.E{Key: "sourceid", Value: sourceid},
	}

	_, err = collection.InsertOne(context.Background(), newItemDoc)

	if err != nil {
		log.Fatal(err)
	}

	defer client.Disconnect(context.Background())

	// id := res.InsertedID
	// log.Printf(id.(string))
}

func mongoRead() []task {
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://mongo:27017"))
	err = client.Connect(context.TODO())
	if err != nil {
		log.Println("Error on connecting to mongo-db for tasks")
		log.Println(err)
	}

	collection := client.Database("governor").Collection("tasks")

	cur, err := collection.Find(context.Background(), bson.D{})

	if err != nil {
		log.Println("Error on finding document in collection for tasks")
		log.Println(err)
	}

	defer cur.Close(context.Background())
	var tasks []task

	//err = ioutil.WriteFile("tasks.html", []byte(r))

	for cur.Next(context.Background()) {
		t := task{}
		err := cur.Decode(&t)
		if err != nil {
			log.Println("Error decoding document")
			log.Println(err)
		}

		tasks = append(tasks, t)
	}
	if err := cur.Err(); err != nil {
		log.Fatal(err)
	}

	defer client.Disconnect(context.Background())

	return tasks
}

func tasks(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case "GET":
		log.Println("Method is:\t" + r.Method)
		log.Println("Request URL is:\t" + r.RequestURI)

		tasks := mongoRead()

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		err := json.NewEncoder(w).Encode(tasks)
		if err != nil {
			log.Println(err)
		}

	case "POST":
		log.Println("Method is:\t" + r.Method)
		log.Println("Request URL is:\t" + r.RequestURI)

		// mongoWrite(u, a, e, s, si)

	default:
		fmt.Fprintf(w, "Sorry, only GET and POST methods are supported.")
	}
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
