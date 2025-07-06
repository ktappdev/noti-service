package handlers

import (
	"bufio"
	"fmt"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/valyala/fasthttp"
)

// TestSSEHandler creates a simple SSE endpoint for testing
func TestSSEHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Set SSE headers for proper streaming
		c.Set("Content-Type", "text/event-stream")
		c.Set("Cache-Control", "no-cache")
		c.Set("Connection", "keep-alive")

		log.Printf("Test SSE connection started")

		// Send initial message immediately
		initialMsg := fmt.Sprintf("data: {\"message\": \"Test SSE connected\", \"timestamp\": \"%s\"}\n\n", 
			time.Now().Format(time.RFC3339))
		
		if _, err := c.Write([]byte(initialMsg)); err != nil {
			log.Printf("Error writing initial message: %v", err)
			return err
		}

		log.Printf("Initial message sent, starting stream...")

		// Use a much simpler approach - just send a few messages with delays
		c.Context().SetBodyStreamWriter(fasthttp.StreamWriter(func(w *bufio.Writer) {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("Test SSE panic recovered: %v", r)
				}
				log.Printf("Test SSE connection ended")
			}()

			// Validate everything is not nil
			if w == nil {
				log.Printf("StreamWriter received nil buffer")
				return
			}

			log.Printf("Starting simple message sending...")

			// Send 5 messages with 2-second delays
			for i := 1; i <= 5; i++ {
				log.Printf("About to send message %d", i)
				
				// Use time.Sleep instead of ticker to avoid channel issues
				if i > 1 {
					time.Sleep(2 * time.Second)
				}
				
				testMsg := fmt.Sprintf("data: {\"message\": \"Test message %d\", \"timestamp\": \"%s\"}\n\n", 
					i, time.Now().Format(time.RFC3339))
				
				log.Printf("Writing message %d", i)
				w.WriteString(testMsg)
				w.Flush()
				log.Printf("Successfully sent message %d", i)
			}

			log.Printf("All messages sent, ending connection")
		}))

		return nil
	}
}