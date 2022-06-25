// Recipes API
//
// This is a sample recipes API.
// You can find out more about the API at https://github.com/PacktPublishing/BuildingDistributed-Applications-in-Gin.
//
// Schemes: http
// Host: localhost:8080
// BasePath: /
// Version: 1.0.0
// Contact: Fernando Soto <fsoto82@gmail.com> https://www.curdev.com
//
// Consumes:
// - application/json
//
// Produces:
// - application/json
// swagger:meta
package main

import (
	"context"
	"cookingapp/handlers"
	"cookingapp/models"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var recipesHandler *handlers.RecipesHandler

func init() {
	ctx := context.Background()
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	redisStatus := redisClient.Ping(ctx)
	log.Println("Redis status: ", redisStatus)

	client, err := mongo.Connect(ctx,
		options.Client().ApplyURI(os.Getenv("MONGO_URI")))
	if err = client.Ping(context.TODO(),
		readpref.Primary()); err != nil {
		log.Fatal("Error connecting to database: ", err)
	}
	collection := client.Database(os.Getenv("MONGO_DATABASE")).Collection("recipes")
	log.Println("Connected to MongoDB!!!")
	recipesHandler = handlers.NewRecipesHandler(ctx, collection, redisClient)

	//loadData(ctx, collection)
}

func loadData(ctx context.Context, collection *mongo.Collection) {
	recipes := make([]models.Recipe, 0)
	file, _ := ioutil.ReadFile("recipes.json")
	_ = json.Unmarshal(file, &recipes)
	var listOfRecipes []interface{}
	for _, recipe := range recipes {
		listOfRecipes = append(listOfRecipes, recipe)
	}
	insertManyResult, err := collection.InsertMany(ctx, listOfRecipes)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Inserted recipes: ", len(insertManyResult.InsertedIDs))
}

func AuthMiddleWare(key string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetHeader("X-API-KEY") != key {
			c.AbortWithStatus(401)
		}
		c.Next()
	}
}

func main() {
	router := gin.Default()
	recipes := router.Group("/recipes")
	{
		recipes.GET("", recipesHandler.ListRecipesHandler)
		recipes.GET("/search", recipesHandler.SearchRecipesHandler)
		authorized := recipes.Group("")
		authorized.Use(AuthMiddleWare(os.Getenv("X_API_KEY")))
		{
			authorized.POST("", recipesHandler.NewRecipeHandler)
			authorized.PUT("/:id", recipesHandler.UpdateRecipeHandler)
			authorized.DELETE("/:id", recipesHandler.DeleteRecipeHandler)
		}
	}
	router.Run()
}
