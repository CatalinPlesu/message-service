package application

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/CatalinPlesu/message-service/handler"
	"github.com/CatalinPlesu/message-service/repository/message"
)

func (a *App) loadRoutes() {
	router := chi.NewRouter()

	router.Use(middleware.Logger)

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	router.Route("/messages", a.loadMessageRoutes)

	a.router = router
}

func (a *App) loadMessageRoutes(router chi.Router) {
	messageHandler := &handler.Message{
		Repo: &message.RedisRepo{
			Client: a.rdb,
		},
	}

	router.Post("/", messageHandler.Create)                                   
	router.Get("/", messageHandler.List)                                      
	router.Get("/{id}", messageHandler.GetByID)                               
	router.Put("/{id}", messageHandler.UpdateByID)                            
	router.Delete("/{id}", messageHandler.DeleteByID)                         
}
