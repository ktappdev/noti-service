package handlers

import (
	"github.com/gofiber/fiber/v2"
)

// SSEHelpHandler serves documentation about the SSE implementation
func SSEHelpHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set("Content-Type", "text/plain; charset=utf-8")
		
		helpText := `
SERVER-SENT EVENTS (SSE) DOCUMENTATION
=====================================

This notification service implements Server-Sent Events (SSE) using only Go's standard library 
and the Fiber web framework. No additional SSE libraries were required.

WHAT IS SERVER-SENT EVENTS (SSE)?
=================================

SSE is a web standard that allows a server to push data to a web page in real-time. 
Unlike WebSockets, SSE is unidirectional (server → client only) and uses regular HTTP connections.

Key characteristics:
- Uses standard HTTP/HTTPS protocol
- Automatic reconnection by browsers
- Simple text-based format
- Built-in browser support via EventSource API
- Lightweight compared to WebSockets

HOW SSE WORKS TECHNICALLY
========================

1. CLIENT INITIATES CONNECTION:
   - Browser creates EventSource object
   - Sends GET request with Accept: text/event-stream header
   - Server responds with Content-Type: text/event-stream

2. SERVER KEEPS CONNECTION OPEN:
   - Server doesn't close the HTTP response
   - Sends data in specific SSE format
   - Each message ends with double newline (\n\n)

3. MESSAGE FORMAT:
   data: {"message": "Hello World"}\n\n
   
   Optional fields:
   id: 123\n
   event: custom-event\n
   data: {"message": "Hello"}\n\n

4. BROWSER RECEIVES DATA:
   - EventSource fires 'message' events
   - Automatic JSON parsing if needed
   - Built-in reconnection on connection loss

OUR SSE IMPLEMENTATION DETAILS
==============================

ENDPOINT: GET /notifications/stream?user_id=USER_ID

REQUIRED HEADERS (set automatically):
- Content-Type: text/event-stream
- Cache-Control: no-cache  
- Connection: keep-alive

ARCHITECTURE COMPONENTS:

1. SSE HUB (sse/hub.go):
   - Central manager for all SSE connections
   - Maintains map of userID -> []*SSEClient
   - Handles client registration/unregistration
   - Broadcasts messages to specific users
   - Runs in separate goroutine

2. SSE CLIENT STRUCTURE:
   type SSEClient struct {
       UserID  string        // User identifier
       Channel chan []byte   // Message queue (buffered, size 10)
       Done    chan bool     // Cleanup signal
       ID      string        // Unique client ID (userID_timestamp)
   }

3. SSE HANDLER (handlers/sse.go):
   - Validates user_id parameter
   - Creates unique client instance
   - Registers client with hub
   - Sends initial connection message
   - Sends existing unread notifications
   - Manages connection lifecycle
   - Uses Fiber's StreamWriter for efficient streaming

4. MESSAGE BROADCASTING:
   - Integrated into notification creation endpoints
   - Integrated into notification read status updates
   - Uses hub.BroadcastToUser() method
   - Non-blocking sends (drops if client channel full)

MESSAGE TYPES SENT:
==================

1. CONNECTION CONFIRMATION:
{
  "user_id": "user123",
  "type": "system",
  "event": "connected",
  "notification": {
    "message": "Connected to notification stream",
    "time": "2024-01-01T12:00:00Z"
  }
}

2. EXISTING NOTIFICATIONS (on connect):
{
  "user_id": "user123", 
  "type": "user" | "owner",
  "event": "existing_notification",
  "notification": { /* full notification object */ }
}

3. NEW NOTIFICATIONS (real-time):
{
  "user_id": "user123",
  "type": "user" | "owner", 
  "event": "new_notification",
  "notification": { /* full notification object */ }
}

4. READ STATUS UPDATES:
{
  "user_id": "user123",
  "type": "user" | "owner",
  "event": "notification_read", 
  "notification": {
    "notification_id": "notif123",
    "type": "user",
    "read": true,
    "timestamp": "2024-01-01T12:00:00Z"
  }
}

FRONTEND USAGE EXAMPLE:
======================

// Connect to SSE stream
const eventSource = new EventSource('http://localhost:3001/notifications/stream?user_id=user123');

// Handle all messages
eventSource.onmessage = function(event) {
    const data = JSON.parse(event.data);
    console.log('Received:', data);
    
    switch(data.event) {
        case 'connected':
            console.log('Connected to notification stream');
            break;
        case 'new_notification':
            showNewNotification(data.notification);
            break;
        case 'notification_read':
            updateNotificationStatus(data.notification);
            break;
        case 'existing_notification':
            addExistingNotification(data.notification);
            break;
    }
};

// Handle connection errors
eventSource.onerror = function(event) {
    console.error('SSE connection error:', event);
    // Browser will automatically attempt to reconnect
};

// Close connection when done
// eventSource.close();

IMPLEMENTATION BENEFITS:
=======================

1. NO EXTERNAL DEPENDENCIES:
   - Uses only Go standard library + Fiber
   - No additional SSE libraries needed
   - Lightweight and fast

2. SCALABLE ARCHITECTURE:
   - Hub pattern allows multiple clients per user
   - Non-blocking message sending
   - Efficient memory usage with buffered channels

3. ROBUST ERROR HANDLING:
   - Client cleanup on disconnect
   - Graceful handling of full channels
   - Automatic browser reconnection

4. PRODUCTION READY:
   - CORS configured for production domains
   - Proper HTTP headers
   - Connection lifecycle management

TESTING THE IMPLEMENTATION:
==========================

1. START SERVER:
   go run main.go

2. TEST SSE CONNECTION:
   curl -N -H "Accept: text/event-stream" \
   "http://localhost:3001/notifications/stream?user_id=test123"

3. CREATE NOTIFICATION (in another terminal):
   curl -X POST http://localhost:3001/notifications/product-owner \
   -H "Content-Type: application/json" \
   -d '{"owner_id":"test123","product_name":"Test Product","from_name":"Test User"}'

4. OBSERVE REAL-TIME MESSAGE:
   You should see the notification appear in the SSE stream immediately.

BROWSER COMPATIBILITY:
=====================

SSE is supported in all modern browsers:
- Chrome/Edge: Full support
- Firefox: Full support  
- Safari: Full support
- IE: Not supported (use polyfill if needed)

For IE support, consider using EventSource polyfill or fallback to polling.

PERFORMANCE CONSIDERATIONS:
==========================

1. CONNECTION LIMITS:
   - Browsers limit concurrent SSE connections per domain (usually 6)
   - Consider connection pooling for multiple tabs

2. MEMORY USAGE:
   - Each client uses ~1KB memory for channels
   - Hub automatically cleans up disconnected clients

3. NETWORK EFFICIENCY:
   - SSE reuses HTTP connections
   - More efficient than polling for real-time updates
   - Less overhead than WebSockets for one-way communication

DEBUGGING TIPS:
==============

1. CHECK SERVER LOGS:
   - Client registration/unregistration messages
   - Broadcasting success/failure logs

2. BROWSER DEV TOOLS:
   - Network tab shows SSE connection
   - Console shows EventSource events
   - Application tab shows EventSource status

3. CURL TESTING:
   - Use curl -N for streaming responses
   - Test connection before frontend integration

COMPARISON WITH ALTERNATIVES:
============================

SSE vs POLLING:
+ Real-time updates (no delay)
+ More efficient (persistent connection)
+ Automatic reconnection
- Keeps connection open (resource usage)

SSE vs WebSockets:
+ Simpler implementation
+ Automatic reconnection
+ Works through proxies/firewalls better
+ HTTP/2 compatible
- One-way communication only
- Less efficient for high-frequency bidirectional data

This implementation provides a robust, scalable SSE solution using minimal dependencies
and following Go best practices for concurrent programming.
`
		
		return c.SendString(helpText)
	}
}