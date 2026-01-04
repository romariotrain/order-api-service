// Package main - точка входа в наше приложение
// Здесь мы инициализируем все компоненты и запускаем HTTP сервер
package main

// Импортируем необходимые пакеты
import (
	"log"      // Пакет для логирования (вывода сообщений)
	"net/http" // Стандартный пакет для работы с HTTP
	"os"       // Пакет для работы с операционной системой (переменные окружения, и т.д.)

	"order-api-service/internal/database" // Наш пакет для работы с БД
)

// main - главная функция, с которой начинается выполнение программы
func main() {
	// Получаем строку подключения к БД из переменной окружения
	// os.Getenv читает переменную окружения DATABASE_URL
	// Если переменная не установлена, используется значение по умолчанию
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		// Значение по умолчанию для локальной разработки
		databaseURL = "postgres://orderuser:orderpass@localhost:5434/orderdb?sslmode=disable"
	}

	// Получаем порт сервера из переменной окружения
	serverPort := os.Getenv("SERVER_PORT")
	if serverPort == "" {
		// Значение по умолчанию - порт 8080
		serverPort = "8080"
	}

	// Выводим сообщение о запуске приложения
	log.Println("Запуск приложения Task Manager API...")

	// Подключаемся к базе данных
	// database.Connect возвращает подключение к БД или ошибку
	db, err := database.Connect(databaseURL)
	if err != nil {
		// log.Fatal выводит сообщение об ошибке и останавливает программу
		// Используется для критических ошибок, без которых приложение не может работать
		log.Fatal("Ошибка подключения к базе данных:", err)
	}

	// defer означает, что функция выполнится в конце main (перед завершением программы)
	// db.Close() закрывает соединение с БД
	// Это важно для корректного освобождения ресурсов
	defer db.Close()

	log.Println("Успешное подключение к базе данных")

	// Настраиваем роутинг (маршрутизацию) HTTP запросов
	// Создаем новый ServeMux - это маршрутизатор запросов
	// Он определяет, какая функция должна обработать каждый запрос
	mux := http.NewServeMux()

	// Оборачиваем наш mux в middleware для логирования
	// loggingMiddleware будет вызываться перед каждым запросом
	loggedMux := loggingMiddleware(mux)

	// Добавляем CORS middleware для разрешения запросов с других доменов
	// Это нужно, если фронтенд приложение работает на другом порте/домене
	corsHandler := corsMiddleware(loggedMux)

	// Формируем адрес сервера
	// ":" + serverPort создает строку вида ":8080"
	serverAddr := ":" + serverPort

	// Запускаем HTTP сервер
	// http.ListenAndServe слушает указанный адрес и обрабатывает запросы с помощью corsHandler
	// Эта функция блокирующая - программа будет работать, пока сервер не остановят
	err = http.ListenAndServe(serverAddr, corsHandler)

	// Если ListenAndServe вернул ошибку, значит сервер не смог запуститься
	if err != nil {
		log.Fatal("Ошибка запуска сервера:", err)
	}
}

// methodHandler - helper функция для проверки HTTP метода
// Возвращает обработчик, который проверяет метод запроса
// handlerFunc - функция-обработчик, которую нужно вызвать
// allowedMethod - разрешенный HTTP метод (GET, POST, PUT, DELETE)
func methodHandler(handlerFunc http.HandlerFunc, allowedMethod string) http.HandlerFunc {
	// Возвращаем новую функцию-обработчик
	return func(w http.ResponseWriter, r *http.Request) {
		// Проверяем, совпадает ли метод запроса с разрешенным
		if r.Method != allowedMethod {
			// Если метод не совпадает, возвращаем ошибку 405 (Method Not Allowed)
			http.Error(w, "Метод не разрешен", http.StatusMethodNotAllowed)
			return
		}
		// Если метод совпадает, вызываем оригинальный обработчик
		handlerFunc(w, r)
	}
}

func loggingMiddleware(next http.Handler) http.Handler {
	// Возвращаем новый обработчик
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Логируем информацию о входящем запросе
		// r.Method - HTTP метод (GET, POST, и т.д.)
		// r.URL.Path - путь запроса (/tasks, /tasks/1, и т.д.)
		// r.RemoteAddr - IP адрес клиента
		log.Printf("%s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)

		// Вызываем следующий обработчик в цепочке
		next.ServeHTTP(w, r)
	})
}

// corsMiddleware - middleware для настройки CORS (Cross-Origin Resource Sharing)
// CORS позволяет браузерам делать запросы к API с других доменов
// Без этого фронтенд приложение не сможет обращаться к нашему API
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Устанавливаем заголовки CORS

		// Access-Control-Allow-Origin - разрешаем запросы с любых доменов
		// В продакшене здесь должен быть конкретный домен, а не "*"
		w.Header().Set("Access-Control-Allow-Origin", "*")

		// Access-Control-Allow-Methods - разрешенные HTTP методы
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

		// Access-Control-Allow-Headers - разрешенные заголовки в запросе
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// Обрабатываем preflight запрос
		// Браузеры отправляют OPTIONS запрос перед основным запросом для проверки CORS
		if r.Method == "OPTIONS" {
			// Отвечаем статусом 200 и завершаем обработку
			w.WriteHeader(http.StatusOK)
			return
		}

		// Вызываем следующий обработчик
		next.ServeHTTP(w, r)
	})
}
