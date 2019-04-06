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

const errorRatePercentage = 50

// this program launches a test server for receiving the notifications from notifier
func main() {
	engine := gin.Default()

	engine.POST("/api/echo", func(c *gin.Context) {
		body, err := ioutil.ReadAll(c.Request.Body)
		if err != nil {
			log.Printf("Error reading body: %v", err)
			c.String(http.StatusBadRequest, fmt.Sprintf("can't read body: %v", err))
			return
		}

		// check if we have to force an error
		src := rand.NewSource(time.Now().UnixNano())
		rnd := rand.New(src)
		value := rnd.Intn(100)
		if value > errorRatePercentage {
			log.Printf("Error forced")
			c.String(http.StatusGatewayTimeout, "")
		} else {
			c.String(http.StatusOK, string(body))
		}
	})

	engine.Run(port())
}

func port() string {
	port := os.Getenv("PORT")

	if len(port) == 0 {
		port = "9090"
	}
	return ":" + port
}
