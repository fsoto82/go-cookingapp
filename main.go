package main

import (
	"encoding/json"
	"github.com/rs/xid"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

var recipes []Recipe

func init() {
	recipes = make([]Recipe, 0)
	file, _ := ioutil.ReadFile("recipes.json")
	_ = json.Unmarshal(file, &recipes)
}

type Recipe struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Tags         []string  `json:"tags"`
	Ingredients  []string  `json:"ingredients"`
	Instructions []string  `json:"instructions"`
	PublishedAt  time.Time `json:"published_at"`
}

func ListRecipesHandler(c *gin.Context) {
	c.JSON(http.StatusOK, recipes)
}

func SearchRecipesHandler(c *gin.Context) {
	tag := c.Query("tag")
	listOfRecipes := make([]Recipe, 0)

	for _, r := range recipes {
		found := false
		for _, t := range r.Tags {
			if strings.EqualFold(t, tag) {
				found = true
			}
		}
		if found {
			listOfRecipes = append(listOfRecipes, r)
		}
	}
	if len(listOfRecipes) > 0 {
		c.JSON(http.StatusOK, listOfRecipes)
	} else {
		c.JSON(http.StatusNotFound, gin.H{})
	}
}

func NewRecipeHandler(c *gin.Context) {
	var recipe Recipe
	if err := c.ShouldBindJSON(&recipe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	recipe.ID = xid.New().String()
	recipe.PublishedAt = time.Now()
	recipes = append(recipes, recipe)
	c.JSON(http.StatusOK, recipe)
}

func UpdateRecipeHandler(c *gin.Context) {
	id := c.Param("id")
	var recipe Recipe
	if err := c.ShouldBindJSON(&recipe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	for i, r := range recipes {
		if r.ID == id {
			recipes[i] = recipe
			c.JSON(http.StatusOK, recipe)
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "Recipe not found"})
}

func DeleteRecipeHandler(c *gin.Context) {
	id := c.Param("id")
	index := -1
	for i := range recipes {
		if recipes[i].ID == id {
			index = i
		}
	}
	if index == -1 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Recipe not found"})
		return
	}
	recipes = append(recipes[:index], recipes[index+1:]...)
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
