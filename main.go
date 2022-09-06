package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/patrickmn/go-cache"
)

func getWeatherFromAPI(API_KEY string) (string, bool) {
	url := "https://api.openweathermap.org/data/2.5/weather?q=Stuttgart,de&units=metric&APPID=" + API_KEY
	response, error := http.Get(url)
	if error != nil {
		log.Fatal(error)
		return "", true
	}
	defer response.Body.Close()
	text, _ := io.ReadAll(response.Body)
	return string(text), false
}

func main() {
	// Loading .env file
	godotenv.Load()

	// loading parameters from .env or environment variables
	API_KEY := os.Getenv("API_KEY")
	HOST := os.Getenv("HOST")
	PORT := os.Getenv("PORT")

	// setting up cache
	CACHE_DURATION := 5 * time.Minute
	memcache := cache.New(CACHE_DURATION, CACHE_DURATION*2)

	// setting up Gin router
	router := gin.Default()
	router.Static("/static", "./static") // static files
	router.LoadHTMLGlob("templates/*")   // templates

	router.SetTrustedProxies(nil) // stops warning - no proxies needed -
	//see https://pkg.go.dev/github.com/gin-gonic/gin#readme-don-t-trust-all-proxies

	// Route GET /
	router.GET("/", func(c *gin.Context) {
		var weatherStr string
		weatherCache, found := memcache.Get("weather") // loading values from cache
		if !found {
			error := false
			weatherStr, error = getWeatherFromAPI(API_KEY) // loading values from external API
			if error {
				fmt.Println(weatherStr)
				log.Fatal("Error in API call")
				return
			}
			fmt.Println("Cache created")
			memcache.Set("weather", weatherStr, cache.DefaultExpiration) // saving values to cache
		} else {
			fmt.Println("Loaded from Cache")
			weatherStr = weatherCache.(string)
		}

		var weather map[string]interface{}
		json.Unmarshal([]byte(weatherStr), &weather) // string to json

		/*
					   condition: JSON["weather"][0]["description"]
					   temperature: weather["main"]["temp"]

			       Cheat sheet:
			       JSON-Array:  ja := json["key"].([]interface{})
			       JSON-Object: jo := json["key"].(map[string]interface{})
		*/
		condition := weather["weather"].([]interface{})[0].(map[string]interface{})["description"]
		temperature := fmt.Sprintf("%.1f", weather["main"].(map[string]interface{})["temp"])

		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"condition":   condition,
			"temperature": temperature,
		})
	})

	// Route GET ping (example API route)
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	// starting webapp
	router.Run(HOST + ":" + PORT) // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
