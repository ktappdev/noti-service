# 📋 Notification REST API Specification

## Overview
This document specifies the REST API endpoints for notification management. These endpoints handle actions and queries, while the SSE system handles real-time delivery.

**Base URL:** `https://notifications.reviewit.gy`

## Authentication
All endpoints require user authentication. Include user identification in requests as specified per endpoint.

---

## 📨 Core Notification Endpoints

### 1. Get All Notifications
**Purpose:** Fallback method to get notifications (SSE is preferred)

```http
GET /notifications?user_id={user_id}
```

**Query Parameters:**
- `user_id` (required): The user ID to get notifications for
- `limit` (optional): Number of notifications to return (default: 50)
- `offset` (optional): Pagination offset (default: 0)
- `read` (optional): Filter by read status (`true`, `false`, or omit for all)

**Response:**
```json
{
  "user_notifications": [
    {
      "id": "123",
      "parent_id": "comment_456",
      "parent_user_id": "user_2wtRg8rDyrbdImQYvsIMlCOQ7qM",
      "message": "Someone replied to your comment",
      "read": false,
      "created_at": "2025-07-06T21:56:31Z"
    }
  ],
  "owner_notifications": [
    {
      "id": "124",
      "owner_id": "user_2wtRg8rDyrbdImQYvsIMlCOQ7qM",
      "product_id": "product_789",
      "message": "New comment on your product",
      "read": false,
      "created_at": "2025-07-06T21:58:15Z"
    }
  ],
  "total_count": 2,
  "unread_count": 2
}
```

**Status Codes:**
- `200`: Success
- `400`: Invalid user_id parameter
- `500`: Server error

---

### 2. Mark Notification as Read
**Purpose:** Mark a specific notification as read

```http
PUT /notifications/{notification_id}/read
```

**Path Parameters:**
- `notification_id` (required): The notification ID to mark as read

**Query Parameters:**
- `user_id` (required): The user ID (for authorization)
- `type` (required): Notification type (`"user"` or `"owner"`)

**Request Body:** None

**Response:**
```json
{
  "message": "Notification marked as read",
  "notification_id": "123",
  "success": true
}
```

**Status Codes:**
- `200`: Successfully marked as read
- `400`: Invalid parameters
- `404`: Notification not found
- `500`: Server error

**SSE Side Effect:**
After successful update, SSE will broadcast:
```json
{
  "userID": "user_2wtRg8rDyrbdImQYvsIMlCOQ7qM",
  "type": "user",
  "event": "notification_read",
  "notification": {
    "message": "Notification marked as read",
    "notification_id": "123"
  }
}
```

---

### 3. Mark All Notifications as Read
**Purpose:** Mark all unread notifications as read for a user

```http
PUT /notifications/read-all
```

**Query Parameters:**
- `user_id` (required): The user ID
- `type` (optional): Notification type (`"user"`, `"owner"`, or omit for both)

**Request Body:** None

**Response:**
```json
{
  "message": "All notifications marked as read",
  "user_notifications_updated": 5,
  "owner_notifications_updated": 3,
  "total_updated": 8,
  "success": true
}
```

**Status Codes:**
- `200`: Successfully marked all as read
- `400`: Invalid user_id parameter
- `500`: Server error

**SSE Side Effect:**
Multiple `notification_read` events will be sent for each notification marked as read.

---

### 4. Get Notification Details
**Purpose:** Get detailed information about a specific notification

```http
GET /notifications/{notification_id}
```

**Path Parameters:**
- `notification_id` (required): The notification ID

**Query Parameters:**
- `user_id` (required): The user ID (for authorization)
- `type` (required): Notification type (`"user"` or `"owner"`)

**Response:**
```json
{
  "notification": {
    "id": "123",
    "parent_id": "comment_456",
    "parent_user_id": "user_2wtRg8rDyrbdImQYvsIMlCOQ7qM",
    "message": "Someone replied to your comment",
    "read": true,
    "created_at": "2025-07-06T21:56:31Z",
    "read_at": "2025-07-06T22:10:15Z"
  },
  "type": "user"
}
```

**Status Codes:**
- `200`: Success
- `400`: Invalid parameters
- `404`: Notification not found
- `500`: Server error

---

### 5. Delete Notification
**Purpose:** Delete a specific notification

```http
DELETE /notifications/{notification_id}
```

**Path Parameters:**
- `notification_id` (required): The notification ID to delete

**Query Parameters:**
- `user_id` (required): The user ID (for authorization)
- `type` (required): Notification type (`"user"` or `"owner"`)

**Request Body:** None

**Response:**
```json
{
  "message": "Notification deleted successfully",
  "notification_id": "123",
  "success": true
}
```

**Status Codes:**
- `200`: Successfully deleted
- `400`: Invalid parameters
- `404`: Notification not found
- `500`: Server error

---

