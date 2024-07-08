package main

import (
	"context"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"log"
	"time"
)

func main() {
	db, err := sqlx.Connect("mysql", "puser:1234@(localhost:16033)/aaa")
	if err != nil {
		log.Fatalln(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalln(err)
	}

	c := context.Background()
	d := time.Now().Add(5 * time.Second)
	cc, cancel := context.WithDeadline(c, d)
	defer cancel()

	var result []int
	err = db.SelectContext(cc, &result, "select id from user")
	if err != nil {
		log.Fatalln(err)
	}
	log.Println(result)

	r, err := db.ExecContext(cc, "insert into user values()") // if master dead, deadline occur after duration
	if err != nil {
		log.Fatalln("exec failed: ", err)
	}

	l, err := r.LastInsertId()
	if err != nil {
		log.Fatalln(err)
	}
	log.Println(l)
}
