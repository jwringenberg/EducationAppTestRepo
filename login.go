package main

import (
       "database/sql"
       "encoding/json"
       "io/ioutil"
       "fmt"

       _ "github.com/mattn/go-sqlite3"
)

type Login struct {
   Type     string `json:"type"`
   Nickname string `json:"nickname"`
   Password string `json:"password"`
}
type LoginFile struct {
   Logins []Login `"json:logins"`
}

func check(e error) bool {
     return e != nil
}

func loadDB() {
     fmt.Println("Entering loadDB");
     
     database, _ := sql.Open("sqlite3", "./auth.db")

     dat, err := ioutil.ReadFile("./class.json")
     if check(err) {
     	return
     }
     res := LoginFile{}
     json.Unmarshal(dat, &res);

     insert := "INSERT INTO IANAuth (nickname, type, password) VALUES(?, ?, ?)";
     for _, row := range res.Logins {
     	 database.Exec(insert, row.Nickname, row.Type, row.Password);
     }
     database.Close()
}

func InitDB() {
     database, _ := sql.Open("sqlite3", "./auth.db")
     database.Exec("CREATE TABLE IF NOT EXISTS IANAuth (id INTEGER PRIMARY KEY, " +
                                                       "nickname TEXT, " +
		 			  	       "password TEXT, " +
						       "type TEXT)")
     rows, _ := database.Query("SELECT id FROM IANAuth")
     if !rows.Next() {
     	database.Close()
        loadDB()
     }
}

func ValidLogin(id, password string) bool {
     database, _ := sql.Open("sqlite3", "./auth.db")
     defer database.Close();
     rows, _ := database.Query("SELECT id FROM IANAuth WHERE nickname = ? AND password = ?")
     return rows.Next()
}

func IsTeacher(id string) bool {
     database, _ := sql.Open("sqlite3", "./auth.db")
     defer database.Close();
     rows, _ := database.Query("SELECT type FROM IANAuth WHERE nickname = ?")
     typ := ""
     rows.Scan(&typ)
     return typ == "TEACHER"
}