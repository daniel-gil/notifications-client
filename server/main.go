package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

const errorRatePercentage = 10

var rnd *rand.Rand

// this program launches a test server for receiving the notifications from notifier
func main() {
	engine := gin.Default()

	engine.POST("/api/notifications", func(c *gin.Context) {
		body, err := ioutil.ReadAll(c.Request.Body)
		if err != nil {
			log.Printf("Error reading body: %v", err)
			c.String(http.StatusBadRequest, fmt.Sprintf("can't read body: %v", err))
			return
		}

		// check if we have to force an error
		value := getRandomValue()
		log.Printf("RANDOM: %v\n", value)
		if value > errorRatePercentage {
			c.String(http.StatusOK, string(body))
		} else {
			log.Printf("Error forced")
			c.String(http.StatusBadRequest, "")
		}
	})

	engine.Run(port())
}

func init() {
	src := rand.NewSource(time.Now().UnixNano())
	rnd = rand.New(src)
}

func getRandomValue() int {
	return rnd.Intn(100)
}

func port() string {
	port := os.Getenv("PORT")

	if len(port) == 0 {
		port = "9090"
	}
	return ":" + port
}
