package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

var rnd *rand.Rand

// this program launches a test server for receiving the notifications from notifier
func main() {
	errorRatePercentage := flag.Int("error", 0, "Error rate percentage to simulate failures")
	flag.Parse()
	log.Printf("Server configuration: errorRatePercentage=%d%%", *errorRatePercentage)

	engine := gin.Default()

	engine.POST("/api/notifications", func(c *gin.Context) {
		body, err := ioutil.ReadAll(c.Request.Body)
		if err != nil {
			log.Printf("Error reading body: %v", err)
			c.String(http.StatusBadRequest, fmt.Sprintf("can't read body: %v", err))
			return
		}

		// check if we have to force an error
		if *errorRatePercentage == 0 {
			c.String(http.StatusOK, string(body))
		} else {
			value := getRandomValue()
			if value > *errorRatePercentage {
				c.String(http.StatusOK, string(body))
			} else {
				c.String(http.StatusBadRequest, "")
			}
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
