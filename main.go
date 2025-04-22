package main

import (
	"context" // Potrzebne do zarządzania kontekstem operacji MongoDB
	"log"     // Do logowania błędów
	"net/http"
	"time" // Do ustawiania timeoutów

	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"          // Główny pakiet sterownika MongoDB
	"go.mongodb.org/mongo-driver/mongo/options"  // Opcje konfiguracyjne MongoDB
	"go.mongodb.org/mongo-driver/mongo/readpref" // Do sprawdzania połączenia
)

// Funkcja pomocnicza do nawiązywania połączenia z MongoDB
func connectDB() (*mongo.Client, error) {
	// Ustaw URI połączenia - zmień na swoje, jeśli jest inne
	// Możesz też wczytać to ze zmiennej środowiskowej lub pliku konfiguracyjnego
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")

	// Ustawienie timeoutu dla próby połączenia
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel() // Ważne, aby zwolnić zasoby kontekstu

	// Nawiązanie połączenia
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	// Sprawdzenie, czy połączenie działa (ping do serwera)
	ctxPing, cancelPing := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancelPing()
	err = client.Ping(ctxPing, readpref.Primary())
	if err != nil {
		log.Println("Nie udało się połączyć z MongoDB (ping):", err)
		// Możesz zdecydować, czy chcesz zakończyć aplikację, czy próbować dalej
		// client.Disconnect(context.Background()) // Rozłącz, jeśli ping się nie udał
		return nil, err
	}

	log.Println("Połączono z MongoDB!")
	return client, nil
}

func main() {
	// Nawiąż połączenie z MongoDB przy starcie aplikacji
	client, err := connectDB()
	if err != nil {
		log.Fatal("Nie udało się nawiązać połączenia z MongoDB:", err) // Zakończ aplikację, jeśli nie można się połączyć
	}

	// Zaplanuj rozłączenie przy zamykaniu aplikacji
	// Używamy kontekstu tła, ponieważ główny kontekst mógł już zostać anulowany
	defer func() {
		if err = client.Disconnect(context.Background()); err != nil {
			log.Fatal("Błąd podczas rozłączania z MongoDB:", err)
		}
		log.Println("Rozłączono z MongoDB.")
	}()

	router := gin.Default()

	// Serwowanie plików statycznych (Twoja konfiguracja frontendu)
	router.Use(static.Serve("/", static.LocalFile("./client/build", true)))

	// Grupa API
	api := router.Group("/api")
	{
		// Przekazanie klienta MongoDB do handlera
		// Można to zrobić na różne sposoby (np. przez middleware i kontekst Gin),
		// ale dla prostoty przekażemy go bezpośrednio.
		api.GET("/", func(ctx *gin.Context) {
			// Tutaj możesz użyć zmiennej 'client' do interakcji z bazą danych
			// Przykład: Pobranie kolekcji i wykonanie operacji
			// collection := client.Database("twoja_baza").Collection("twoja_kolekcja")
			// Tutaj możesz np. coś wstawić, znaleźć itp.

			// Na razie zwracamy prostą odpowiedź
			ctx.JSON(http.StatusOK, gin.H{
				"message":      "pong",
				"db_connected": true, // Dodajemy informację, że DB jest (powinno być) połączone
			})
		})

		// Możesz dodać więcej endpointów API, które korzystają z 'client'
		// np. api.POST("/items", func(ctx *gin.Context) { ... })
	}

	log.Println("Serwer uruchomiony na porcie :5000")
	// Uruchomienie serwera Gin
	if err := router.Run(":5000"); err != nil {
		log.Fatal("Nie udało się uruchomić serwera Gin:", err)
	}
}
