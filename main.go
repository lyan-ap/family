package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"log"
	"net/http"
	// "io/ioutil"
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
	List []IOItem `json:"list"`
	Total int         `json:"total"`
}

type Response struct {
	Data    PageData `json:"data"`
	Message string      `json:"message,omitempty"`
	Error   string      `json:"error,omitempty"`
	Status  int         `json:"status"`
}

func dbConn() (db *sql.DB) {

		// Open up our database connection.
		db, err := sql.Open("mysql", "tester:secret@tcp(db:3306)/test")

		// if there is an error opening the connection, handle it
		if err != nil {
			log.Print(err.Error())
		}
		fmt.Println("Connected to MySQL database")

		defer db.Close()
	return db
}


func enableCors(w http.ResponseWriter, r *http.Request) {
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

func makePage(tbName string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		enableCors(w, r)
		db := dbConn()
		defer db.Close()

		if r.Method == "POST" {
			var searchParams SearchParams
			var list []IOItem

			err := json.NewDecoder(r.Body).Decode(&searchParams)

			if err != nil {
				http.Error(w, "Error parsing request body", http.StatusBadRequest)
				response := Response{Data: PageData{List:list,Total:0}, Status: http.StatusBadRequest, Message: "Error parsing request body"}

				fmt.Println("response data:", response, searchParams)
				json.NewEncoder(w).Encode(response)
				return
			}

			limit := 20
			offset := (searchParams.PageNum - 1) * limit
			rows, err := db.Query("SELECT * FROM "+tbName+" WHERE mobile LIKE ? AND wechat LIKE ? LIMIT ? OFFSET ?",
				"%"+searchParams.Mobile+"%", "%"+searchParams.Wechat+"%", limit, offset)
			if err != nil {
				panic(err.Error())
			}
			defer rows.Close()

			for rows.Next() {
				var jsonData json.RawMessage
				var data IOItem
				if err := rows.Scan(&data.ID, &data.Mobile, &data.Wechat, &jsonData); err != nil {
					log.Fatal(err)
				}	// return nil, fmt.Errorf("Index %v", err)
				
				err = json.Unmarshal(jsonData, &data.JsonData)
				if err != nil {
					log.Fatal(err)
				}
				list = append(list, data)
			}

            total, err := getTotalRowCount(db, tbName,searchParams)
            if err != nil {
                log.Fatal(err)
            }
			response := Response{Data: PageData{List:list,Total:total}, Status: http.StatusOK, Message: "success"}

			fmt.Println("response data:", response)
			json.NewEncoder(w).Encode(response)

			defer db.Close()
		}
	}
}

func makeDelete(tbName string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		enableCors(w, r)
		db := dbConn()
		if r.Method == "DELETE" {
			emp := r.URL.Query().Get("id")
			delForm, err := db.Prepare("DELETE FROM " + tbName + " WHERE id=?")
			if err != nil {
				panic(err.Error())
			}
			delForm.Exec(emp)
			log.Println("DELETE")
		}
		defer db.Close()
	}
}

func makeUpdate(tbName string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		enableCors(w, r)

		db := dbConn()
		if r.Method == "POST" {
			id := r.FormValue("id")
			mobile := r.FormValue("mobile")
			wechat := r.FormValue("wechat")
			insForm, err := db.Prepare("UPDATE " + tbName + " SET mobile=? ,wechat=? WHERE id=?")
			if err != nil {
				panic(err.Error())
			}
			insForm.Exec(mobile, wechat, id)
			fmt.Println(mobile, wechat)
			log.Println("UPDATE: Title: " + mobile)
		}
		defer db.Close()
	}
}

func makeInsert(tbName string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		enableCors(w, r)
		db := dbConn()
		if r.Method == "POST" {
			var data IOItem
			err := json.NewDecoder(r.Body).Decode(&data)
			if err != nil {
				http.Error(w, "Error parsing request body", http.StatusBadRequest)
				return
			}
			// fmt.Println("Request data:", data)

			jsonValue, err := json.Marshal(data.JsonData)
			if err != nil {
				http.Error(w, "Error encoding JSON data", http.StatusInternalServerError)
				return
			}

			insForm, err := db.Prepare("INSERT INTO " + tbName + "(mobile, wechat,json_data) VALUES(?,?,?)")
			if err != nil {
				panic(err.Error())
			}
			insForm.Exec(data.Mobile, data.Wechat, jsonValue)
		}
		w.WriteHeader(http.StatusCreated)

		defer db.Close()
	}
}


func makeDetail(tbName string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		db := dbConn()
		if r.Method == "GET" {
			id := r.URL.Query().Get("id")
			rows, err := db.Query("SELECT * FROM "+tbName+" WHERE id=?", id)
			if err != nil {
				panic(err.Error())
			}
			if err != nil {
				panic(err.Error())
			}
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
			// fmt.Println("response data:", data)
			json.NewEncoder(w).Encode(data)
			defer db.Close()
		}
	}
}
func homePage(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode("response")

}

func main() {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/io/list", makePage("io"))
	router.HandleFunc("/la/list", makePage("la"))
	router.HandleFunc("/insert/io", makeInsert("io"))
	router.HandleFunc("/insert/la", makeInsert("la"))
	router.HandleFunc("/io/delete", makeDelete("io"))
	router.HandleFunc("/la/delete", makeDelete("la"))
	router.HandleFunc("/io/update", makeUpdate("io"))
	router.HandleFunc("/la/update", makeUpdate("la"))
	router.HandleFunc("/io/detail", makeDetail("io"))
	router.HandleFunc("/la/detail", makeDetail("la"))
	
	c := cors.New(cors.Options{
		AllowedMethods:     []string{http.MethodGet, http.MethodPost, http.MethodDelete, http.MethodOptions},
		AllowCredentials:   true,
		AllowedHeaders:     []string{"Content-Type", "Bearer", "Bearer ", "content-type", "Origin", "Accept"},
		OptionsPassthrough: true,
        AllowedOrigins: []string{"http://localhost:8889","http://localhost:8080", "http://localhost:3000","https://yanlee26.github.io"},
		// Enable Debugging for testing, consider disabling in production
		Debug: true,
	})

	handler := c.Handler(router)

	log.Fatal(http.ListenAndServe(":8080", handler))
}
