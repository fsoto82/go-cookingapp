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
	"cookingapp/models"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

var recipes []models.Recipe

var (
	ctx        context.Context
	client     *mongo.Client
	collection *mongo.Collection
	err        error
)

func init() {
	ctx = context.Background()
	client, err = mongo.Connect(ctx,
		options.Client().ApplyURI(os.Getenv("MONGO_URI")))
	if err = client.Ping(context.TODO(),
		readpref.Primary()); err != nil {
		log.Fatal(err)
	}
	collection = client.Database(os.Getenv("MONGO_DATABASE")).Collection("recipes")
	log.Println("Connected to MongoDB!!!")

	//loadData()
}

func loadData() {
	recipes = make([]models.Recipe, 0)
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

// swagger:operation GET /recipes recipes listRecipes
// Returns list of recipes
// ---
// produces:
// - application/json
// responses:
//   '200':
//     description: Successful operation
func ListRecipesHandler(c *gin.Context) {
	cur, err := collection.Find(ctx, bson.M{})
	if err != nil {
		log.Println("Error searching recipes ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error searching recipes"})
	}
	defer cur.Close(ctx)
	recipes := make([]models.Recipe, 0)
	for cur.Next(ctx) {
		var recipe models.Recipe
		_ = cur.Decode(&recipe)
		recipes = append(recipes, recipe)
	}
	c.JSON(http.StatusOK, recipes)
}

// swagger:operation GET /recipes/search recipes searchRecipes
// Returns list of recipes searching by tag
// ---
// parameters:
// - name: tag
//   in: query
//   description: tag to search
//   required: true
//   type: string
// produces:
// - application/json
// responses:
//   '200':
//     description: Successful operation
//   '404':
//     description: Not found recipes
func SearchRecipesHandler(c *gin.Context) {
	tag := c.Query("tag")
	cur, err := collection.Find(ctx, bson.M{"tags": tag})
	if err != nil {
		log.Println("Error searching recipes ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error searching recipes"})
	}
	defer cur.Close(ctx)
	recipes := make([]models.Recipe, 0)
	for cur.Next(ctx) {
		var recipe models.Recipe
		_ = cur.Decode(&recipe)
		recipes = append(recipes, recipe)
	}
	c.JSON(http.StatusOK, recipes)
}

// swagger:operation POST /recipes recipes createRecipe
// Create a new recipe
// ---
// produces:
// - application/json
// responses:
//   '200':
//     description: Successful operation
//   '400':
//     description: Invalid input
func NewRecipeHandler(c *gin.Context) {
	var recipe models.Recipe
	if err := c.ShouldBindJSON(&recipe); err != nil {
		log.Println("Error reading recipe: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error reading recipe"})
		return
	}
	recipe.ID = primitive.NewObjectID()
	recipe.PublishedAt = time.Now()
	_, err := collection.InsertOne(ctx, recipe)
	if err != nil {
		log.Println("Error inserting a new recipe: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error inserting a new recipe"})
		return
	}
	c.JSON(http.StatusOK, recipe)
}

// swagger:operation PUT /recipes/{id} recipes updateRecipe
// Update an existing recipe
// ---
// parameters:
// - name: id
//   in: path
//   description: ID of the recipe
//   required: true
//   type: string
// produces:
// - application/json
// responses:
//   '200':
//     description: Successful operation
//   '400':
//     description: Invalid input
//   '404':
//     description: Invalid recipe ID
func UpdateRecipeHandler(c *gin.Context) {
	id := c.Param("id")
	var recipe models.Recipe
	if err := c.ShouldBindJSON(&recipe); err != nil {
		log.Println("Error reading recipe: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error reading recipe"})
		return
	}
	objectId, _ := primitive.ObjectIDFromHex(id)
	updateResult, err := collection.UpdateOne(ctx,
		bson.M{"_id": objectId},
		bson.D{{
			"$set", bson.D{
				{"name", recipe.Name},
				{"instructions", recipe.Instructions},
				{"ingredients", recipe.Ingredients},
				{"tags", recipe.Tags},
			},
		}})
	if err != nil {
		log.Println("Error updating recipe: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating a existing recipe"})
		return
	}
	if updateResult.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Recipe not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Recipe has been updated"})
}

// swagger:operation DELETE /recipes/{id} recipes deleteRecipe
// Delete an existing recipe
// ---
// parameters:
// - name: id
//   in: path
//   description: ID of the recipe
//   required: true
//   type: string
// produces:
// - application/json
// responses:
//   '200':
//     description: Successful operation
//   '404':
//     description: Invalid recipe ID
func DeleteRecipeHandler(c *gin.Context) {
	id := c.Param("id")
	objectId, _ := primitive.ObjectIDFromHex(id)
	deleteResult, err := collection.DeleteOne(ctx, bson.M{"_id": objectId})
	if err != nil {
		log.Println("Error deleting recipe: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting a existing recipe"})
		return
	}
	if deleteResult.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Recipe not found"})
		return
	}
	c.JSON(http.StatusAccepted, gin.H{"message": "Recipe deleted"})
}

func main() {
	router := gin.Default()
	router.GET("/recipes", ListRecipesHandler)
	router.GET("/recipes/search", SearchRecipesHandler)
	router.POST("/recipes", NewRecipeHandler)
	router.PUT("/recipes/:id", UpdateRecipeHandler)
	router.DELETE("/recipes/:id", DeleteRecipeHandler)
	router.Run()
}
