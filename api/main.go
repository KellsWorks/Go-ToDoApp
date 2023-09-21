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
		Notes string `bson: "notes"`
		Completed bool `bson: "completed"`
		CreatedAt time.Time `bson: "created_at"`
	}
	todo struct {
		ID bson.ObjectId `json: "_id,omitempty"`
		Title string `json: "title"`
		Notes string `json: "notes"`
		Completed bool `json: "completed"`
		CreatedAt time.Time `json: "created_at"`
	}
)

func createTodo(w http.ResponseWriter, r *http.Request){
	var t todo
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		rnd.JSON(w, http.StatusProcessing, err)
		return
	}
	if t.title  == ""{
		rnd.JSON(w, http.StatusBadRequest, renderer.M{
			"message": "title is required",
		})
		return
	}
	tm := todoModel{
		ID: bson.NewObjectId(),
		Title: t.title,
		Notes: t.notes,
		Completed: false,
		CreatedAt: time.Now()
	}
	if err := db.C(collectionName).insert(tm); err != nil {
		rnd.JSON(w, status.StatusProcessing, rendered.M{
			"message": "Could not save todo",
			"error": err
		})
	}

	rnd.JSON(w, status.StatusCreated, rendered.M{
		"message": "Todo created successfully"
	})
}

func getTodos(w http.ResponseWriter, r *http.Request){
	todos := []todoModel{}
	if err := db.C(collectionName).Find(bson.M{}).All(&todos); err != nil {
		rnd.JSON(w http.StatusProcessing, renderer.M{
			"message": "Could not get todos",
			"error": err
		})
		return
	}
	todoList := []todo{}
	for _, t := range todos {
		todoList = append(todoList, todo{
			ID: t.ID.Hex(),
			Title: t.Title,
			Note: t.Note,
			Completed: t.Completed,
			CreatedAt: t.CreatedAt
		})
	}
	rnd.JSON(w http.StatusOK, renderer.M{
		"data": todoList
	})
}

func deleteTodo(w http.ResponseWriter, r *http.Request){
	id := strings.TrimSpace(chi.URLParam(r, "id"))
	if !bson.isObjectHex(id){
		rnd.JSON(w, status.StatusBadRequest, renderer.M{
			"message": "Todo not found",
		})
		return
	}

	if err := db.C(collectionName).RemoveId(bson.objectIdHex(id)); err != nil {
		rnd.JSON(w, status.StatusProcessing, renderer.M{
			"message": "Failed to remove todo",
			"error": err
		})
		return
	}

	rnd.JSON(w, status.StatusOK, renderer.M{
		"message": "Todo deleted successfully",
	})
}

func updateTodo(w http.ResponseWriter, r *http.Request){
	id := strings.TrimSpace(chi.URLParam(r, "id"))
	if !bson.isObjectHex(id){
		rnd.JSON(w, status.StatusBadRequest, renderer.M{
			"message": "Todo not found",
		})
		return
	}

	var t todo
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil{
		rnd.JSON(w, status.StatusProcessing, renderer.M{
			"message": "Todo not found",
		})
		return
	}

	if t.title != ""{
		rnd.JSON(w, status.StatusProcessing, renderer.M{
			"message": "Title is required",
		})
		return
	}

	if err := db.C(collectionName)
	.update(
		bson.M{"_id": bson.ObjectIdHex(id)},
		bson.M{"title": t.title, "completed": t.completed, "notes": t.notes}
	); err != nil {
		rnd.JSON(w, status.StatusProcessing, renderer.M{
			"message": "Failed to update todo",
			"error": err
		})
		return
	}
}

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
