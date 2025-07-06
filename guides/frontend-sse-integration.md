# Frontend SSE Integration Guide

## Overview

This guide explains how to integrate Server-Sent Events (SSE) with your notification service frontend. The service now supports both traditional polling and real-time SSE notifications.

> **✨ Updated for Refactored Architecture**: This guide is compatible with the new modular codebase structure.

## Table of Contents

1. [Quick Start](#quick-start)
2. [API Endpoints](#api-endpoints)
3. [SSE Implementation](#sse-implementation)
4. [Fallback Strategy](#fallback-strategy)
5. [Event Types](#event-types)
6. [Error Handling](#error-handling)
7. [Best Practices](#best-practices)
8. [Examples](#examples)

## Quick Start

### 🚀 Basic SSE Connection (Copy & Paste Ready)

```javascript
const userId = "your-user-id"; // Replace with actual user ID
const eventSource = new EventSource(`http://localhost:3001/notifications/stream?user_id=${userId}`);

eventSource.onmessage = function(event) {
    const data = JSON.parse(event.data);
    console.log('New notification:', data);
    
    // Handle different notification types
    if (data.event === 'new_notification') {
        showNotificationToUser(data.notification);
    } else if (data.event === 'notification_read') {
        updateNotificationStatus(data.notification);
    }
};

eventSource.onerror = function(event) {
    console.error('SSE connection error:', event);
    // Implement reconnection logic here
};

// Don't forget to close the connection when done
// eventSource.close();
```

### 🎯 Key Points:
- **User ID is required** in the URL query parameter
- **Multiple connections** per user are supported (different tabs/devices)
- **Automatic reconnection** is recommended for production use

## API Endpoints

### Existing Polling Endpoints (Still Available)
- `GET /notifications?user_id=X` - Get all notifications
- `GET /notifications/latest?user_id=X` - Get latest notifications
- `GET /notifications/unread?user_id=X` - Get unread notifications ✨ **Now Connected!**
- `POST /notifications/:id/read?type=X` - Mark notification as read
- `DELETE /notifications?user_id=X` - Delete read notifications

### New SSE Endpoint
- `GET /notifications/stream?user_id=X` - Real-time notification stream

### Create Notification Endpoints
- `POST /notifications/product-owner` - Create product owner notification
- `POST /notifications/reply` - Create reply notification
- `POST /users` - Create user

## SSE Implementation

### 1. Basic Connection Setup

```javascript
class NotificationService {
    constructor(userId, baseUrl = 'http://localhost:3001') {
        this.userId = userId;
        this.baseUrl = baseUrl;
        this.eventSource = null;
        this.isConnected = false;
        this.reconnectAttempts = 0;
        this.maxReconnectAttempts = 5;
        this.reconnectDelay = 1000; // Start with 1 second
    }

    connect() {
        if (this.eventSource) {
            this.disconnect();
        }

        const url = `${this.baseUrl}/notifications/stream?user_id=${this.userId}`;
        this.eventSource = new EventSource(url);

        this.eventSource.onopen = (event) => {
            console.log('SSE connection opened');
            this.isConnected = true;
            this.reconnectAttempts = 0;
            this.reconnectDelay = 1000;
            this.onConnectionOpen?.(event);
        };

        this.eventSource.onmessage = (event) => {
            try {
                const data = JSON.parse(event.data);
                this.handleMessage(data);
            } catch (error) {
                console.error('Error parsing SSE message:', error);
            }
        };

        this.eventSource.onerror = (event) => {
            console.error('SSE connection error:', event);
            this.isConnected = false;
            this.handleReconnection();
            this.onConnectionError?.(event);
        };
    }

    handleMessage(data) {
        console.log('Received notification:', data);
        
        switch (data.event) {
            case 'connected':
                this.onConnected?.(data);
                break;
            case 'new_notification':
                this.onNewNotification?.(data);
                break;
            case 'existing_notification':
                this.onExistingNotification?.(data);
                break;
            case 'notification_read':
                this.onNotificationRead?.(data);
                break;
            default:
                this.onUnknownEvent?.(data);
        }
    }

    handleReconnection() {
        if (this.reconnectAttempts < this.maxReconnectAttempts) {
            this.reconnectAttempts++;
            console.log(`Attempting to reconnect (${this.reconnectAttempts}/${this.maxReconnectAttempts}) in ${this.reconnectDelay}ms`);
            
            setTimeout(() => {
                this.connect();
            }, this.reconnectDelay);
            
            // Exponential backoff
            this.reconnectDelay = Math.min(this.reconnectDelay * 2, 30000);
        } else {
            console.error('Max reconnection attempts reached');
            this.onMaxReconnectAttemptsReached?.();
        }
    }

    disconnect() {
        if (this.eventSource) {
            this.eventSource.close();
            this.eventSource = null;
            this.isConnected = false;
        }
    }

    // Callback methods (override these)
    onConnectionOpen(event) {}
    onConnected(data) {}
    onNewNotification(data) {}
    onExistingNotification(data) {}
    onNotificationRead(data) {}
    onConnectionError(event) {}
    onUnknownEvent(data) {}
    onMaxReconnectAttemptsReached() {}
}
```

### 2. Usage Example

```javascript
const notificationService = new NotificationService('user-123');

// Set up event handlers
notificationService.onConnected = (data) => {
    console.log('Connected to notification stream:', data.notification.message);
    showConnectionStatus('Connected');
};

notificationService.onNewNotification = (data) => {
    console.log('New notification received:', data);
    displayNotification(data.notification);
    updateNotificationBadge();
    playNotificationSound();
};

notificationService.onExistingNotification = (data) => {
    console.log('Existing notification loaded:', data);
    displayNotification(data.notification, false); // Don't show as "new"
};

notificationService.onNotificationRead = (data) => {
    console.log('Notification marked as read:', data);
    markNotificationAsReadInUI(data.notification.notification_id);
};

notificationService.onConnectionError = (event) => {
    console.error('Connection lost, falling back to polling');
    showConnectionStatus('Disconnected - Using polling');
    startPollingFallback();
};

notificationService.onMaxReconnectAttemptsReached = () => {
    console.error('Could not reconnect to SSE, using polling only');
    showConnectionStatus('Offline - Using polling');
    startPollingFallback();
};

// Start the connection
notificationService.connect();
```

## Fallback Strategy

### Hybrid Approach (Recommended)

```javascript
class HybridNotificationService {
    constructor(userId, baseUrl = 'http://localhost:3001') {
        this.userId = userId;
        this.baseUrl = baseUrl;
        this.sseService = new NotificationService(userId, baseUrl);
        this.pollingInterval = null;
        this.pollingDelay = 30000; // 30 seconds
        this.useSSE = true;
    }

    start() {
        if (this.useSSE && this.supportsSSE()) {
            this.startSSE();
        } else {
            this.startPolling();
        }
    }

    supportsSSE() {
        return typeof EventSource !== 'undefined';
    }

    startSSE() {
        this.sseService.onMaxReconnectAttemptsReached = () => {
            console.log('SSE failed, falling back to polling');
            this.startPolling();
        };

        this.sseService.connect();
    }

    startPolling() {
        this.stopPolling(); // Clear any existing polling
        
        const poll = async () => {
            try {
                const response = await fetch(`${this.baseUrl}/notifications/unread?user_id=${this.userId}`);
                const data = await response.json();
                this.handlePollingData(data);
            } catch (error) {
                console.error('Polling error:', error);
            }
        };

        // Initial poll
        poll();
        
        // Set up interval
        this.pollingInterval = setInterval(poll, this.pollingDelay);
    }

    stopPolling() {
        if (this.pollingInterval) {
            clearInterval(this.pollingInterval);
            this.pollingInterval = null;
        }
    }

    handlePollingData(data) {
        // Process polling data and update UI
        const allNotifications = [
            ...(data.user_notifications || []),
            ...(data.owner_notifications || [])
        ];
        
        allNotifications.forEach(notification => {
            this.onNewNotification?.({
                type: data.user_notifications?.includes(notification) ? 'user' : 'owner',
                notification: notification,
                event: 'new_notification'
            });
        });
    }

    stop() {
        this.sseService.disconnect();
        this.stopPolling();
    }

    // Callback methods
    onNewNotification(data) {}
    onNotificationRead(data) {}
}
```

## Event Types

### SSE Message Structure

```javascript
{
    "user_id": "user-123",
    "type": "user" | "owner" | "system",
    "event": "connected" | "new_notification" | "existing_notification" | "notification_read",
    "notification": {
        // Notification object (varies by type)
    }
}
```

### Event Types Explained

1. **`connected`** - Initial connection confirmation
   ```javascript
   {
       "event": "connected",
       "type": "system",
       "notification": {
           "message": "Connected to notification stream",
           "time": "2024-01-15T10:30:00Z"
       }
   }
   ```

2. **`new_notification`** - Real-time new notification
   ```javascript
   {
       "event": "new_notification",
       "type": "user", // or "owner"
       "notification": {
           "id": "notif-123",
           "content": "Someone replied to your comment",
           "created_at": "2024-01-15T10:30:00Z",
           "read": false,
           // ... other notification fields
       }
   }
   ```

3. **`existing_notification`** - Unread notifications sent on connection
   ```javascript
   {
       "event": "existing_notification",
       "type": "owner",
       "notification": {
           "id": "notif-456",
           "review_title": "Great product!",
           "created_at": "2024-01-15T09:15:00Z",
           "read": false,
           // ... other notification fields
       }
   }
   ```

4. **`notification_read`** - Notification marked as read
   ```javascript
   {
       "event": "notification_read",
       "type": "user",
       "notification": {
           "notification_id": "notif-123",
           "type": "user",
           "read": true,
           "timestamp": "2024-01-15T10:35:00Z"
       }
   }
   ```

## Error Handling

### Connection Issues

```javascript
// Handle different error scenarios
notificationService.onConnectionError = (event) => {
    if (event.target.readyState === EventSource.CLOSED) {
        console.log('SSE connection was closed');
    } else if (event.target.readyState === EventSource.CONNECTING) {
        console.log('SSE is reconnecting...');
    }
    
    // Show user-friendly message
    showNotificationStatus('Connection lost, retrying...');
};

// Handle network issues
window.addEventListener('online', () => {
    if (!notificationService.isConnected) {
        console.log('Network restored, reconnecting...');
        notificationService.connect();
    }
});

window.addEventListener('offline', () => {
    console.log('Network lost');
    showNotificationStatus('Offline');
});
```

### API Error Handling

```javascript
// When marking notifications as read
async function markAsRead(notificationId, type) {
    try {
        const response = await fetch(`/notifications/${notificationId}/read?type=${type}`, {
            method: 'POST'
        });
        
        if (!response.ok) {
            throw new Error(`HTTP ${response.status}: ${response.statusText}`);
        }
        
        // SSE will automatically broadcast the read status
        console.log('Notification marked as read');
        
    } catch (error) {
        console.error('Failed to mark notification as read:', error);
        // Show error to user
        showError('Failed to mark notification as read');
    }
}
```

## Best Practices

### 1. Resource Management

```javascript
// Clean up when component unmounts or page unloads
window.addEventListener('beforeunload', () => {
    notificationService.disconnect();
});

// In React
useEffect(() => {
    notificationService.connect();
    
    return () => {
        notificationService.disconnect();
    };
}, []);
```

### 2. Performance Optimization

```javascript
// Debounce UI updates for rapid notifications
const debouncedUpdateUI = debounce((notifications) => {
    updateNotificationList(notifications);
}, 100);

notificationService.onNewNotification = (data) => {
    debouncedUpdateUI(data);
};

// Limit notification history in memory
const MAX_NOTIFICATIONS = 100;
let notificationHistory = [];

function addNotification(notification) {
    notificationHistory.unshift(notification);
    if (notificationHistory.length > MAX_NOTIFICATIONS) {
        notificationHistory = notificationHistory.slice(0, MAX_NOTIFICATIONS);
    }
}
```

### 3. User Experience

```javascript
// Show connection status
function showConnectionStatus(status) {
    const statusElement = document.getElementById('connection-status');
    statusElement.textContent = status;
    statusElement.className = status.includes('Connected') ? 'status-connected' : 'status-disconnected';
}

// Notification permissions
async function requestNotificationPermission() {
    if ('Notification' in window) {
        const permission = await Notification.requestPermission();
        return permission === 'granted';
    }
    return false;
}

// Browser notifications for new notifications
notificationService.onNewNotification = (data) => {
    if (document.hidden && Notification.permission === 'granted') {
        new Notification('New Notification', {
            body: data.notification.content || data.notification.review_title,
            icon: '/notification-icon.png'
        });
    }
};
```

## Examples

### React Hook Example

```javascript
import { useState, useEffect, useCallback } from 'react';

function useNotifications(userId) {
    const [notifications, setNotifications] = useState([]);
    const [isConnected, setIsConnected] = useState(false);
    const [connectionStatus, setConnectionStatus] = useState('Disconnected');

    const notificationService = useMemo(() => 
        new NotificationService(userId), [userId]
    );

    useEffect(() => {
        notificationService.onConnected = () => {
            setIsConnected(true);
            setConnectionStatus('Connected');
        };

        notificationService.onNewNotification = (data) => {
            setNotifications(prev => [data.notification, ...prev]);
        };

        notificationService.onExistingNotification = (data) => {
            setNotifications(prev => [...prev, data.notification]);
        };

        notificationService.onNotificationRead = (data) => {
            setNotifications(prev => 
                prev.map(notif => 
                    notif.id === data.notification.notification_id 
                        ? { ...notif, read: true }
                        : notif
                )
            );
        };

        notificationService.onConnectionError = () => {
            setIsConnected(false);
            setConnectionStatus('Disconnected');
        };

        notificationService.connect();

        return () => {
            notificationService.disconnect();
        };
    }, [notificationService]);

    const markAsRead = useCallback(async (notificationId, type) => {
        try {
            const response = await fetch(`/notifications/${notificationId}/read?type=${type}`, {
                method: 'POST'
            });
            if (!response.ok) throw new Error('Failed to mark as read');
        } catch (error) {
            console.error('Error marking notification as read:', error);
        }
    }, []);

    return {
        notifications,
        isConnected,
        connectionStatus,
        markAsRead
    };
}

// Usage in component
function NotificationComponent({ userId }) {
    const { notifications, isConnected, connectionStatus, markAsRead } = useNotifications(userId);

    return (
        <div>
            <div className={`status ${isConnected ? 'connected' : 'disconnected'}`}>
                {connectionStatus}
            </div>
            <div className="notifications">
                {notifications.map(notification => (
                    <div 
                        key={notification.id} 
                        className={`notification ${notification.read ? 'read' : 'unread'}`}
                        onClick={() => markAsRead(notification.id, 'user')}
                    >
                        {notification.content || notification.review_title}
                    </div>
                ))}
            </div>
        </div>
    );
}
```

### Vue.js Example

```javascript
// Vue 3 Composition API
import { ref, onMounted, onUnmounted } from 'vue';

export function useNotifications(userId) {
    const notifications = ref([]);
    const isConnected = ref(false);
    const connectionStatus = ref('Disconnected');
    
    let notificationService;

    onMounted(() => {
        notificationService = new NotificationService(userId);
        
        notificationService.onConnected = () => {
            isConnected.value = true;
            connectionStatus.value = 'Connected';
        };

        notificationService.onNewNotification = (data) => {
            notifications.value.unshift(data.notification);
        };

        notificationService.onConnectionError = () => {
            isConnected.value = false;
            connectionStatus.value = 'Disconnected';
        };

        notificationService.connect();
    });

    onUnmounted(() => {
        if (notificationService) {
            notificationService.disconnect();
        }
    });

    return {
        notifications,
        isConnected,
        connectionStatus
    };
}
```

### Vanilla JavaScript Example

```html
<!DOCTYPE html>
<html>
<head>
    <title>Notification Demo</title>
    <style>
        .notification {
            padding: 10px;
            margin: 5px 0;
            border-radius: 4px;
            background: #f0f0f0;
        }
        .notification.unread {
            background: #e3f2fd;
            font-weight: bold;
        }
        .status {
            padding: 5px 10px;
            border-radius: 4px;
            margin-bottom: 10px;
        }
        .status.connected {
            background: #c8e6c9;
            color: #2e7d32;
        }
        .status.disconnected {
            background: #ffcdd2;
            color: #c62828;
        }
    </style>
</head>
<body>
    <div id="app">
        <div id="status" class="status disconnected">Disconnected</div>
        <div id="notifications"></div>
    </div>

    <script>
        const userId = 'user-123'; // Replace with actual user ID
        const notificationService = new NotificationService(userId);
        
        const statusEl = document.getElementById('status');
        const notificationsEl = document.getElementById('notifications');

        notificationService.onConnected = (data) => {
            statusEl.textContent = 'Connected';
            statusEl.className = 'status connected';
        };

        notificationService.onNewNotification = (data) => {
            const notifEl = document.createElement('div');
            notifEl.className = 'notification unread';
            notifEl.textContent = data.notification.content || data.notification.review_title;
            notifEl.onclick = () => markAsRead(data.notification.id, data.type);
            notificationsEl.insertBefore(notifEl, notificationsEl.firstChild);
        };

        notificationService.onConnectionError = () => {
            statusEl.textContent = 'Disconnected';
            statusEl.className = 'status disconnected';
        };

        async function markAsRead(notificationId, type) {
            try {
                await fetch(`/notifications/${notificationId}/read?type=${type}`, {
                    method: 'POST'
                });
            } catch (error) {
                console.error('Failed to mark as read:', error);
            }
        }

        notificationService.connect();
    </script>
</body>
</html>
```

## Testing SSE Connection

### Browser Developer Tools

1. Open Network tab
2. Look for `/notifications/stream` request
3. Should show "EventStream" type
4. Check for continuous connection

### Manual Testing

```bash
# Test SSE endpoint directly
curl -N -H "Accept: text/event-stream" \
  "http://localhost:3001/notifications/stream?user_id=test-user"

# Should output:
# data: {"user_id":"test-user","type":"system","event":"connected",...}
```

### Create Test Notifications

```bash
# Create a test user first
curl -X POST http://localhost:3001/users \
  -H "Content-Type: application/json" \
  -d '{"id":"test-user","username":"testuser","full_name":"Test User"}'

# Create a test notification
curl -X POST http://localhost:3001/notifications/product-owner \
  -H "Content-Type: application/json" \
  -d '{
    "id":"test-notif-1",
    "owner_id":"test-user",
    "product_id":"prod-1",
    "product_name":"Test Product",
    "business_id":"biz-1",
    "review_title":"Great product!",
    "from_name":"John Doe",
    "from_id":"user-2",
    "read":false
  }'
```

## Troubleshooting

### Common Issues

1. **CORS Errors**: Ensure your frontend domain is allowed in CORS settings
2. **Connection Drops**: Check network stability and implement proper reconnection
3. **Memory Leaks**: Always disconnect SSE when components unmount
4. **Duplicate Notifications**: Implement deduplication logic based on notification IDs

### Debug Mode

```javascript
// Enable debug logging
const notificationService = new NotificationService(userId);
notificationService.debug = true;

// Override handleMessage for debugging
const originalHandleMessage = notificationService.handleMessage;
notificationService.handleMessage = function(data) {
    console.log('[DEBUG] SSE Message:', data);
    originalHandleMessage.call(this, data);
};
```

---

## Summary

You now have a robust notification system that supports:

- ✅ **Real-time SSE notifications**
- ✅ **Polling fallback for reliability**
- ✅ **Automatic reconnection**
- ✅ **Read status synchronization**
- ✅ **Multiple client support**
- ✅ **Cross-browser compatibility**

The service maintains backward compatibility with existing polling clients while providing real-time capabilities for modern applications.