# 📡 SSE Notification System Specification

## Overview
The notification system uses Server-Sent Events (SSE) to deliver real-time notifications to the frontend. This document specifies the complete message format and expected behavior.

## Connection Details

### Endpoint
```
GET https://notifications.reviewit.gy/notifications/stream?user_id={user_id}
```

### Headers
- `Content-Type: text/event-stream`
- `Cache-Control: no-cache`
- `Connection: keep-alive`

### CORS
- Supports cross-origin requests from `https://reviewit.gy`
- No credentials required

## Message Format

All SSE messages follow this structure:
```
data: {JSON_OBJECT}\n\n
```

### Base Message Schema
```typescript
interface NotificationMessage {
  userID: string;           // The target user ID
  type: "system" | "user" | "owner";  // Message category
  event: string;            // Event type (see below)
  notification: object;     // Event-specific payload
}
```

## Event Types

### 1. Connection Events

#### `connected`
**When:** Immediately upon SSE connection establishment
**Purpose:** Confirms successful connection

```json
{
  "userID": "user_2wtRg8rDyrbdImQYvsIMlCOQ7qM",
  "type": "system",
  "event": "connected",
  "notification": {
    "message": "Connected to notification stream",
    "time": "2025-07-06T21:56:31Z"
  }
}
```

#### `heartbeat`
**When:** Every 30 seconds
**Purpose:** Keep connection alive

```json
{
  "type": "heartbeat",
  "timestamp": "2025-07-06T21:56:31Z"
}
```

### 2. Notification Events

#### `existing_notification`
**When:** After connection, sends all unread notifications
**Purpose:** Initial notification load (replaces REST API calls)

```json
{
  "userID": "user_2wtRg8rDyrbdImQYvsIMlCOQ7qM",
  "type": "user",  // or "owner"
  "event": "existing_notification",
  "notification": {
    "id": "123",
    "parent_id": "comment_456",
    "parent_user_id": "user_2wtRg8rDyrbdImQYvsIMlCOQ7qM",
    "message": "Someone replied to your comment",
    "read": false,
    "created_at": "2025-07-06T21:56:31Z"
  }
}
```

#### `new_notification`
**When:** Real-time, when new notifications are created
**Purpose:** Live notification delivery

```json
{
  "userID": "user_2wtRg8rDyrbdImQYvsIMlCOQ7qM",
  "type": "owner",  // or "user"
  "event": "new_notification",
  "notification": {
    "id": "124",
    "owner_id": "user_2wtRg8rDyrbdImQYvsIMlCOQ7qM",
    "product_id": "product_789",
    "message": "New comment on your product",
    "read": false,
    "created_at": "2025-07-06T21:58:15Z"
  }
}
```

#### `notification_read`
**When:** When a notification is marked as read
**Purpose:** Update UI to reflect read status

```json
{
  "userID": "user_2wtRg8rDyrbdImQYvsIMlCOQ7qM",
  "type": "owner",  // or "user"
  "event": "notification_read",
  "notification": {
    "message": "Notification marked as read",
    "notification_id": "123"
  }
}
```

## Notification Types

### User Notifications (`type: "user"`)
- Triggered when someone replies to user's comment
- Stored in `user_notifications` table
- Fields: `id`, `parent_id`, `parent_user_id`, `message`, `read`, `created_at`

### Owner Notifications (`type: "owner"`)
- Triggered when someone comments on user's product
- Stored in `product_owner_notifications` table  
- Fields: `id`, `owner_id`, `product_id`, `message`, `read`, `created_at`

## Frontend Implementation Guide

### 1. Basic Connection
```javascript
const eventSource = new EventSource(
  'https://notifications.reviewit.gy/notifications/stream?user_id=USER_ID'
);

eventSource.onmessage = (event) => {
  const data = JSON.parse(event.data);
  handleNotification(data);
};

eventSource.onerror = (error) => {
  console.error('SSE connection error:', error);
  // Implement reconnection logic
};
```

### 2. Message Handling
```javascript
function handleNotification(data) {
  switch(data.event) {
    case 'connected':
      // Clear existing notifications, prepare for fresh data
      notifications = [];
      break;
      
    case 'existing_notification':
      // Add to notification list
      notifications.push(data.notification);
      updateUI();
      break;
      
    case 'new_notification':
      // Add new notification + show popup/toast
      notifications.push(data.notification);
      showNotificationPopup(data.notification);
      updateUI();
      break;
      
    case 'notification_read':
      // Mark notification as read
      const notif = notifications.find(n => n.id === data.notification.notification_id);
      if (notif) notif.read = true;
      updateUI();
      break;
      
    case 'heartbeat':
      // Connection alive - optional: update connection status
      break;
  }
}
```

### 3. State Management
```javascript
class NotificationManager {
  constructor() {
    this.notifications = [];
    this.eventSource = null;
  }
  
  connect(userId) {
    this.eventSource = new EventSource(
      `https://notifications.reviewit.gy/notifications/stream?user_id=${userId}`
    );
    this.eventSource.onmessage = (event) => {
      this.handleMessage(JSON.parse(event.data));
    };
  }
  
  getUnreadCount() {
    return this.notifications.filter(n => !n.read).length;
  }
  
  markAsRead(notificationId) {
    // Call REST API to mark as read
    // SSE will send notification_read event to update UI
  }
}
```

## Important Notes

### ✅ Do This:
- **Use SSE as single source of truth** for notifications
- **Clear local state on 'connected' event** (browser refresh)
- **Handle reconnections gracefully**
- **Store notifications in frontend state/memory**
- **Use REST API only for actions** (mark as read, delete)

### ❌ Don't Do This:
- **Don't fetch notifications via REST API on page load** (SSE sends them)
- **Don't ignore 'existing_notification' events**
- **Don't assume message order** (handle out-of-order delivery)
- **Don't store SSE data in localStorage** (always get fresh on connect)

## Error Handling

### Connection Errors
```javascript
eventSource.onerror = (error) => {
  if (eventSource.readyState === EventSource.CLOSED) {
    // Connection closed, implement exponential backoff reconnection
    setTimeout(() => reconnect(), Math.min(1000 * Math.pow(2, retryCount), 30000));
  }
};
```

### Message Parsing Errors
```javascript
eventSource.onmessage = (event) => {
  try {
    const data = JSON.parse(event.data);
    handleNotification(data);
  } catch (error) {
    console.error('Failed to parse SSE message:', event.data, error);
  }
};
```

## Testing

### Manual Testing
1. Open browser console on `https://reviewit.gy`
2. Connect to SSE stream
3. Create notifications via API
4. Verify real-time delivery
5. Test browser refresh (should get all unread notifications)

### Expected Flow
1. **Page Load** → SSE connects → `connected` event → `existing_notification` events
2. **New Activity** → `new_notification` event → Show popup/update UI
3. **Mark as Read** → REST API call → `notification_read` event → Update UI
4. **Browser Refresh** → Repeat step 1

## REST API Integration

While SSE handles notification delivery, you still need REST API for actions:

### Mark Notification as Read
```
PUT /notifications/{id}/read
```

### Get Notification Details (if needed)
```
GET /notifications/{id}
```

The SSE system will automatically send `notification_read` events when notifications are marked as read via the REST API.