### 6. Get Unread Count
**Purpose:** Get count of unread notifications

```http
GET /notifications/unread-count?user_id={user_id}
```

**Query Parameters:**
- `user_id` (required): The user ID

**Response:**
```json
{
  "user_notifications_unread": 3,
  "owner_notifications_unread": 2,
  "total_unread": 5
}
```

**Status Codes:**
- `200`: Success
- `400`: Invalid user_id parameter
- `500`: Server error

---

## 🔧 Frontend Implementation Examples

### Mark Notification as Read
```javascript
async function markNotificationAsRead(notificationId, userId, type) {
  try {
    const response = await fetch(
      `/notifications/${notificationId}/read?user_id=${userId}&type=${type}`,
      { method: 'PUT' }
    );
    
    if (!response.ok) {
      throw new Error(`HTTP ${response.status}`);
    }
    
    const result = await response.json();
    console.log('Notification marked as read:', result);
    
    // Don't update UI here - wait for SSE event
    
  } catch (error) {
    console.error('Failed to mark notification as read:', error);
    // Show error message to user
  }
}
```

### Mark All as Read
```javascript
async function markAllAsRead(userId) {
  try {
    const response = await fetch(
      `/notifications/read-all?user_id=${userId}`,
      { method: 'PUT' }
    );
    
    const result = await response.json();
    console.log(`Marked ${result.total_updated} notifications as read`);
    
  } catch (error) {
    console.error('Failed to mark all as read:', error);
  }
}
```

### Get Notification Details
```javascript
async function getNotificationDetails(notificationId, userId, type) {
  try {
    const response = await fetch(
      `/notifications/${notificationId}?user_id=${userId}&type=${type}`
    );
    
    if (!response.ok) {
      throw new Error(`HTTP ${response.status}`);
    }
    
    const result = await response.json();
    return result.notification;
    
  } catch (error) {
    console.error('Failed to get notification details:', error);
    return null;
  }
}
```

### Delete Notification
```javascript
async function deleteNotification(notificationId, userId, type) {
  try {
    const response = await fetch(
      `/notifications/${notificationId}?user_id=${userId}&type=${type}`,
      { method: 'DELETE' }
    );
    
    if (!response.ok) {
      throw new Error(`HTTP ${response.status}`);
    }
    
    const result = await response.json();
    console.log('Notification deleted:', result);
    
    // Remove from local state immediately (no SSE event for deletes)
    notifications = notifications.filter(n => n.id !== notificationId);
    updateUI();
    
  } catch (error) {
    console.error('Failed to delete notification:', error);
  }
}
```

---

## 🎯 Integration with SSE

### Recommended Frontend Pattern
```javascript
class NotificationManager {
  constructor() {
    this.notifications = [];
    this.eventSource = null;
  }
  
  // SSE handles real-time updates
  connectSSE(userId) {
    this.eventSource = new EventSource(
      `https://notifications.reviewit.gy/notifications/stream?user_id=${userId}`
    );
    
    this.eventSource.onmessage = (event) => {
      const data = JSON.parse(event.data);
      this.handleSSEMessage(data);
    };
  }
  
  // REST API handles actions
  async markAsRead(notificationId, type) {
    await fetch(`/notifications/${notificationId}/read?user_id=${this.userId}&type=${type}`, {
      method: 'PUT'
    });
    // UI will update via SSE event
  }
  
  async markAllAsRead() {
    await fetch(`/notifications/read-all?user_id=${this.userId}`, {
      method: 'PUT'
    });
    // UI will update via multiple SSE events
  }
  
  async deleteNotification(notificationId, type) {
    await fetch(`/notifications/${notificationId}?user_id=${this.userId}&type=${type}`, {
      method: 'DELETE'
    });
    // Remove from local state immediately (no SSE for deletes)
    this.notifications = this.notifications.filter(n => n.id !== notificationId);
    this.updateUI();
  }
}
```

---

## ⚠️ Important Notes

### ✅ Best Practices:
- **Use SSE for notification delivery** (preferred over GET /notifications)
- **Use REST API for actions** (mark as read, delete, etc.)
- **Wait for SSE events** to update UI after actions
- **Handle errors gracefully** with user feedback
- **Don't poll** the GET endpoints - use SSE instead

### 🔄 Data Flow:
1. **Page Load**: SSE connects → delivers all unread notifications
2. **User Action**: REST API call → database update → SSE event → UI update
3. **New Notifications**: SSE delivers in real-time
4. **Browser Refresh**: SSE reconnects → fresh notification load

### 🚨 Error Handling:
- **Network errors**: Show user-friendly messages
- **404 errors**: Notification might have been deleted by another client
- **500 errors**: Retry with exponential backoff
- **SSE disconnection**: Implement reconnection logic

### 🔒 Security:
- All endpoints validate `user_id` parameter
- Users can only access their own notifications
- Notification IDs are validated against user ownership