# 🚀 Quick Frontend Setup Guide

## 1. Minimal HTML Example (5 minutes)

Create an `index.html` file and copy this code:

```html
<!DOCTYPE html>
<html>
<head>
    <title>Notification Service Test</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .notification { 
            background: #f0f8ff; 
            border: 1px solid #0066cc; 
            padding: 10px; 
            margin: 10px 0; 
            border-radius: 5px; 
        }
        .status { 
            padding: 5px 10px; 
            border-radius: 3px; 
            margin: 5px 0; 
        }
        .connected { background: #d4edda; color: #155724; }
        .error { background: #f8d7da; color: #721c24; }
        button { padding: 10px 15px; margin: 5px; cursor: pointer; }
    </style>
</head>
<body>
    <h1>🔔 Notification Service Test</h1>
    
    <div>
        <label>User ID: <input type="text" id="userId" value="test-user-123"></label>
        <button onclick="connect()">Connect to Notifications</button>
        <button onclick="disconnect()">Disconnect</button>
    </div>
    
    <div id="status" class="status">Not connected</div>
    
    <h3>📥 Live Notifications:</h3>
    <div id="notifications"></div>
    
    <h3>🧪 Test Actions:</h3>
    <button onclick="createTestUser()">1. Create Test User</button>
    <button onclick="createTestNotification()">2. Create Test Notification</button>
    <button onclick="markAsRead()">3. Mark Last as Read</button>

    <script>
        let eventSource = null;
        let lastNotificationId = null;
        
        function connect() {
            const userId = document.getElementById('userId').value;
            if (!userId) {
                alert('Please enter a User ID');
                return;
            }
            
            disconnect(); // Close existing connection
            
            const url = `http://localhost:3001/notifications/stream?user_id=${userId}`;
            eventSource = new EventSource(url);
            
            eventSource.onopen = function() {
                updateStatus('Connected to notification stream', 'connected');
            };
            
            eventSource.onmessage = function(event) {
                const data = JSON.parse(event.data);
                console.log('Received:', data);
                displayNotification(data);
            };
            
            eventSource.onerror = function(event) {
                updateStatus('Connection error - check if server is running on port 3001', 'error');
                console.error('SSE Error:', event);
            };
        }
        
        function disconnect() {
            if (eventSource) {
                eventSource.close();
                eventSource = null;
                updateStatus('Disconnected', 'error');
            }
        }
        
        function updateStatus(message, type) {
            const status = document.getElementById('status');
            status.textContent = message;
            status.className = `status ${type}`;
        }
        
        function displayNotification(data) {
            const container = document.getElementById('notifications');
            const div = document.createElement('div');
            div.className = 'notification';
            
            if (data.event === 'new_notification') {
                lastNotificationId = data.notification.id;
                div.innerHTML = `
                    <strong>🔔 ${data.type.toUpperCase()} Notification</strong><br>
                    <strong>ID:</strong> ${data.notification.id}<br>
                    <strong>Content:</strong> ${data.notification.content || data.notification.review_title || 'N/A'}<br>
                    <strong>From:</strong> ${data.notification.from_name || 'System'}<br>
                    <strong>Time:</strong> ${new Date().toLocaleTimeString()}
                `;
            } else if (data.event === 'notification_read') {
                div.innerHTML = `
                    <strong>✅ Notification Marked as Read</strong><br>
                    <strong>ID:</strong> ${data.notification.notification_id}<br>
                    <strong>Time:</strong> ${new Date().toLocaleTimeString()}
                `;
            } else {
                div.innerHTML = `
                    <strong>📡 ${data.event}</strong><br>
                    ${JSON.stringify(data.notification, null, 2)}
                `;
            }
            
            container.insertBefore(div, container.firstChild);
        }
        
        // Test functions
        async function createTestUser() {
            const userId = document.getElementById('userId').value;
            try {
                const response = await fetch('http://localhost:3001/users', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({
                        id: userId,
                        username: userId,
                        full_name: `Test User ${userId}`
                    })
                });
                
                if (response.ok) {
                    alert('✅ Test user created successfully!');
                } else {
                    const error = await response.text();
                    alert(`❌ Error: ${error}`);
                }
            } catch (error) {
                alert(`❌ Network error: ${error.message}`);
            }
        }
        
        async function createTestNotification() {
            const userId = document.getElementById('userId').value;
            try {
                const response = await fetch('http://localhost:3001/notifications/product-owner', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({
                        id: `notif-${Date.now()}`,
                        owner_id: userId,
                        product_id: 'test-product-123',
                        product_name: 'Test Product',
                        business_id: 'test-business-123',
                        review_title: 'This is a test notification!',
                        from_name: 'Test System',
                        from_id: 'system',
                        read: false
                    })
                });
                
                if (response.ok) {
                    alert('✅ Test notification created! Check the live notifications above.');
                } else {
                    const error = await response.text();
                    alert(`❌ Error: ${error}`);
                }
            } catch (error) {
                alert(`❌ Network error: ${error.message}`);
            }
        }
        
        async function markAsRead() {
            if (!lastNotificationId) {
                alert('❌ No notification to mark as read. Create a test notification first.');
                return;
            }
            
            try {
                const response = await fetch(`http://localhost:3001/notifications/${lastNotificationId}/read?type=owner`, {
                    method: 'POST'
                });
                
                if (response.ok) {
                    alert('✅ Notification marked as read!');
                } else {
                    const error = await response.text();
                    alert(`❌ Error: ${error}`);
                }
            } catch (error) {
                alert(`❌ Network error: ${error.message}`);
            }
        }
        
        // Auto-connect on page load
        window.onload = function() {
            updateStatus('Ready to connect', '');
        };
        
        // Clean up on page unload
        window.onbeforeunload = function() {
            disconnect();
        };
    </script>
