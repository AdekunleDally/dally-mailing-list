package main

import (
	"database/sql"
	"log"

	"sync"

	"github.com/AdekunleDally/mailinglist/grpcapi"

	"github.com/AdekunleDally/mailinglist/jsonapi"
	"github.com/AdekunleDally/mailinglist/mdb"
	"github.com/alexflint/go-arg"
)

var args struct {
	DbPath   string `arg:"env:MAILINGLIST_DB"`
	BindJson string `arg:"env:MAILINGLIST_BIND_JSON"`
	BindGrpc string `arg:"env:MAILINGLIST_BIND_GRPC"`
}

func main() {
	arg.MustParse(&args)

	if args.DbPath == "" {
		args.DbPath = "list.db"
	}
	if args.BindJson == "" {
		args.BindJson = ":8080"
	}

	if args.BindGrpc == "" {
		args.BindGrpc = ":8081"
	}
	log.Printf("Using database '%v'\n", args.DbPath)
	db, err := sql.Open("sqlite3", args.DbPath)

	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	mdb.TryCreate(db)

	var wg sync.WaitGroup

	wg.Add(1)

	go func() {
		defer wg.Done()
		log.Printf("Starting JSON API Server...\n")
		jsonapi.Serve(db, args.BindJson)
	}()

	wg.Add(1)

	go func() {
		defer wg.Done()
		log.Printf("Starting GRPC API Server...\n")
		grpcapi.Serve(db, args.BindGrpc)
	}()

	wg.Wait()
}
