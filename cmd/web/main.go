package main

import (
	"flag"
	"fmt"
	"go-stripe-payment/internal/driver"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
)

const version = "1.0.0"
const cssVersion = "1"

type config struct {
	port int
	env  string
	api  string
	db   struct {
		dsn string
	}
	stripe struct {
		secret string
		key    string
	}
}

type application struct {
	config        config
	infolog       *log.Logger
	errorlog      *log.Logger
	templateCache map[string]*template.Template
	version       string
}

func (app *application) serve() error {
	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", app.config.port),
		Handler:           app.routes(),
		IdleTimeout:       30 * time.Second,
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      30 * time.Second,
	}

	app.infolog.Printf("Starting HTTP server in %s mode on port %d", app.config.env, app.config.port)

	return srv.ListenAndServe()
}

func main() {
	var cfg config

	flag.IntVar(&cfg.port, "port", 4000, "Server port to listen on")
	flag.StringVar(&cfg.env, "env", "development", "Application environment {development|production}")
	flag.StringVar(&cfg.db.dsn, "dsn", "ashish:secret@tcp(localhost:3306)/widgets?parseTime=true&tls=false", "DSN")
	flag.StringVar(&cfg.api, "api", "http://localhost:4001", "URL to api")

	flag.Parse()

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	cfg.stripe.key = os.Getenv("STRIPE_KEY")
	cfg.stripe.key = os.Getenv("STRIPE_SECRET")

	infolog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorlog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	conn, err := driver.OpenDB(cfg.db.dsn)
	if err != nil {
		errorlog.Fatal(err)
	}
	defer conn.Close()

	tc := make(map[string]*template.Template)

	app := &application{
		config:        cfg,
		infolog:       infolog,
		errorlog:      errorlog,
		templateCache: tc,
		version:       version,
	}

	err = app.serve()
	if err != nil {
		app.errorlog.Println(err)
		log.Fatal(err)
	}
}
