package api

import (
	"context"
	"github.com/labstack/echo"
	"github.com/mongodb/mongo-go-driver/mongo"
	"net/http"
	"time"
)

func LaunchServer() {
	server := echo.New()
	server.GET("/", func(c echo.Context) error {
		client, _ := mongo.NewClient("mongodb://localhost:27017")
		ctx, f := context.WithTimeout(context.Background(), 5*time.Second)
		err := client.Connect(ctx)
		if err != nil {
			f()
			return c.String(http.StatusBadRequest, err.Error())
		}
		server.Logger.Info("Connected to the database")
		f()
		return c.String(http.StatusOK, "Hello, world !")
	})

	server.Logger.Fatal(server.Start(":1323"))
}
