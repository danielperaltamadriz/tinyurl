package api

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/danielperaltamadriz/tinyurl/config"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/oklog/ulid/v2"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const (
	_defaultPort = 8080
)

type API struct {
	server *echo.Echo
	cfg    config.API

	db db
}

func NewAPI(cfg config.Config) (*API, error) {
	db, err := NewDB(cfg.DB)
	if err != nil {
		return nil, fmt.Errorf("NewDB: %w", err)
	}
	if cfg.API.Port == 0 {
		cfg.API.Port = _defaultPort
	}
	server := echo.New()

	server.Use(middleware.Logger())
	server.Use(middleware.Recover())
	api := &API{
		cfg: cfg.API,
		db:  *db,
	}
	server.GET("/tiny/:id", api.getTiny)
	server.POST("/tiny", api.postTiny)

	api.server = server

	return api, nil
}

func (a *API) Shutdown() error {
	fmt.Println("Shutting down server")
	err := a.server.Shutdown(context.Background())
	if err != nil {
		log.Printf("failed to close server: %v", err)
	}
	return a.db.Shutdown()
}

func (a *API) Handler() http.HandlerFunc {
	return a.server.ServeHTTP
}

func (a *API) Start() error {
	err := a.server.Start(":" + strconv.Itoa(a.cfg.Port))
	if err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}
	return nil
}

type TinyURL struct {
	ID  string `gorm:"primaryKey"`
	URL string
}

func (a *API) postTiny(c echo.Context) error {
	var request struct {
		URL string `json:"url"`
	}

	if err := c.Bind(&request); err != nil {
		return c.JSON(400, map[string]string{"error": "Invalid JSON"})
	}

	println("Received POST request with URL:", request.URL)
	model := &TinyURL{
		ID:  ulid.Make().String(),
		URL: request.URL,
	}
	tx := a.db.client.Create(model)
	if tx.Error != nil {
		c.Logger().Errorf("failed to create value in db, error: %v", tx.Error)
		return c.NoContent(http.StatusInternalServerError)
	}

	return c.JSON(201, Response{
		TinyURL: "/tiny/" + model.ID,
	})
}
func (a *API) getTiny(c echo.Context) error {
	id := c.Param("id")
	resp := TinyURL{
		ID: id,
	}
	tx := a.db.client.First(&resp)
	if tx.Error != nil {
		c.Logger().Infof("failed to find element with id: %s, error:%v", id, tx.Error)
		return c.NoContent(http.StatusInternalServerError)
	}
	if tx.RowsAffected == 0 {
		c.Logger().Infof("url with id: %s not found", id)
		return c.NoContent(http.StatusNotFound)
	}
	return c.Redirect(http.StatusFound, resp.URL)
}

type db struct {
	cfg    config.DB
	client *gorm.DB
}

func NewDB(config config.DB) (*db, error) {
	client, err := gorm.Open(postgres.Open(config.ConnectionString), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to create db: %w", err)
	}

	if err = client.AutoMigrate(&TinyURL{}); err != nil {
		return nil, fmt.Errorf("client.AutoMigrate: %w", err)
	}

	return &db{
		cfg:    config,
		client: client,
	}, nil
}

func (db *db) Shutdown() error {
	fmt.Println("Shutting down db server")
	clientDB, err := db.client.DB()
	if err != nil {
		return fmt.Errorf("failed to get internal db: %w", err)
	}
	return clientDB.Close()
}
