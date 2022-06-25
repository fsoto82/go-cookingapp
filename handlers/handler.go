package handlers

import (
	"context"
	"cookingapp/models"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"net/http"
	"time"
)

type RecipesHandler struct {
	ctx         context.Context
	collection  *mongo.Collection
	redisClient *redis.Client
}

func NewRecipesHandler(ctx context.Context, collection *mongo.Collection, redisClient *redis.Client) *RecipesHandler {
	return &RecipesHandler{
		ctx:         ctx,
		collection:  collection,
		redisClient: redisClient,
	}
}

// swagger:operation GET /recipes recipes listRecipes
// Returns list of recipes
// ---
// produces:
// - application/json
// responses:
//   '200':
//     description: Successful operation
func (handler *RecipesHandler) ListRecipesHandler(c *gin.Context) {
	recipes := make([]models.Recipe, 0)
	// Try to read from Redis
	if val, err := handler.redisClient.Get(handler.ctx, "recipes").Result(); err == redis.Nil {
		log.Println("Request to MongoDB")
		cur, err := handler.collection.Find(handler.ctx, bson.M{})
		if err != nil {
			log.Println("Error searching recipes ", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error searching recipes"})
			return
		}
		defer cur.Close(handler.ctx)

		for cur.Next(handler.ctx) {
			var recipe models.Recipe
			_ = cur.Decode(&recipe)
			recipes = append(recipes, recipe)
		}
		// Insert in Redis
		data, _ := json.Marshal(recipes)
		handler.redisClient.Set(handler.ctx, "recipes", string(data), 0)
	} else if err != nil {
		log.Println("Error searching recipes ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error searching recipes"})
		return
	} else {
		log.Println("Request to Redis")
		json.Unmarshal([]byte(val), &recipes)
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
func (handler *RecipesHandler) SearchRecipesHandler(c *gin.Context) {
	tag := c.Query("tag")
	cur, err := handler.collection.Find(handler.ctx, bson.M{"tags": tag})
	if err != nil {
		log.Println("Error searching recipes ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error searching recipes"})
	}
	defer cur.Close(handler.ctx)
	recipes := make([]models.Recipe, 0)
	for cur.Next(handler.ctx) {
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
func (handler *RecipesHandler) NewRecipeHandler(c *gin.Context) {
	var recipe models.Recipe
	if err := c.ShouldBindJSON(&recipe); err != nil {
		log.Println("Error reading recipe: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error reading recipe"})
		return
	}
	recipe.ID = primitive.NewObjectID()
	recipe.PublishedAt = time.Now()
	_, err := handler.collection.InsertOne(handler.ctx, recipe)
	if err != nil {
		log.Println("Error inserting a new recipe: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error inserting a new recipe"})
		return
	}

	log.Println("Remove data from Redis")
	handler.redisClient.Del(handler.ctx, "recipes")

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
func (handler *RecipesHandler) UpdateRecipeHandler(c *gin.Context) {
	id := c.Param("id")
	var recipe models.Recipe
	if err := c.ShouldBindJSON(&recipe); err != nil {
		log.Println("Error reading recipe: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error reading recipe"})
		return
	}
	objectId, _ := primitive.ObjectIDFromHex(id)
	updateResult, err := handler.collection.UpdateOne(handler.ctx,
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

	log.Println("Remove data from Redis")
	handler.redisClient.Del(handler.ctx, "recipes")

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
func (handler *RecipesHandler) DeleteRecipeHandler(c *gin.Context) {
	id := c.Param("id")
	objectId, _ := primitive.ObjectIDFromHex(id)
	deleteResult, err := handler.collection.DeleteOne(handler.ctx, bson.M{"_id": objectId})
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
