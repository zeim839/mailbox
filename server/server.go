package main

import (
	"context"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/zeim839/mailbox/config"
	"github.com/zeim839/mailbox/core"
	"github.com/zeim839/mailbox/data"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"log"
	"net/http"
	"time"
)

func main() {
	config, err := config.LoadConfig(".")
	if err != nil {
		log.Fatal(err)
	}

	// Connect to MongoDB.
	mongoconn := options.Client().ApplyURI(config.MongoURI)
	mongoclient, err := mongo.Connect(context.TODO(), mongoconn)
	if err != nil {
		log.Fatal(err)
	}
	if err := mongoclient.Ping(context.TODO(), readpref.Primary()); err != nil {
		log.Fatal(err)
	}
	defer mongoclient.Disconnect(context.TODO())
	log.Print("MongoDB successfully connected...")

	// Create database controller.
	coll := mongoclient.Database("MAILBOX").Collection("entries")
	mongo, _ := data.NewMongo(coll)

	// Set up Gin.
	gin.SetMode(config.GinMode)
	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"POST, PUT, GET, DELETE"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	if config.CaptchaSecret != "" {
		log.Print("Captcha successfully configured")
		r.POST("/mailbox/submit", core.CreateWithCaptcha(mongo, config.CaptchaSecret))
	} else {
		log.Print("Captcha not configured")
		r.POST("/mailbox/submit", core.Create(mongo))
	}

	if config.Username != "" && config.Password != "" {
		log.Print("Basic auth successfully configured")
		r.GET("/mailbox/entry/:id", core.BasicAuthMw(config.Username,
			config.Password), core.Read(mongo))
		r.DELETE("/mailbox/entry/:id", core.BasicAuthMw(config.Username,
			config.Password), core.Delete(mongo))
		r.GET("/mailbox/entries/", core.BasicAuthMw(config.Username,
			config.Password), core.ReadAll(mongo))
	} else {
		log.Print("Basic auth not configured")
		r.GET("/mailbox/entry/:id", core.Read(mongo))
		r.DELETE("/mailbox/entry/:id", core.Delete(mongo))
		r.GET("/mailbox/entries/", core.ReadAll(mongo))
	}

	r.GET("/status", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	r.Run("0.0.0.0:" + config.Port)
}
