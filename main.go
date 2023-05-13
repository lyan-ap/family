package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

type IOItem struct {
	ID       int64
	Mobile   string
	Wechat   string
	JsonData map[string]interface{}
}

type SearchParams struct {
	Mobile  string
	Wechat  string
	PageNum int
}

type PageData struct {
	List  []IOItem `json:"list"`
	Total int      `json:"total"`
}

type Response struct {
	Data    PageData `json:"data"`
	Message string   `json:"message,omitempty"`
	Error   string   `json:"error,omitempty"`
	Status  int      `json:"status"`
}

type ResponseDetail struct {
	Data IOItem 
	Message string   `json:"message,omitempty"`
	Error   string   `json:"error,omitempty"`
	Status  int      `json:"status"`
}

type ResponseCommon struct {
	Message string   `json:"message,omitempty"`
	Error   string   `json:"error,omitempty"`
	Status  int      `json:"status"`
}
type JSONData map[string]interface{}

func dbConn() (db *sql.DB) {
	db, err := sql.Open("mysql", "tester:secret@tcp(db:3306)/test")

	if err != nil {
		log.Print(err.Error())
	}
	fmt.Println("Connected to MySQL database")

	return db
}

func getTableName(r *http.Request) string {
	pathSegments := strings.Split(r.URL.Path, "/")
	lastSegment := pathSegments[1]
	fmt.Println("path data:", pathSegments[1])

	return lastSegment
}

func enableCors(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	w.Header().Set("Content-Type", "application/json")
}

func getTotalRowCount(db *sql.DB, tableName string, searchParams SearchParams) (int, error) {
	var count int

	query := "SELECT COUNT(*) FROM " + tableName + " WHERE 1=1"

	if searchParams.Wechat != "" {
		query += " AND wechat LIKE '%" + searchParams.Wechat + "%'"
	}
	if searchParams.Mobile != "" {
		query += " AND mobile LIKE '%" + searchParams.Mobile + "%'"
	}

	err := db.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func getUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		enableCors(w, r)
		db := dbConn()
		defer db.Close()
		var users []IOItem
		var searchParams SearchParams
		err := json.NewDecoder(r.Body).Decode(&searchParams)

		if err != nil {
			http.Error(w, "Error parsing request body", http.StatusBadRequest)
			response := Response{Data: PageData{List:users,Total:0}, Status: http.StatusBadRequest, Message: "Error parsing request body"}

			fmt.Println("response data:", response, searchParams)
			json.NewEncoder(w).Encode(response)
			return
		}

		table := getTableName(r)
		limit := 20
		offset := (searchParams.PageNum - 1) * limit

		results, err := db.Query("SELECT * FROM " + table +" WHERE mobile LIKE ? AND wechat LIKE ? LIMIT ? OFFSET ?",
		"%"+searchParams.Mobile+"%", "%"+searchParams.Wechat+"%", limit, offset)
		if err != nil {
			panic(err.Error()) // proper error handling instead of panic in your app
		}

		for results.Next() {
			var jsonData json.RawMessage
			var u IOItem
			// for each row, scan the result into our tag composite object
			if err := results.Scan(&u.ID, &u.Mobile, &u.Wechat, &jsonData); err != nil {
				panic(err.Error())
			}

			err = json.Unmarshal(jsonData, &u.JsonData)
			if err != nil {
				log.Fatal(err)
			}

			users = append(users, u)
		}
		total, err := getTotalRowCount(db, table,searchParams)
		if err != nil {
				log.Fatal(err)
		}
		response := Response{Data: PageData{List:users,Total:total}, Status: http.StatusOK, Message: "success"}

		fmt.Println("Endpoint Hit: usersPage")
		json.NewEncoder(w).Encode(response)
	}else{
		response := Response{ Status: http.StatusForbidden, Message: "error method"}
		json.NewEncoder(w).Encode(response)
	}
}

