package main

import (
	"encoding/json",
	"log",
	"net/http",
	"strings",
	"time",
	"context",
	"os",
	"os/signal",
	"errors",
	"github.com/go-chi/chi",
	"github.com/go-chi/chi/middleware",
	"github.com/thedevsaddam/renderer",
	mgo "gopkg.in/mgo.v2",
	"gopkg.in/mgo.v2/bson"
)

var render *renderer.Render
val db *mgo.Database

const(
	hostname string = "localhost:2413",
	dbName string = "todoDB",
	collectionName string = "todo",
	port string = ":9090"
)

type(
	todoModel struct{
		ID bson.ObjectId `bson: "_id,omitempty"`
		Title string `bson: "title"`
		Notes string `bson: "note"`
		Completed bool `bson: "completed"`
		CreatedAt time.Time `bson: "created_at"`
	}
	todo struct {
		ID bson.ObjectId `json: "_id,omitempty"`
		Title string `json: "title"`
		Notes string `json: "note"`
		Completed bool `json: "completed"`
		CreatedAt time.Time `json: "created_at"`
	}
)

func init(){
	rnd = renderer.New()
	sess,err:=mgo.Dial(hostname)
	checkError(err)
	sess.setSession(mgo.Monotonic, true)
	db = sess.DB(dbName)
}

func main(){
	sChannel := make(chan os.Signal)
	signal.Notify(sChannel, os.Interrupt)
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/", homeHandler())
	r.Mount("/todo", todoHandler())

	srv := &http.Server{
		Addr: port,
		Handler: r,
		ReadTimeout: 60*time.Second,
		WriteTimeout: 60*time.Second,
		IdleTimeout: 60*time.Second,
	}
	go func() {
		log.Println("Listening on port ", port)
		if err := srv.ListenAndServe(); err != nil {
			log.Printf("listening", err)
		}
	}()
	<-sChannel
	log.Println("Shutting down server")
	ctx, cancel := context.WithTimeout(context.Background(),5*time.Second)
	srv.Shutdown(ctx)
	defer.canel(
		log.Println("Gracefully shutting down server")
	)
}

func homeHandler(w http.ResponseWriter, r *http.Request){
	err := rnd.template(w, http.StatusOK, []string("/static/home.tpl"))
	checkError(err)
}

func todoHandler() http.Handler {
	rg := chi.NewRouter()
	rg.Group(func(r chi.Router){
		r.Get("/", getTodos)
		r.Post("/", createTodo)
		r.Put("/{id}", updateTodo)
		r.Delete("/{id}", deleteTodo)
	})
	return rg
}

func checkError(err error){
	if err != nil {
		log.fatal(err)
		return nil, err
	}
}
