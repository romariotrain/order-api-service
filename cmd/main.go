package main

import (
	"log"
	"net/http"
	"os"

	"order-api-service/internal/database"
	"order-api-service/internal/handlers"
	"order-api-service/internal/middlewares"
)

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://orderuser:orderpass@localhost:5434/orderdb?sslmode=disable"
	}

	serverPort := os.Getenv("SERVER_PORT")
	if serverPort == "" {
		serverPort = "8080"
	}

	log.Println("Запуск приложения Order API...")

	db, err := database.Connect(databaseURL)
	if err != nil {
		log.Fatal("Ошибка подключения к базе данных:", err)
	}
	defer db.Close()

	orderStore := database.NewOrderStore(db)
	h := handlers.NewHandler(orderStore)

	mux := http.NewServeMux()

	// Один endpoint /orders, разные методы
	mux.HandleFunc("/orders", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.GetAllOrders(w, r)
		case http.MethodPost:
			h.CreateOrder(w, r)
		default:
			http.Error(w, "Метод не разрешен", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/healthz", h.Health)

	mux.HandleFunc("/orders/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Метод не разрешен", http.StatusMethodNotAllowed)
			return
		}
		h.GetOrderByID(w, r)
	})

	// Middleware цепочка
	var handler http.Handler = mux
	handler = middlewares.Logging(handler)
	handler = middlewares.CORS(handler)

	serverAddr := ":" + serverPort
	log.Println("HTTP сервер слушает", serverAddr)

	if err := http.ListenAndServe(serverAddr, handler); err != nil {
		log.Fatal("Ошибка запуска сервера:", err)
	}
}
