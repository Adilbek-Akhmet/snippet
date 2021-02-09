package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/golangcollege/sessions"
	"github.com/jackc/pgx/v4/pgxpool"
	"html/template"
	"log"
	"net/http"
	"os"
	"se01.com/pkg/models/postgres"
	"time"
)

type application struct {
	errorLog      *log.Logger
	infoLog       *log.Logger
	session       *sessions.Session
	snippets      *postgres.SnippetModel
	templateCache map[string]*template.Template
}

func main() {
	connString := "postgres://web:pass@host.docker.internal:5432/snippet"
	dsn := flag.String("dsn", connString, "PostgreSQL data source name")

	addr := flag.String("addr", ":4000", "HTTP Network Address")

	secret := flag.String("secret", "s6Ndh+pPbnzHbS*+9Pk8qGWhTzbpa@ge", "Secret key")
	flag.Parse()

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	conn, err := openDB(*dsn)
	if err != nil {
		errorLog.Fatal(err)
	}
	defer conn.Close()

	var checkDBConn string
	err = conn.QueryRow(context.Background(), "select 'Connected to database'").Scan(&checkDBConn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)
		os.Exit(1)
	}
	infoLog.Print(checkDBConn)

	templateCache, err := newTemplateCache("./ui/html/")
	if err != nil {
		errorLog.Fatal(err)
	}

	session := sessions.New([]byte(*secret))
	session.Lifetime = 12 * time.Hour

	app := application{
		errorLog:      errorLog,
		infoLog:       infoLog,
		session:       session,
		snippets:      &postgres.SnippetModel{DB: conn},
		templateCache: templateCache,
	}

	srv := &http.Server{
		Addr:     *addr,
		Handler:  app.routes(),
		ErrorLog: app.errorLog,
	}

	app.infoLog.Printf("Starting a server on %s", *addr)
	err = srv.ListenAndServe()
	app.errorLog.Fatal(err)

}

func openDB(dsn string) (*pgxpool.Pool, error) {
	conn, err := pgxpool.Connect(context.Background(), dsn)
	if err != nil {
		return nil, err
	}

	return conn, nil
}
