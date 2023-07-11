package main

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type user struct {
	id        int
	username  string
	password  string
	createdAt time.Time
}

func main() {
	// DB open
	db, err := sql.Open("mysql", "username:password@(127.0.0.1:3306)/dbname?parseTime=true")

	if err != nil {
		fmt.Println("db open error")
		return
	}

	{
		// Create Query
		query := `
		CREATE TABLE users (
			id INT AUTO_INCREMENT,
			username TEXT NOT NULL,
			password TEXT NOT NULL,
			created_at DATETIME,
			PRIMARY KEY (id)
		)
	`
		_, err := db.Exec(query)
		if err != nil {
			fmt.Printf("query %s error", query)
			return
		}
	}

	{
		username := "johndoe"
		password := "secret"
		createdAt := time.Now()
		result, errInsert := db.Exec("INSERT INTO users (username, password, created_at) VALUES (?, ?, ?)", username, password, createdAt)

		if errInsert != nil {
			fmt.Println("Insert error")
			return
		}

		userID, errLastId := result.LastInsertId()
		if errLastId != nil {
			fmt.Println("Get Last Insert Id Error")
			return
		}
		fmt.Println(userID)
	}

	{
		var (
			id        int
			username  string
			password  string
			createdAt time.Time
		)
		query := `SELECT id, username, password, created_at FROM users WHERE id = ?`
		err := db.QueryRow(query, 1).Scan(&id, &username, &password, &createdAt)

		if err != nil {
			fmt.Printf("Select query %s Error \n", query)
		}
	}

	{
		rows, err := db.Query(`SELECT id, username, password, created_At FROM users`)
		if err != nil {
			fmt.Println(err)
		}
		var users []user
		for rows.Next() {
			var u user
			err := rows.Scan(&u.id, &u.username, &u.password, &u.createdAt)

			if err != nil {
				fmt.Println(err)
			}

			users = append(users, u)
		}
		{
			err := rows.Err()
			if err != nil {
				fmt.Println(err)
			}
		}
	}

	{
		_, err := db.Exec("DELETE FROM users WHERE id = ?", 1)
		if err != nil {
			fmt.Println(err)
		}
	}
}