</body>
</html>
```

## 2. How to Test

1. **Start your notification service:**
   ```bash
   go run main.go
   ```

2. **Open the HTML file** in your browser

3. **Follow the test steps:**
   - Click "Connect to Notifications" 
   - Click "1. Create Test User"
   - Click "2. Create Test Notification" 
   - Watch the notification appear in real-time! 🎉
   - Click "3. Mark Last as Read" to see read status updates

## 3. Integration into Your App

### React Example:
```jsx
import { useState, useEffect } from 'react';

function useNotifications(userId) {
    const [notifications, setNotifications] = useState([]);
    const [isConnected, setIsConnected] = useState(false);

    useEffect(() => {
        if (!userId) return;

        const eventSource = new EventSource(
            `http://localhost:3001/notifications/stream?user_id=${userId}`
        );

        eventSource.onopen = () => setIsConnected(true);
        
        eventSource.onmessage = (event) => {
            const data = JSON.parse(event.data);
            if (data.event === 'new_notification') {
                setNotifications(prev => [data.notification, ...prev]);
            }
        };

        eventSource.onerror = () => setIsConnected(false);

        return () => eventSource.close();
    }, [userId]);

    return { notifications, isConnected };
}

// Usage in component:
function NotificationComponent({ userId }) {
    const { notifications, isConnected } = useNotifications(userId);
    
    return (
        <div>
            <div>Status: {isConnected ? '🟢 Connected' : '🔴 Disconnected'}</div>
            {notifications.map(notif => (
                <div key={notif.id}>{notif.content || notif.review_title}</div>
            ))}
        </div>
    );
}
```

### Vue.js Example:
```vue
<template>
  <div>
    <div>Status: {{ isConnected ? '🟢 Connected' : '🔴 Disconnected' }}</div>
    <div v-for="notif in notifications" :key="notif.id">
      {{ notif.content || notif.review_title }}
    </div>
  </div>
</template>

<script>
export default {
  props: ['userId'],
  data() {
    return {
      notifications: [],
      isConnected: false,
      eventSource: null
    };
  },
  watch: {
    userId: {
      immediate: true,
      handler(newUserId) {
        this.connect(newUserId);
      }
    }
  },
  methods: {
    connect(userId) {
      if (this.eventSource) this.eventSource.close();
      if (!userId) return;

      this.eventSource = new EventSource(
        `http://localhost:3001/notifications/stream?user_id=${userId}`
      );

      this.eventSource.onopen = () => this.isConnected = true;
      this.eventSource.onmessage = (event) => {
        const data = JSON.parse(event.data);
        if (data.event === 'new_notification') {
          this.notifications.unshift(data.notification);
        }
      };
      this.eventSource.onerror = () => this.isConnected = false;
    }
  },
  beforeUnmount() {
    if (this.eventSource) this.eventSource.close();
  }
};
</script>
```

## 4. Production Considerations

- **CORS**: Configure your server for production domains
- **Authentication**: Add proper user authentication
- **Error Handling**: Implement robust reconnection logic
- **Rate Limiting**: Consider implementing rate limits
- **HTTPS**: Use secure connections in production

## 🎯 Next Steps

1. Test with the HTML example above
2. Integrate the patterns into your existing frontend
3. Add proper error handling and reconnection logic
4. Consider implementing notification persistence for offline users

**Need help?** Check the detailed guide at `guides/frontend-sse-integration.md` for advanced patterns and best practices!