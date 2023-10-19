package main

import _ "github.com/lib/pq"
import (
	"database/sql"
    "fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"

	"github.com/daniilgaltsev/rss_aggregator/internal/database"
)

type apiConfig struct {
	DB *database.Queries
}

func (config *apiConfig) postUsersHandler(w http.ResponseWriter, r *http.Request) {
	handlePostUsers(w, r, config.DB)
}

func (config *apiConfig) getUsersHandler(w http.ResponseWriter, r *http.Request) {
	handleGetUsers(w, r, config.DB)
}

func (config *apiConfig) postFeedsHandler(w http.ResponseWriter, r *http.Request) {
	handlePostFeeds(w, r, config.DB)
}

func (config *apiConfig) getFeedsHandler(w http.ResponseWriter, r *http.Request) {
	handleGetFeeds(w, r, config.DB)
}

func (config *apiConfig) PostFeedFollowsHandler(w http.ResponseWriter, r *http.Request) {
	handlePostFeedFollows(w, r, config.DB)
}

func (config *apiConfig) deleteFeedFollowsHandler(w http.ResponseWriter, r *http.Request) {
	handleDeleteFeedFollows(w, r, config.DB)
}

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
		os.Exit(1)
	}

	fmt.Println("Connecting to the database...")
	dbURL := os.Getenv("DBURL")
	if dbURL == "" {
		fmt.Println("DBURL not set")
		os.Exit(1)
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		fmt.Println("Error connecting to the database:", err)
		os.Exit(1)
	}
	dbQueries := database.New(db)

	fmt.Println("Starting the server...")
	port := os.Getenv("PORT")
	if port == "" {
		fmt.Println("PORT not set")
		os.Exit(1)
	}

	config := apiConfig{
		DB: dbQueries,
	}
	router := chi.NewRouter()
	router.Use(cors.Handler(
		cors.Options{
			AllowedOrigins: []string{"https://*", "http://*"},
			AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders: []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
			ExposedHeaders: []string{"Link"},
			AllowCredentials: false,
			MaxAge: 300,
		},
	))

	v1Router := chi.NewRouter()
	v1Router.Get("/readiness", readinessHandler)
	v1Router.Get("/err", errHandler)
	v1Router.Post("/users", config.postUsersHandler)
	v1Router.Get("/users", config.getUsersHandler)
	v1Router.Post("/feeds", config.postFeedsHandler)
	v1Router.Get("/feeds", config.getFeedsHandler)
	v1Router.Post("/feed_follows", config.PostFeedFollowsHandler)
	v1Router.Delete("/feed_follows/{feedFollowId}", config.deleteFeedFollowsHandler)
	router.Mount("/v1", v1Router)


	server := http.Server{
		Addr: "0.0.0.0:" + port,
		Handler: router,
	}

	err = server.ListenAndServe()

	fmt.Println("Server stopped")
	fmt.Println(err)
}
