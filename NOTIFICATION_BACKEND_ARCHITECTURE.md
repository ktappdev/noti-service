# ReviewIt Notification System Architecture

## Overview

This document outlines the recommended architecture for the ReviewIt notification system backend. It addresses the current challenges with notification creation, user management, and the handling of various notification types (comments, replies, likes).

## Core Concepts

### Notification Types

1. **Review Notifications**: When a user creates a review for a product
2. **Comment Notifications**: When a user comments directly on a review
3. **Reply Notifications**: When a user replies to a comment
4. **Like Notifications**: When a user likes a review or comment
5. **Product Owner Notifications**: When any activity happens on a product (reviews, comments)

### Entity Relationships

- **Reviews** belong to **Products** and are created by **Users**
- **Comments** can be attached to **Reviews** or can be replies to other **Comments**
- **Likes** can be attached to **Reviews** or **Comments**
- **Notifications** are sent to **Users** based on specific actions

## API Endpoints

### User Management

```
POST /users
```

**Purpose**: Create or update a user in the notification service database.

**Payload**:
```json
{
  "id": "user_123abc",
  "username": "johndoe",
  "full_name": "John Doe"
}
```

**Response**:
- `200 OK`: User created/updated successfully
- `400 Bad Request`: Invalid user data
- `500 Internal Server Error`: Server error

### Notification Creation

#### 1. Review Comment Notification

```
POST /notifications/comment
```

**Purpose**: Notify a review author when someone comments on their review.

**Payload**:
```json
{
  "id": "comment_123abc",
  "review_id": "review_456def",
  "comment_id": "comment_123abc",
  "from_id": "user_789ghi",
  "from_name": "Jane Smith",
  "content": "This is a comment on a review",
  "target_user_id": "user_101jkl",
  "product_id": "product_202mno",
  "read": false
}
```

#### 2. Comment Reply Notification

```
POST /notifications/reply
```

**Purpose**: Notify a comment author when someone replies to their comment.

**Payload**:
```json
{
  "id": "reply_123abc",
  "review_id": "review_456def",
  "comment_id": "reply_123abc",
  "parent_id": "comment_789ghi",
  "parent_user_id": "user_101jkl",
  "from_id": "user_202mno",
  "from_name": "Alex Johnson",
  "content": "This is a reply to a comment",
  "product_id": "product_303pqr",
  "read": false
}
```

**Key Differences**:
- `parent_id` is the ID of the parent comment (not a review ID or user ID)
- `parent_user_id` is the ID of the user who should receive the notification

#### 3. Like Notification

```
POST /notifications/like
```

**Purpose**: Notify a user when someone likes their review or comment.

**Payload**:
```json
{
  "id": "like_123abc",
  "target_type": "review", // or "comment"
  "target_id": "review_456def", // or comment ID
  "from_id": "user_789ghi",
  "from_name": "Chris Brown",
  "target_user_id": "user_101jkl",
  "product_id": "product_202mno",
  "review_id": "review_456def",
  "comment_id": null, // or comment ID if target_type is "comment"
  "read": false
}
```

#### 4. Product Owner Notification

```
POST /notifications/product-owner
```

**Purpose**: Notify a product owner when activity happens on their product.

**Payload**:
```json
{
  "id": "notification_123abc",
  "owner_id": "user_456def",
  "business_id": "business_789ghi",
  "review_title": "Great product!",
  "from_name": "Pat Wilson",
  "from_id": "user_101jkl",
  "product_id": "product_202mno",
  "product_name": "Awesome Widget",
  "review_id": "review_303pqr",
  "comment_id": null, // or comment ID if this is about a comment
  "notification_type": "review", // or "comment", "reply"
  "read": false
}
```

#### 5. System Notification (Admin / Platform)

```
POST /notifications/system
```

**Purpose**: Deliver platform-wide or admin-initiated messages to one or many users. Typical use-cases:

* Bug-fix announcement
* Product-claim approved / rejected
* Review approved / rejected
* Plan / billing notifications