func addUser(w http.ResponseWriter, r *http.Request) {
	enableCors(w, r)
	db := dbConn()
	defer db.Close()
	if r.Method == http.MethodPost {
		table := getTableName(r)
		var data IOItem
		err := json.NewDecoder(r.Body).Decode(&data)
		if err != nil {
			http.Error(w, "Error parsing request body", http.StatusBadRequest)
			return
		}
		fmt.Println("path data:", table)

		jsonValue, err := json.Marshal(data.JsonData)
		if err != nil {
			http.Error(w, "Error encoding JSON data", http.StatusInternalServerError)
			return
		}

		insForm, err := db.Prepare("INSERT INTO " + table + "(mobile, wechat, json_data) VALUES(?,?,?)")
		if err != nil {
			panic(err.Error())
		}
		insForm.Exec(data.Mobile, data.Wechat, jsonValue)
		response := ResponseCommon{ Status: http.StatusCreated, Message: "success"}
		json.NewEncoder(w).Encode(response)
	}else{
		response := Response{ Status: http.StatusForbidden, Message: "error method"}
		json.NewEncoder(w).Encode(response)
	}

}

func delUser(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodDelete {
		enableCors(w, r)
		db := dbConn()
		defer db.Close()
		emp := r.URL.Query().Get("id")
		table := getTableName(r)

		delForm, err := db.Prepare("DELETE FROM " + table + " WHERE id=?")
		if err != nil {
			panic(err.Error())
		}
		delForm.Exec(emp)
		response := ResponseCommon{ Status: http.StatusOK, Message: "success"}
		json.NewEncoder(w).Encode(response)
	}else{
		response := Response{ Status: http.StatusForbidden, Message: "error method"}
		json.NewEncoder(w).Encode(response)
	}
}

func updateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		enableCors(w, r)
		db := dbConn()
		defer db.Close()
		table := getTableName(r)
		id := r.FormValue("id")
		mobile := r.FormValue("mobile")
		wechat := r.FormValue("wechat")

		updateForm, err := db.Prepare("UPDATE " + table + " SET mobile=?, wechat=?  WHERE id=?")
		if err != nil {
			panic(err.Error())
		}
		updateForm.Exec(mobile, wechat, id)
		log.Println("UPDATE: Title: " + mobile)
		response := ResponseCommon{ Status: http.StatusOK, Message: "success"}
		json.NewEncoder(w).Encode(response)

	}else{
		response := Response{ Status: http.StatusForbidden, Message: "error method"}
		json.NewEncoder(w).Encode(response)
	}
}

func getDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		enableCors(w, r)
		db := dbConn()
		defer db.Close()
		id := r.URL.Query().Get("id")
		table := getTableName(r)

		rows, err := db.Query("SELECT * FROM "+table+" WHERE id=?", id)
		defer rows.Close()

		var data IOItem
		for rows.Next() {
			var jsonData json.RawMessage
			if err := rows.Scan(&data.ID, &data.Mobile, &data.Wechat, &jsonData); err != nil {
				log.Fatal(err)
			}
			err = json.Unmarshal(jsonData, &data.JsonData)
			if err != nil {
				log.Fatal(err)
			}
		}
		response := ResponseDetail{Data: data, Status: http.StatusOK, Message: "success"}

		// fmt.Println("response data:", data)
		json.NewEncoder(w).Encode(response)
	}else{
		response := Response{ Status: http.StatusForbidden, Message: "error method"}
		json.NewEncoder(w).Encode(response)
	}

}
func main() {
	http.HandleFunc("/io/users", getUsers)
	http.HandleFunc("/la/users", getUsers)
	http.HandleFunc("/io/users/add", addUser)
	http.HandleFunc("/la/users/add", addUser)
	http.HandleFunc("/io/users/detail", getDetail)
	http.HandleFunc("/la/users/detail", getDetail)
	http.HandleFunc("/io/users/del", delUser)
	http.HandleFunc("/la/users/del", delUser)
	http.HandleFunc("/io/users/update", updateUser)
	http.HandleFunc("/la/users/update", updateUser)
	log.Fatal(http.ListenAndServe(":1357", nil))
}
