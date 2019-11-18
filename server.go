package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	jwtmiddleware "github.com/auth0/go-jwt-middleware"
	"github.com/buger/jsonparser"
	"github.com/codegangsta/negroni"
	"github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type dns struct {
	ID         primitive.ObjectID `json:"id" bson:"_id"`
	Number     int64              `json:"number" bson:"number"`
	Source     string             `json:"source" bson:"source"`
	SourceID   string             `json:"sourceid" bson:"sourceid"`
	Record     string             `json:"record" bson:"record"`
	RecordType string             `json:"recordtype" bson:"recordtype"`
	Zone       string             `json:"zone" bson:"zone"`
	Target     string             `json:"target" bson:"target"`
	Action     string             `json:"action" bson:"action"`
	State      string             `json:"state" bson:"state"`
	Email      string             `json:"email" bson:"email"`
}

//Read DNS API data from mongo
func readDNS() []dns {
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://mongo:27017"))
	err = client.Connect(context.TODO())
	if err != nil {
		log.Fatal(err)
	}

	collection := client.Database("governor").Collection("dns")

	cur, err := collection.Find(context.Background(), nil)

	if err != nil {
		log.Fatal(err)
	}

	defer cur.Close(context.Background())
	var dnses []dns

	//err = ioutil.WriteFile("tasks.html", []byte(r))

	for cur.Next(context.Background()) {
		d := dns{}
		err := cur.Decode(&d)
		if err != nil {
			log.Fatal(err)
		}

		dnses = append(dnses, d)
	}
	if err := cur.Err(); err != nil {
		log.Fatal(err)
	}

	defer client.Disconnect(context.Background())
	//	log.Print(tasks)
	return dnses

}

//write DNS API data to mongo
func writeDNS(record string, recordtype string, zone string, target string, action string, email string, source string, sourceid string) {
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://mongo:27017"))
	if err != nil {
		log.Fatal(err)
	}
	err = client.Connect(context.TODO())
	if err != nil {
		log.Fatal(err)
	}

	collection := client.Database("governor").Collection("dns")

	id, _ := collection.CountDocuments(context.Background(), nil)

	newItemDoc := bson.D{
		{"number", id + 1},
		{"record", record},
		{"recordtype", recordtype},
		{"zone", zone},
		{"target", target},
		{"action", action},
		{"state", "active"},
		{"email", email},
		{"source", source},
		{"sourceid", sourceid},
	}
	_, err = collection.InsertOne(context.Background(), newItemDoc)

	if err != nil {
		log.Fatal(err)
	}

	defer client.Disconnect(context.Background())
}

//handle requests to /api/dns
func dnsHandler(w http.ResponseWriter, r *http.Request) {

	p := strings.Split(r.URL.Path, "/")
	fmt.Println(p)
	d := readDNS()

	switch r.Method {

	case "GET":
		if p[3] != "" {
			for _, dns := range d {
				// var j []byte
				id := `ObjectID("` + p[3] + `")`
				if (dns.ID).String() == id {
					j, err := json.Marshal(dns)
					if err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return
					}
					w.Header().Set("Content-Type", "application/json")
					w.Write(j)

				}
			}

		} else {

			j, err := json.Marshal(d)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.Write(j)
		}

	case "POST":
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			panic(err)
		}
		log.Println(string(body))
		var d dns
		err = json.Unmarshal(body, &d)
		if err != nil {
			panic(err)
		}
		log.Println(d)

		client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://mongo:27017"))
		err = client.Connect(context.TODO())
		if err != nil {
			log.Fatal(err)
		}

		collection := client.Database("governor").Collection("dns")
		doc, err := bson.Marshal(d)
		filter := bson.D{{"_id", d.ID}}

		res, err := collection.ReplaceOne(context.Background(), filter, doc)
		log.Println(res.UpsertedID)

		if res.UpsertedID == nil {
			writeDNS(d.Record, d.RecordType, d.Zone, d.Target, d.Action, d.Email, d.Source, d.SourceID)

		}

	default:
	}
}
func getPemCert(token *jwt.Token) (string, error) {
	cert := ""
	resp, err := http.Get("https://verfio.auth0.com/.well-known/jwks.json")

	if err != nil {
		return cert, err
	}
	defer resp.Body.Close()

	var jwks = Jwks{}
	err = json.NewDecoder(resp.Body).Decode(&jwks)

	if err != nil {
		return cert, err
	}

	for k := range jwks.Keys {
		if token.Header["kid"] == jwks.Keys[k].Kid {
			cert = "-----BEGIN CERTIFICATE-----\n" + jwks.Keys[k].X5c[0] + "\n-----END CERTIFICATE-----"
		}
	}

	if cert == "" {
		err := errors.New("Unable to find appropriate key")
		return cert, err
	}

	return cert, nil
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

//Response struct for auth0
type Response struct {
	Message string `json:"message"`
}

//Jwks struct for auth0
type Jwks struct {
	Keys []JSONWebKeys `json:"keys"`
}

//JSONWebKeys struct for auth0
type JSONWebKeys struct {
	Kty string   `json:"kty"`
	Kid string   `json:"kid"`
	Use string   `json:"use"`
	N   string   `json:"n"`
	E   string   `json:"e"`
	X5c []string `json:"x5c"`
}

// Task is task
type task struct {
	//ID     bson.ObjectID `bson:"_id,omitempty"`
	ID       primitive.ObjectID `json:"id" bson:"_id"`
	Number   int64              `json:"number" bson:"number"`
	Source   string             `json:"source" bson:"source"`
	SourceID string             `json:"sourceid" bson:"sourceid"`
	User     string             `json:"user" bson:"user"`
	Action   string             `json:"action" bson:"action"`
	State    string             `json:"state" bson:"state"`
	Email    string             `json:"email" bson:"email"`
}

func sortTasks(s []task) []task {

	// fmt.Println(s)

	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}

	// fmt.Println(s)

	return s
}

