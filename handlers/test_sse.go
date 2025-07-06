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

		// Use a simpler streaming approach
		c.Context().SetBodyStreamWriter(fasthttp.StreamWriter(func(w *bufio.Writer) {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("Test SSE panic recovered: %v", r)
				}
				log.Printf("Test SSE connection ended")
			}()

			// Validate writer is not nil
			if w == nil {
				log.Printf("StreamWriter received nil buffer")
				return
			}

			ticker := time.NewTicker(2 * time.Second)
			defer ticker.Stop()
			
			counter := 0
			maxMessages := 10 // Reduced for testing

			log.Printf("Starting SSE message loop")

			for {
				select {
				case <-ticker.C:
					counter++
					log.Printf("Sending test message %d", counter)
					
					if counter > maxMessages {
						finalMsg := fmt.Sprintf("data: {\"message\": \"Test completed\", \"counter\": %d, \"timestamp\": \"%s\"}\n\n", 
							counter, time.Now().Format(time.RFC3339))
						w.WriteString(finalMsg)
						w.Flush()
						log.Printf("Test completed, sent %d messages", counter)
						return
					}

					testMsg := fmt.Sprintf("data: {\"message\": \"Test message %d\", \"counter\": %d, \"timestamp\": \"%s\"}\n\n", 
						counter, counter, time.Now().Format(time.RFC3339))
					
					w.WriteString(testMsg)
					w.Flush()
					log.Printf("Successfully sent test message %d", counter)

				case <-c.Context().Done():
					log.Printf("Test SSE context cancelled after %d messages", counter)
					return
				}
			}
		}))

		return nil
	}
}