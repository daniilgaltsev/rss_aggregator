package main


import (
    "fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
)

func main() {
	fmt.Println("Starting the server...")

	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
		os.Exit(1)
	}

	port := os.Getenv("PORT")
	if port == "" {
		fmt.Println("PORT not set")
		os.Exit(1)
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
	router.Mount("/v1", v1Router)


	server := http.Server{
		Addr: "0.0.0.0:" + port,
		Handler: router,
	}

	err = server.ListenAndServe()

	fmt.Println("Server stopped")
	fmt.Println(err)
}