**Payload**:
```json
{
  "id": "sys_123abc",
  "target_user_ids": ["user_456def"],            // Array to allow multi-cast (optional; empty array means broadcast to all)
  "title": "Your product claim was approved!",   // Short heading for UI badges / toasts
  "message": "Your claim for 'Awesome Widget' is now live.",
  "cta_url": "/dashboard/claims/awesome-widget", // Optional – where the user should be taken when they click the notification
  "icon": "info",                                // Optional – icon hint (info / success / warning / error)
  "read": false,
  "created_at": "2025-07-15T12:00:00Z",
  "notification_type": "system"                  // Distinguishes from user-generated notifications
}
```

**Notes**
1. If `target_user_ids` is omitted or empty, the notification is considered a broadcast; backend should enqueue it for all active users.
2. Backend must check that each `target_user_id` exists; create if missing.
3. Frontend should show these under a **System** tab or badge and render `title` + `message`.
4. SSE events use `event: system` to differentiate on the stream.

### Notification Management

```
GET /notifications?user_id=user_123abc
```

**Purpose**: Retrieve all notifications for a specific user.

**Response**:
```json
{
  "userNotifications": [...],
  "ownerNotifications": [...],
  "likeNotifications": [...]
}
```

```
PUT /notifications/{id}/read
```

**Purpose**: Mark a notification as read.

**Response**:
- `200 OK`: Notification marked as read
- `404 Not Found`: Notification not found
- `500 Internal Server Error`: Server error

### Real-time Notifications

```
GET /notifications/stream?user_id=user_123abc
```

**Purpose**: SSE endpoint for real-time notification updates.

**Events**:
- `notification`: New notification
- `read`: Notification marked as read
- `error`: Error event

## Implementation Guidelines

### User Creation Flow

1. **Always ensure users exist before sending notifications**:
   - Create the user in the main database
   - Create the user in the notification service database
   - Only then send notifications targeting that user

2. **User Creation Payload**:
   ```json
   {
     "id": "user_123abc",
     "username": "johndoe",
     "full_name": "John Doe"
   }
   ```

### Notification Creation Flow

1. **For comments on reviews**:
   - `parent_id` should be the review ID
   - `parent_user_id` should be the review author's user ID

2. **For replies to comments**:
   - `parent_id` should be the parent comment ID
   - `parent_user_id` should be the parent comment author's user ID

3. **For likes**:
   - `target_id` should be the review or comment ID
   - `target_user_id` should be the author of the liked content

### Error Handling

1. **User Not Found**:
   - If a notification targets a user that doesn't exist, return a clear error
   - Frontend should handle this by creating the user first, then retrying

2. **Parent Content Not Found**:
   - If a notification references a review or comment that doesn't exist, return a clear error
   - Frontend should validate content existence before sending notifications

## Frontend Integration

1. **User Creation**:
   - Call the user creation endpoint whenever a new user is created in the main system
   - Ensure user exists before sending any notifications

2. **SSE Connection**:
   - Connect to the SSE endpoint on user login
   - Handle reconnection if the connection is lost
   - Display toast notifications for new events

3. **Notification UI**:
   - Implement a notification bell with unread count
   - Show notification list with read/unread status
   - Provide direct navigation to the relevant content

## Current Issues and Solutions

### Issue 1: Parent User ID Confusion

**Problem**: The notification service expects different values for `parent_id` and `parent_user_id` than what the frontend is providing.

**Solution**:
- For comments on reviews: `parent_id` should be the review ID
- For replies to comments: `parent_id` should be the parent comment ID
- `parent_user_id` should always be the ID of the user who should receive the notification

### Issue 2: User Creation Before Notification

**Problem**: Notifications fail if the target user doesn't exist in the notification service database.

**Solution**:
- Always ensure both the sender and recipient users exist in the notification service before sending a notification
- Implement proper awaiting of user creation calls
- Add error handling for user creation failures

### Issue 3: Notification Navigation

**Problem**: Users need to navigate directly to the relevant content from a notification.

**Solution**:
- Include all necessary IDs in the notification payload
- Implement "View" buttons in the notification UI
- Use proper routing to navigate to the exact location (review, comment, reply)
