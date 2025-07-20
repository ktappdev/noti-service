package handlers

import (
	"github.com/gofiber/fiber/v2"
)

// DocsHandler serves comprehensive documentation for the notification service
func DocsHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		docs := `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>ReviewIt Notification Service Documentation</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            line-height: 1.6;
            color: #333;
            max-width: 1200px;
            margin: 0 auto;
            padding: 20px;
            background: #f8f9fa;
        }
        .container {
            background: white;
            padding: 30px;
            border-radius: 8px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
        }
        h1 {
            color: #2c3e50;
            border-bottom: 3px solid #3498db;
            padding-bottom: 10px;
        }
        h2 {
            color: #34495e;
            margin-top: 30px;
            border-left: 4px solid #3498db;
            padding-left: 15px;
        }
        h3 {
            color: #2c3e50;
            margin-top: 25px;
        }
        .endpoint {
            background: #f8f9fa;
            border: 1px solid #dee2e6;
            border-radius: 5px;
            padding: 15px;
            margin: 10px 0;
        }
        .method {
            display: inline-block;
            padding: 4px 8px;
            border-radius: 3px;
            font-weight: bold;
            font-size: 12px;
            margin-right: 10px;
        }
        .get { background: #28a745; color: white; }
        .post { background: #007bff; color: white; }
        .put { background: #ffc107; color: black; }
        .delete { background: #dc3545; color: white; }
        code {
            background: #f1f3f4;
            padding: 2px 6px;
            border-radius: 3px;
            font-family: 'Monaco', 'Consolas', monospace;
        }
        pre {
            background: #2d3748;
            color: #e2e8f0;
            padding: 15px;
            border-radius: 5px;
            overflow-x: auto;
            font-size: 14px;
        }
        .notification-type {
            background: #e3f2fd;
            border-left: 4px solid #2196f3;
            padding: 10px;
            margin: 10px 0;
        }
        .feature-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
            gap: 20px;
            margin: 20px 0;
        }
        .feature-card {
            background: #f8f9fa;
            border: 1px solid #dee2e6;
            border-radius: 8px;
            padding: 20px;
        }
        .status-badge {
            display: inline-block;
            padding: 4px 12px;
            border-radius: 12px;
            font-size: 12px;
            font-weight: bold;
        }
        .status-live { background: #d4edda; color: #155724; }
        .status-beta { background: #fff3cd; color: #856404; }
        .toc {
            background: #f8f9fa;
            border: 1px solid #dee2e6;
            border-radius: 5px;
            padding: 20px;
            margin: 20px 0;
        }
        .toc ul {
            list-style: none;
            padding-left: 0;
        }
        .toc li {
            margin: 8px 0;
        }
        .toc a {
            color: #007bff;
            text-decoration: none;
        }
        .toc a:hover {
            text-decoration: underline;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>📡 ReviewIt Notification Service Documentation</h1>
        
        <div class="toc">
            <h3>📋 Table of Contents</h3>
            <ul>
                <li><a href="#overview">🎯 Service Overview</a></li>
                <li><a href="#architecture">🏗️ Architecture</a></li>
                <li><a href="#notification-types">📨 Notification Types</a></li>
                <li><a href="#rest-api">🔗 REST API Endpoints</a></li>
                <li><a href="#sse">📡 Server-Sent Events (SSE)</a></li>
                <li><a href="#health">💚 Health & Monitoring</a></li>
                <li><a href="#integration">🔧 Frontend Integration</a></li>
                <li><a href="#examples">💡 Code Examples</a></li>
            </ul>
        </div>

        <section id="overview">
            <h2>🎯 Service Overview</h2>
            <p>The ReviewIt Notification Service provides real-time notifications for user interactions across the ReviewIt platform. It supports multiple notification types with both REST API and real-time SSE delivery.</p>
            
            <div class="feature-grid">
                <div class="feature-card">
                    <h3>🚀 Real-time Delivery</h3>
                    <p>Server-Sent Events (SSE) for instant notification delivery to connected clients.</p>
                </div>
                <div class="feature-card">
                    <h3>📱 Multiple Types</h3>
                    <p>Support for likes, comments, replies, reviews, and system notifications.</p>
                </div>
                <div class="feature-card">
                    <h3>🔄 Unified Schema</h3>
                    <p>Consistent notification format with target URLs for easy frontend integration.</p>
                </div>
                <div class="feature-card">
                    <h3>📊 Scalable</h3>
                    <p>Efficient database design with optimized indexes for high-performance queries.</p>
                </div>
            </div>
        </section>

        <section id="architecture">
            <h2>🏗️ Architecture</h2>
            <h3>Components</h3>
            <ul>
                <li><strong>REST API</strong>: HTTP endpoints for notification CRUD operations</li>
                <li><strong>SSE Hub</strong>: Real-time message broadcasting to connected clients</li>
                <li><strong>Database Layer</strong>: PostgreSQL with optimized schemas and indexes</li>
                <li><strong>ReviewIt Integration</strong>: Lookups for user and content information</li>
            </ul>
            
            <h3>Data Flow</h3>
            <pre>
Frontend → REST API → Database → SSE Hub → Connected Clients
    ↓           ↓         ↓          ↓
  Create    Validate   Store    Broadcast
            </pre>
        </section>

        <section id="notification-types">
            <h2>📨 Notification Types</h2>
            <p>The service supports a unified schema with these notification types:</p>
            
            <div class="notification-type">
                <h3>👍 Like Notifications</h3>
                <ul>
                    <li><code>like_review</code> - Someone liked your review</li>
                    <li><code>like_comment</code> - Someone liked your comment</li>
                </ul>
            </div>
            
            <div class="notification-type">
                <h3>💬 Comment & Reply Notifications</h3>
                <ul>
                    <li><code>reply_review</code> - Someone commented on your review</li>
                    <li><code>reply_comment</code> - Someone replied to your comment</li>
                </ul>
            </div>
            
            <div class="notification-type">
                <h3>🏪 Owner Notifications</h3>
                <ul>
                    <li><code>owner_review</code> - New review for your product</li>
                </ul>
            </div>
            
            <div class="notification-type">
                <h3>🔔 System Notifications</h3>
                <ul>
                    <li><code>system</code> - Admin announcements and updates</li>
                </ul>
            </div>

            <h3>Unified Schema Fields</h3>
            <pre>
{
  "id": "unique_string",
  "created_at": "ISO-8601_timestamp",
  "read": boolean,
  "notification_type": "like_review|like_comment|reply_review|reply_comment|owner_review|system",
  "from_id": "user_id",
  "from_name": "User Name",
  "target_type": "review|comment",
  "review_id": "always_present",
  "comment_id": "present_when_target_type_is_comment",
  "target_url": "/review/{review_id}[?cid={comment_id}]"
}
            </pre>
        </section>

        <section id="rest-api">
            <h2>🔗 REST API Endpoints</h2>
            
            <h3>User Management</h3>
            <div class="endpoint">
                <span class="method post">POST</span>
                <code>/users</code>
                <p>Create or update a user in the notification service</p>
            </div>
            
            <h3>Notification Creation</h3>
            <div class="endpoint">
                <span class="method post">POST</span>
                <code>/notifications/like</code>
                <p>Create a like notification (review or comment)</p>
            </div>
            
            <div class="endpoint">
                <span class="method post">POST</span>
                <code>/notifications/comment</code>
                <p>Create a comment notification (reply to review)</p>
            </div>
            
            <div class="endpoint">
                <span class="method post">POST</span>
                <code>/notifications/reply</code>
                <p>Create a reply notification (reply to comment)</p>
            </div>
            
            <div class="endpoint">
                <span class="method post">POST</span>
                <code>/notifications/product-owner</code>
                <p>Create a product owner notification (new review)</p>
            </div>
            
            <div class="endpoint">
                <span class="method post">POST</span>
                <code>/notifications/system</code>
                <p>Create a system notification (admin announcements)</p>
            </div>
            
            <h3>Notification Retrieval</h3>
            <div class="endpoint">
                <span class="method get">GET</span>
                <code>/notifications?user_id={id}</code>
                <p>Get all notifications for a user</p>
            </div>
            
            <div class="endpoint">
                <span class="method get">GET</span>
                <code>/notifications/unread?user_id={id}</code>
                <p>Get unread notifications for a user</p>
            </div>
            
            <div class="endpoint">
                <span class="method get">GET</span>
                <code>/notifications/latest?user_id={id}</code>
                <p>Get latest notification for a user</p>
            </div>
            
            <h3>Notification Management</h3>
            <div class="endpoint">
                <span class="method put">PUT</span>
                <code>/notifications/{id}/read?user_id={uid}&type={type}</code>
                <p>Mark a specific notification as read</p>
            </div>
            
            <div class="endpoint">
                <span class="method put">PUT</span>
                <code>/notifications/read-all?user_id={id}[&type={type}]</code>
                <p>Mark all notifications as read (optionally filter by type)</p>
            </div>
            
            <div class="endpoint">
                <span class="method delete">DELETE</span>
                <code>/notifications?user_id={id}</code>
                <p>Delete all read notifications for a user</p>
            </div>
        </section>

        <section id="sse">
            <h2>📡 Server-Sent Events (SSE)</h2>
            
            <div class="endpoint">
                <span class="method get">GET</span>
                <code>/notifications/stream?user_id={id}</code>
                <p>Establish SSE connection for real-time notifications</p>
            </div>
            
            <h3>SSE Message Format</h3>
            <pre>
data: {
  "user_id": "user_123",
  "type": "like|user|owner|system",
  "event": "new_notification|notification_read|notifications_bulk_read",
  "notification": { /* notification object */ }
}
            </pre>
            
            <h3>Event Types</h3>
            <ul>
                <li><code>new_notification</code> - New notification received</li>
                <li><code>notification_read</code> - Single notification marked as read</li>
                <li><code>notifications_bulk_read</code> - Multiple notifications marked as read</li>
            </ul>
            
            <h3>Connection Management</h3>
            <ul>
                <li>Automatic reconnection on connection loss</li>
                <li>Multiple connections per user supported</li>
                <li>Graceful cleanup on disconnect</li>
                <li>Connection count monitoring available</li>
            </ul>
        </section>

        <section id="health">
            <h2>💚 Health & Monitoring</h2>
            
            <div class="endpoint">
                <span class="method get">GET</span>
                <code>/health</code>
                <p>Service health check with database and SSE status</p>
            </div>
            
            <h3>Health Response</h3>
            <pre>
{
  "status": "healthy|unhealthy",
  "database": "connected|disconnected", 
  "sse_connections": 42
}
            </pre>
            
            <h3>Monitoring Features</h3>
            <ul>
                <li>Database connection health</li>
                <li>Active SSE connection count</li>
                <li>Enhanced error logging with context</li>
                <li>Performance-optimized database indexes</li>
            </ul>
        </section>

        <section id="integration">
            <h2>🔧 Frontend Integration</h2>
            
            <h3>SSE Connection Setup</h3>
            <pre>
const eventSource = new EventSource(
  'https://notifications.reviewit.gy/notifications/stream?user_id=user_123'
);

eventSource.onmessage = (event) => {
  const data = JSON.parse(event.data);
  
  if (data.event === 'new_notification') {
    handleNewNotification(data.notification);
  }
};
            </pre>
            
            <h3>Creating Notifications</h3>
            <pre>
// Like notification
await fetch('/notifications/like', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    target_type: 'review',
    target_id: 'review_123',
    from_id: 'user_456',
    from_name: 'John Doe'
  })
});
            </pre>
            
            <h3>Navigation with target_url</h3>
            <pre>
// Direct navigation using target_url
function handleNotificationClick(notification) {
  if (notification.target_url) {
    window.location.href = notification.target_url;
  }
}
            </pre>
        </section>

        <section id="examples">
            <h2>💡 Code Examples</h2>
            
            <h3>Complete React Integration</h3>
            <pre>
import { useEffect, useState } from 'react';

function useNotifications(userId) {
  const [notifications, setNotifications] = useState([]);
  const [unreadCount, setUnreadCount] = useState(0);

  useEffect(() => {
    // Fetch existing notifications
    fetch(` + "`/notifications/unread?user_id=${userId}`" + `)
      .then(res => res.json())
      .then(data => {
        const allNotifications = [
          ...data.user_notifications,
          ...data.owner_notifications,
          ...data.like_notifications,
          ...data.system_notifications
        ];
        setNotifications(allNotifications);
        setUnreadCount(allNotifications.length);
      });

    // Setup SSE
    const eventSource = new EventSource(
      ` + "`/notifications/stream?user_id=${userId}`" + `
    );

    eventSource.onmessage = (event) => {
      const data = JSON.parse(event.data);
      
      if (data.event === 'new_notification') {
        setNotifications(prev => [data.notification, ...prev]);
        setUnreadCount(prev => prev + 1);
      }
    };

    return () => eventSource.close();
  }, [userId]);

  const markAsRead = async (notificationId, type) => {
    await fetch(` + "`/notifications/${notificationId}/read?user_id=${userId}&type=${type}`" + `, {
      method: 'PUT'
    });
    setUnreadCount(prev => Math.max(0, prev - 1));
  };

  return { notifications, unreadCount, markAsRead };
}
            </pre>
        </section>

        <div style="margin-top: 40px; padding-top: 20px; border-top: 1px solid #dee2e6; text-align: center; color: #6c757d;">
            <p>📚 For more detailed API specifications, see the individual documentation files in the repository.</p>
            <p><span class="status-badge status-live">LIVE</span> Service Status: Operational</p>
        </div>
    </div>
</body>
</html>
`
		c.Set("Content-Type", "text/html")
		return c.SendString(docs)
	}
}