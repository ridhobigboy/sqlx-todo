package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

var db *sqlx.DB

type Lable struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

type TodoCore struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Label       []*Lable `json:"label"`
	UpdateAt    string   `json:"updateAt"`
}

type Todo struct {
	Id int `json:"id"`
	TodoCore
}

type InsertTodoLabel struct {
	Id int   `json:"id"`
	L  Lable `json:"label"`
}

type STodoChangeLabelColor struct {
	TodoId  int `json :"todo_id"`
	LabelId int `json:"label_id"`
	Color   int `json:"color"`
}

func TodoCangeLabelColor(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	t := &STodoChangeLabelColor{}
	err := decoder.Decode(t)
	if err != nil {
		w.WriteHeader(500)
		fmt.Println(err.Error())

		return
	}

	sr := "$[" + strconv.Itoa(t.LabelId) + "].color"
	_, err = db.Exec(`UPDATE todo
	SET label = json_replace(label,? , ?)
	WHERE id = ?;`, sr, t.Color, t.TodoId)
	if err != nil {
		w.WriteHeader(500)
		fmt.Println(err.Error())

		return
	}
	w.Header().Set("Conten-Type", "json")
	w.Write([]byte(`{"status":"success"}`))
}

func TodoInsertLabel(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	t := &InsertTodoLabel{}
	err := decoder.Decode(t)
	if err != nil {
		w.WriteHeader(500)
		fmt.Println(err.Error())

		return
	}

	labelJson, err := json.Marshal(t.L)
	_, err = db.Exec(`
	UPDATE todo
		SET label = json_insert(
		  label,
		  '$[' || json_array_length(label) || ']',
		  json(?)
		)
		WHERE id = ?;`, string(labelJson), t.Id)
	if err != nil {
		w.WriteHeader(500)
		fmt.Println(err.Error())

		return
	}
	w.Header().Set("COnten-Type", "json")
	w.Write([]byte(`{"status":"success"}`))
}

func TodoInsert(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	t := &TodoCore{}
	err := decoder.Decode(t)
	if err != nil {
		w.WriteHeader(500)
		fmt.Println(err.Error())

		return
	}
	labelJson, err := json.Marshal(t.Label)
	_, err = db.Exec("INSERT INTO todo VALUES(?, ?, ?, ?, ?)", 404, t.Title, t.Description, string(labelJson), t.UpdateAt)

	if err != nil {
		w.WriteHeader(500)
		fmt.Println(err.Error())

		return
	}
	w.Header().Set("Conten-Type", "json")
	w.Write([]byte(`{"status": "success"}`))
}

func TodoViewAll(w http.ResponseWriter, r *http.Request) {
	var sT []*Todo
	rows, _ := db.Queryx("SELECT * FROM todo")
	for rows.Next() {
		t := &Todo{}
		var temp string
		rows.Scan(&t.Id, &t.Title, &t.Description, &temp, &t.UpdateAt)
		err := json.Unmarshal([]byte(temp), &t.Label)

		if err != nil {
			fmt.Println(err.Error())
		}
		sT = append(sT, t)
	}
	j, _ := json.Marshal(sT)
	w.Header().Set("Content-Type", "json")
	w.Write(j)
}

func TodoViewCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	labelName := vars["label"]
	var sT []*Todo
	rows, err := db.Queryx("SELECT todo.* FROM todo, json_each(todo.label) WHERE json_extract(json_each.value, '$.name') LIKE ?", "%"+labelName+"%")
	if err != nil {
		fmt.Println(err.Error())
		w.WriteHeader(500)

		return
	}

	for rows.Next() {
		t := &Todo{}
		var temp string
		rows.Scan(&t.Id, &t.Title, &t.Description, &temp, &t.UpdateAt)
		err := json.Unmarshal([]byte(temp), &t.Label)

		if err != nil {
			fmt.Println(err.Error())
		}
		sT = append(sT, t)
	}
	j, _ := json.Marshal(sT)
	w.Header().Set("Conten-Type", "json")
	w.Write(j)
}

func main() {

	db, _ = sqlx.Open("sqlite3", "./todo.db")

	r := mux.NewRouter()
	r.HandleFunc("/todo", TodoViewAll).Methods("GET")
	r.HandleFunc("/todo", TodoInsert).Methods("POST")
	r.HandleFunc("/todo/{label}", TodoViewCategory).Methods("GET")
	r.HandleFunc("/InsertTodoLabel", TodoInsertLabel).Methods("POST")
	r.HandleFunc("/changeTodoLabelColor", TodoCangeLabelColor).Methods("POST")

	addr := fmt.Sprintf("localhost:%v", 8000)
	fmt.Println("Server run at port" + addr)
	err := http.ListenAndServe(addr, r)
	if err != nil {
		fmt.Println(err.Error())
	}
}