func sortDNS(s []dns) []dns {

	// fmt.Println(s)

	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}

	// fmt.Println(s)

	return s
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

	id, _ := collection.CountDocuments(context.Background(), nil)

	newItemDoc := bson.D{
		{"number", id + 1},
		{"user", user},
		{"action", action},
		{"state", "active"},
		{"email", email},
		{"source", source},
		{"sourceid", sourceid},
	}

	_, err = collection.InsertOne(context.Background(), newItemDoc)

	if err != nil {
		log.Fatal(err)
	}

	defer client.Disconnect(context.Background())

	// id := res.InsertedID
	// log.Printf(id.(string))

}

func dropMongo() {
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://mongo:27017"))
	err = client.Connect(context.TODO())
	if err != nil {
		log.Fatal(err)
	}

	collection := client.Database("governor").Collection("tasks")

	err = collection.Drop(context.Background())

	//defer client.Disconnect(nil)

}

func dropHandler(w http.ResponseWriter, r *http.Request) {

	dropMongo()
	parseTasks()

}

func dnsHTMLHandler(w http.ResponseWriter, r *http.Request) {

	parseDNS()
	http.ServeFile(w, r, "dns.html")

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

	d, _ := jsonparser.GetUnsafeString(body, "issue", "fields", "description")
	s := "jira"
	si, _ := jsonparser.GetUnsafeString(body, "issue", "key")

	//description, _ := data["description"].(string)
	log.Println(d)

	u, a, e := parseDescription(d)
	log.Println(u + " " + a + " " + e)

	_, err = http.PostForm("http://governor.verf.io/index.html",
		url.Values{"user": {u}, "action": {a}, "email": {e}, "source": {s}, "sourceid": {si}})
	if err != nil {
		log.Fatal(err)
	}
	log.Println("new task sent, user: " + u + ", action: " + a + ", email: " + e)
}

func servicenowHandler(w http.ResponseWriter, r *http.Request) {

	log.Println("servicenow request")
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}
	log.Println(string(body))

	// d, _ := jsonparser.GetUnsafeString(body, "issue", "fields", "description")
	// s := "jira"
	// si, _ := jsonparser.GetUnsafeString(body, "issue", "key")

	// //description, _ := data["description"].(string)
	// log.Println(d)

	f := strings.Fields(string(body))

	u, a, e := parseDescription(string(body))
	s := "servicenow"
	si := f[0]

	// log.Println(u + " " + a + " " + e)

	_, err = http.PostForm("http://governor.verf.io/index.html",
		url.Values{"user": {u}, "action": {a}, "email": {e}, "source": {s}, "sourceid": {si}})
	if err != nil {
		log.Fatal(err)
	}
	log.Println("new task sent, user: " + u + ", action: " + a + ", email: " + e)
}

func usersHandler(w http.ResponseWriter, r *http.Request) {

	p := strings.Split(r.URL.Path, "/")

	t := readMongo()

	switch r.Method {

	case "GET":
		if p[3] != "" {
			for _, task := range t {
				// var j []byte
				id := `ObjectID("` + p[3] + `")`
				if (task.ID).String() == id {
					j, err := json.Marshal(task)
					if err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return
					}
					w.Header().Set("Content-Type", "application/json")
					w.Write(j)

				}
			}

		} else {

			j, err := json.Marshal(t)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.Write(j)
		}

	case "POST":
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			panic(err)
		}
		log.Println(string(body))
		var t task
		err = json.Unmarshal(body, &t)
		if err != nil {
			panic(err)
		}
		log.Println(t)

		client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://mongo:27017"))
		err = client.Connect(context.TODO())
		if err != nil {
			log.Fatal(err)
		}

		collection := client.Database("governor").Collection("tasks")
		doc, err := bson.Marshal(t)
		filter := bson.D{{"_id", t.ID}}

		_, err = collection.ReplaceOne(context.Background(), filter, doc)

	default:
	}

}

