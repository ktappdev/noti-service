# 👍 Like Notification API Specification

## Overview
The like notification system allows users to receive real-time notifications when someone likes their comments or reviews.

## Endpoint

### Create Like Notification
**Purpose:** Notify a user when someone likes their content

```http
POST /notifications/like
```

**Request Body:**
```json
{
  "target_type": "comment",        // Required: "comment" or "review"
  "target_id": "comment_123",      // Required: ID of the liked content
  "from_id": "user_456",           // Required: User who liked the content
  "from_name": "John Doe",         // Required: Name of user who liked
  "product_id": "prod_789",        // Optional: Product context
  "read": false                    // Optional: defaults to false
}
```

**Response (201 Created):**
```json
{
  "id": "like_notif_789",
  "target_user_id": "user_123",
  "target_type": "comment",
  "target_id": "comment_123", 
  "from_id": "user_456",
  "from_name": "John Doe",
  "product_id": "prod_789",
  "created_at": "2024-01-15T10:30:00Z",
  "read": false
}
```

## Backend Auto-Handles
- ✅ **target_user_id** - Automatically looked up from ReviewIt database
- ✅ **User validation** - Checks both target and from users exist
- ✅ **Self-like prevention** - No notification if user likes own content
- ✅ **Real-time SSE** - Automatically broadcasts to target user
- ✅ **Database storage** - Stores in `like_notifications` table

## SSE Message Format
When a like notification is created, the target user receives:

```json
{
  "user_id": "user_123",
  "type": "like",
  "event": "new_notification",
  "notification": {
    "id": "like_notif_789",
    "target_user_id": "user_123",
    "target_type": "comment",
    "target_id": "comment_123",
    "from_id": "user_456", 
    "from_name": "John Doe",
    "product_id": "prod_789",
    "created_at": "2024-01-15T10:30:00Z",
    "read": false
  }
}
```

## Frontend Integration Examples

### JavaScript/React
```javascript
// Send like notification
const createLikeNotification = async (targetType, targetId, fromUserId, fromUserName, productId) => {
  const response = await fetch('https://notifications.reviewit.gy/notifications/like', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      target_type: targetType,    // "comment" or "review"
      target_id: targetId,        // ID of liked content
      from_id: fromUserId,        // Current user's ID
      from_name: fromUserName,    // Current user's name
      product_id: productId,      // Optional product context
      read: false
    })
  });
  
  if (response.ok) {
    console.log('Like notification sent!');
  }
};

// Usage examples
createLikeNotification('comment', 'comment_123', 'user_456', 'John Doe', 'prod_789');
createLikeNotification('review', 'review_456', 'user_789', 'Jane Smith', 'prod_123');
```

### Handle SSE Like Notifications
```javascript
eventSource.addEventListener('message', (event) => {
  const data = JSON.parse(event.data);
  
  if (data.type === 'like' && data.event === 'new_notification') {
    const notification = data.notification;
    
    // Show like notification
    showNotification({
      title: `${notification.from_name} liked your ${notification.target_type}`,
      message: `Someone appreciated your ${notification.target_type}!`,
      type: 'like',
      timestamp: notification.created_at
    });
  }
});
```

## Error Responses

**400 Bad Request:**
- Invalid `target_type` (must be "comment" or "review")
- Missing required fields
- Target user doesn't exist
- From user doesn't exist

**500 Internal Server Error:**
- ReviewIt database connection issues
- Database insertion errors

**200 OK (No notification created):**
- User trying to like their own content

## Database Schema
```sql
CREATE TABLE like_notifications (
    id VARCHAR(255) PRIMARY KEY,
    target_user_id VARCHAR(255) NOT NULL,
    target_type VARCHAR(50) NOT NULL CHECK (target_type IN ('comment', 'review')),
    target_id VARCHAR(255) NOT NULL,
    from_id VARCHAR(255) NOT NULL,
    from_name VARCHAR(255) NOT NULL,
    product_id VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    read BOOLEAN DEFAULT FALSE,
    FOREIGN KEY (target_user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (from_id) REFERENCES users(id) ON DELETE CASCADE
);
```