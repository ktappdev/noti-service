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
		c.Set("Transfer-Encoding", "chunked")
		c.Set("X-Accel-Buffering", "no")

		log.Printf("Test SSE connection started")

		c.Context().SetBodyStreamWriter(fasthttp.StreamWriter(func(w *bufio.Writer) {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("Test SSE panic recovered: %v", r)
				}
				log.Printf("Test SSE connection ended")
			}()

			// Send initial connection message
			initialMsg := fmt.Sprintf("data: {\"message\": \"Test SSE connected\", \"timestamp\": \"%s\"}\n\n", 
				time.Now().Format(time.RFC3339))
			
			if _, err := w.WriteString(initialMsg); err != nil {
				log.Printf("Error writing initial test message: %v", err)
				return
			}
			if err := w.Flush(); err != nil {
				log.Printf("Error flushing initial test message: %v", err)
				return
			}

			// Send a message every 2 seconds for 30 seconds
			ticker := time.NewTicker(2 * time.Second)
			defer ticker.Stop()
			
			counter := 0
			maxMessages := 15 // 30 seconds worth

			for {
				select {
				case <-ticker.C:
					counter++
					if counter > maxMessages {
						// Send final message and close
						finalMsg := fmt.Sprintf("data: {\"message\": \"Test completed\", \"counter\": %d, \"timestamp\": \"%s\"}\n\n", 
							counter, time.Now().Format(time.RFC3339))
						w.WriteString(finalMsg)
						w.Flush()
						return
					}

					// Send regular test message
					testMsg := fmt.Sprintf("data: {\"message\": \"Test message %d\", \"counter\": %d, \"timestamp\": \"%s\"}\n\n", 
						counter, counter, time.Now().Format(time.RFC3339))
					
					if _, err := w.WriteString(testMsg); err != nil {
						log.Printf("Error writing test message %d: %v", counter, err)
						return
					}
					if err := w.Flush(); err != nil {
						log.Printf("Error flushing test message %d: %v", counter, err)
						return
					}

				case <-c.Context().Done():
					log.Printf("Test SSE context cancelled")
					return
				}
			}
		}))

		return nil
	}
}