func readMongo() []task {
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://mongo:27017"))
	err = client.Connect(context.TODO())
	if err != nil {
		log.Println("Error on connecting to mongo-db for tasks")
		log.Fatal(err)
	}

	collection := client.Database("governor").Collection("tasks")

	cur, err := collection.Find(context.Background(), nil)

	if err != nil {
		log.Println("Error on finding document in collection for tasks")
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
			// if strings.Contains(word, "|mailto:") {
			// 	e = (strings.SplitAfter(f[i], "|"))[0]
			// 	e = strings.Trim(f[i], ".,!:?]")
			// } else {
			// 	e = strings.TrimRight(f[i], ".,!:?]")
			// }
			if strings.Contains(word, "|") {
				e = (strings.SplitAfter(f[i], "|"))[0]
				e = strings.Trim(e, "[].,!:?]|")
			} else {
				e = strings.Trim(f[i], "[].,!:?]")
			}

		}

	}
	return u, a, e
}

func parseDNS() {
	// parse template
	tpl, err := template.ParseFiles("dns.gohtml")
	if err != nil {
		log.Fatalln(err)
	}

	// execute template
	//tasks := readMongo()
	dnses := readDNS()
	dnses = sortDNS(dnses)

	f, err := os.Create("dns.html")
	if err != nil {
		log.Println("create file: ", err)
		return
	}

	err = tpl.Execute(f, dnses)
	if err != nil {
		log.Fatalln(err)
	}
}

func parseTasks() {
	// parse template
	tpl, err := template.ParseFiles("tasks.gohtml")
	if err != nil {
		log.Println("Error on parseTasks func")
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
		log.Println("Method is:\t" + r.Method)
		log.Println("Request URL is:\t" + r.RequestURI)
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
		s := r.FormValue("source")
		si := r.FormValue("sourceid")

		mongoWrite(u, a, e, s, si)

		//readMongo()

		parseTasks()

	default:
		fmt.Fprintf(w, "Sorry, only GET and POST methods are supported.")
	}

}

func main() {

	//
	// err := godotenv.Load()
	// if err != nil {
	// 	log.Print("Error loading .env file")
	// }

	jwtMiddleware := jwtmiddleware.New(jwtmiddleware.Options{
		ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
			// Verify 'aud' claim
			//aud := os.Getenv("AUTH0_AUDIENCE")
			aud := "https://governor.verf.io/api"
			checkAud := token.Claims.(jwt.MapClaims).VerifyAudience(aud, false)
			if !checkAud {
				return token, errors.New("Invalid audience")
			}
			// Verify 'iss' claim
			//iss := "https://" + os.Getenv("AUTH0_DOMAIN") + "/"
			iss := "https://verfio.auth0.com/"
			checkIss := token.Claims.(jwt.MapClaims).VerifyIssuer(iss, false)
			if !checkIss {
				return token, errors.New("Invalid issuer")
			}

			cert, err := getPemCert(token)
			if err != nil {
				panic(err.Error())
			}

			result, _ := jwt.ParseRSAPublicKeyFromPEM([]byte(cert))
			return result, nil
		},
		SigningMethod: jwt.SigningMethodRS256,
	})

	// c := cors.New(cors.Options{
	// 	AllowedOrigins:   []string{"http://localhost:3000"},
	// 	AllowCredentials: true,
	// 	AllowedHeaders:   []string{"Authorization"},
	// })
	//

	files, err := filepath.Glob("*")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(files) // contains a list of all files in the current directory

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir(".")))
	mux.HandleFunc("/index.html", wasmHandler)
	//mux.HandleFunc("/drop", dropHandler)
	mux.HandleFunc("/webhooks/jira", jiraHandler)
	mux.HandleFunc("/webhooks/servicenow", servicenowHandler)
	//mux.HandleFunc("/api/users/", usersHandler)
	//
	mux.Handle("/api/users/", negroni.New(
		negroni.HandlerFunc(jwtMiddleware.HandlerWithNext),
		negroni.Wrap(http.HandlerFunc(usersHandler)),
	))
	//
	mux.HandleFunc("/tasks.html", tasksHandler)
	mux.HandleFunc("/dns.html", dnsHTMLHandler)
	mux.HandleFunc("/api/dns/", dnsHandler)
	log.Printf("server started")
	log.Fatal(http.ListenAndServe(":3000", mux))

}
