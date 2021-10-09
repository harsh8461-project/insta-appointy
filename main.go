package main

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client

//structure of the USER
type User struct {
	Id       primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Name     string             `json:"name,omitempty" bson:"name,omitempty"`
	Email    string             `json:"email,omitempty" bson:"email,omitempty"`
	Password string             `json:"password,omitempty" bson:"password,omitempty"`
}

//strucure of the POST
type Post struct {
	Id              primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Uid             string             `json:"uid,omitempty" bson:"uid,omitempty"`
	Caption         string             `json:"caption,omitempty" bson:"caption,omitempty"`
	ImageURL        string             `json:"imageURL,omitempty" bson:"imageURL,omitempty"`
	PostedTimestamp string             `json:"timestamp,omitempty" bson:"timestamp,omitempty"`
}

//Function which gives the API Endpoint for Adding New Users <POST REQUEST>
func addUsers(w http.ResponseWriter, r *http.Request) {
	fmt.Println("method:", r.Method)
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			return
		}
		w.Header().Add("content-type", "application/json")
		var user User
		json.NewDecoder(r.Body).Decode(&user)
		md5HashInBytes := md5.Sum([]byte(user.Password))
		md5HashInString := hex.EncodeToString(md5HashInBytes[:])
		u1 := bson.D{{Key: "name", Value: user.Name},
			{Key: "email", Value: user.Email}, {Key: "password", Value: md5HashInString}}
		collection := client.Database("Mern").Collection("users")
		ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
		result, _ := collection.InsertOne(ctx, u1)
		json.NewEncoder(w).Encode(result)
	}
}

//Function which gives API Endpoint for Getting User info <GET REQUEST>
func Userinfo(w http.ResponseWriter, r *http.Request) {
	fmt.Println("method:", r.Method)
	if r.Method == http.MethodGet {
		w.Header().Add("content-type", "application/json")
		var myExp = regexp.MustCompile(`/users/(?P<id>[a-zA-Z0-9_]+)`)
		match := myExp.FindStringSubmatch(r.URL.Path)
		result := make(map[string]string)
		for i, name := range myExp.SubexpNames() {
			if i != 0 && name != "" {
				result[name] = match[i]
			}
		}
		key := result["id"]
		id, _ := primitive.ObjectIDFromHex(key)
		var user User
		collection := client.Database("Mern").Collection("users")
		ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
		err := collection.FindOne(ctx, User{Id: id}).Decode(&user)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{ "message": "` + err.Error() + `" }`))
			return
		}
		json.NewEncoder(w).Encode(user)
	}
}

//Function which gives the API Endpoint for Adding New Posts <POST REQUEST>
func addPosts(w http.ResponseWriter, r *http.Request) {
	fmt.Println("method:", r.Method)
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			return
		}
		w.Header().Add("content-type", "application/json")
		var post Post
		json.NewDecoder(r.Body).Decode(&post)
		p1 := bson.D{{Key: "uid", Value: post.Uid},
			{Key: "caption", Value: post.Caption}, {Key: "imageURL", Value: post.ImageURL},
			{Key: "timestamp", Value: time.Now()}}
		collection := client.Database("Mern").Collection("posts")
		ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
		result, _ := collection.InsertOne(ctx, p1)
		json.NewEncoder(w).Encode(result)
	}
}

//Function which gives API Endpoint for Getting Post info <GET REQUEST>
func Postinfo(w http.ResponseWriter, r *http.Request) {
	fmt.Println("method:", r.Method)
	if r.Method == http.MethodGet {
		w.Header().Add("content-type", "application/json")
		var myExp = regexp.MustCompile(`/posts/(?P<id>[a-zA-Z0-9_]+)`)
		match := myExp.FindStringSubmatch(r.URL.Path)
		result := make(map[string]string)
		for i, name := range myExp.SubexpNames() {
			if i != 0 && name != "" {
				result[name] = match[i]
			}
		}
		key := result["id"]
		id, _ := primitive.ObjectIDFromHex(key)
		fmt.Println(id)
		var post Post
		collection := client.Database("Mern").Collection("posts")
		ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
		err := collection.FindOne(ctx, Post{Id: id}).Decode(&post)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{ "message": "` + err.Error() + `" }`))
			return
		}
		json.NewEncoder(w).Encode(post)
	}
}

//Function which gives API Endpoint for Getting All Posts by a particular User <GET REQUEST> with pagination implimentation
func AllPost(w http.ResponseWriter, r *http.Request) {
	fmt.Println("method:", r.Method)
	if r.Method == http.MethodGet {
		q, ok := r.URL.Query()["p"]
		var q1 int64 = 1
		if !ok || len(q[0]) < 1 {
			log.Println("Url Param 'p' is missing")
		} else {
			pp, _ := strconv.Atoi(q[0])
			q1 = int64(pp)
		}
		var perPage int64 = 4 //4 data per page
		findOptions := options.Find()
		findOptions.SetLimit(perPage)
		findOptions.SetSkip((int64(q1) - 1) * perPage)

		var myExp = regexp.MustCompile(`/posts/users/(?P<id>[a-zA-Z0-9_]+)`)
		match := myExp.FindStringSubmatch(r.URL.Path)
		result := make(map[string]string)
		for i, name := range myExp.SubexpNames() {
			if i != 0 && name != "" {
				result[name] = match[i]
			}
		}
		key := result["id"]
		w.Header().Add("content-type", "application/json")
		var posts []Post
		collection := client.Database("Mern").Collection("posts")
		ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
		cursor, err := collection.Find(ctx, Post{Uid: key}, findOptions)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message": "` + err.Error() + `"}`))
			return
		}
		defer cursor.Close(ctx)
		for cursor.Next(ctx) {
			var post Post
			cursor.Decode(&post)
			posts = append(posts, post)
		}
		if err := cursor.Err(); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message": "` + err.Error() + `"}`))
			return
		}
		json.NewEncoder(w).Encode(posts)
	}
}

//This function handles all the requests for our API
func handleRequests() {

	fmt.Println("After DB")
	http.HandleFunc("/users", addUsers)
	http.HandleFunc("/users/", Userinfo)
	http.HandleFunc("/posts", addPosts)
	http.HandleFunc("/posts/", Postinfo)
	http.HandleFunc("/posts/users/", AllPost)
	log.Fatal(http.ListenAndServe(":8081", nil))
}

func main() {
	fmt.Println("Starting the application...")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	clientOptions := options.Client().ApplyURI("mongodb+srv://iwpharsh:harsh84618461@cluster0.uh8g1.mongodb.net/Mern?retryWrites=true&w=majority")
	client, _ = mongo.Connect(ctx, clientOptions)
	handleRequests()
}
