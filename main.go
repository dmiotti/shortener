package main

import (
	"fmt"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	h "github.com/dmiotti/shortener/api"
	mr "github.com/dmiotti/shortener/repository/mongodb"
	rr "github.com/dmiotti/shortener/repository/redis"

	"github.com/dmiotti/shortener/shortener"
)

func main() {
	repo := chooseRepo()
	service := shortener.NewRedirectService(repo)
	handler := h.NewHandler(service)

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/{code}", handler.Get)
	r.Post("/", handler.Post)

	errs := make(chan error, 2)
	go func() {
		port := httpPort()
		fmt.Printf("Listening on port %s\n", port)
		errs <- http.ListenAndServe(httpPort(), r)
	}()

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT)
		errs <- fmt.Errorf("%s", <- c)
	}()

	fmt.Printf("Terminated %s", <- errs)
}

func httpPort() string {
	port := "8000"
	if os.Getenv("PORT") != "" {
		port = os.Getenv("PORT")
	}
	return fmt.Sprintf(":%s", port)
}

func chooseRepo() shortener.RedirectRepository {
	switch os.Getenv("URL_DB") {
	case "redis":
		redisURL := os.Getenv("REDIS_URL")
		repo, err := rr.NewRedisRepository(redisURL)
		if err != nil {
			log.Fatal(err)
		}
		return repo
	case "mongo":
		mongoURL := os.Getenv("MONGO_URL")
		mongoDB := os.Getenv("MONGO_DB")
		mongoTimeout, _ := strconv.Atoi(os.Getenv("MONGO_TIMEOUT"))
		repo, err := mr.NewMongoRepository(mongoURL, mongoDB, mongoTimeout)
		if err != nil {
			log.Fatal(err)
		}
		return repo
	}
	return nil
}