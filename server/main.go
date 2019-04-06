package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	engine := gin.Default()

	engine.POST("/api/echo", func(c *gin.Context) {
		body, err := ioutil.ReadAll(c.Request.Body)
		if err != nil {
			log.Printf("Error reading body: %v", err)
			c.String(http.StatusBadRequest, fmt.Sprintf("can't read body: %v", err))
			return
		}
		c.String(http.StatusOK, string(body))
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
