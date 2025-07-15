# 🔔 Frontend Notification Integration Guide

## Overview
Complete guide for integrating with the ReviewIt notification service. This service handles real-time notifications for replies, likes, and product owner alerts.

**Base URL:** `https://notifications.reviewit.gy`

---

## 🚀 Quick Start

### 1. Create Users First
Before sending notifications, ensure users exist in the notification system:

```javascript
// Create a user in the notification system
const createUser = async (userId, username, fullName) => {
  const response = await fetch('https://notifications.reviewit.gy/users', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      id: userId,        // Your user ID from ReviewIt
      username: username,
      full_name: fullName
    })
  });
  return response.ok;
};
```

### 2. Set Up Real-Time Notifications (SSE)
```javascript
// Connect to notification stream
const connectNotifications = (userId) => {
  const eventSource = new EventSource(
    `https://notifications.reviewit.gy/notifications/stream?user_id=${userId}`
  );
  
  eventSource.onmessage = (event) => {
    const data = JSON.parse(event.data);
    handleNotification(data);
  };
  
  return eventSource;
};

const handleNotification = (data) => {
  if (data.event === 'new_notification') {
    // Show notification to user
    showNotificationToast(data);
  }
};
```

---

## 📨 Sending Notifications

### Comment Notifications (NEW!)
**When:** Someone comments directly on a review

```javascript
const sendCommentNotification = async (commentData) => {
  const response = await fetch('https://notifications.reviewit.gy/notifications/comment', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      id: generateUUID(),                        // Generate unique ID
      review_id: commentData.reviewId,           // Review being commented on
      comment_id: commentData.commentId,         // The new comment ID
      parent_user_id: commentData.reviewAuthorId, // 🔑 DIRECT: Review author's user ID
      from_id: commentData.fromUserId,           // User who commented
      from_name: commentData.fromUserName,       // Name of user who commented
      content: commentData.content,              // Comment content
      product_id: commentData.productId,         // Product context
      parent_id: commentData.reviewId,           // Review ID as parent
      read: false
    })
  });
  return response.ok;
};

// Usage example:
sendCommentNotification({
  reviewId: "review_123",
  commentId: "comment_456", 
  reviewAuthorId: "user_789",    // ← Direct user ID of review author
  fromUserId: "user_101",
  fromUserName: "Jane Smith",
  content: "Great review!",
  productId: "product_202"
});
```

### Reply Notifications
**When:** Someone replies to a comment

```javascript
const sendReplyNotification = async (replyData) => {
  const response = await fetch('https://notifications.reviewit.gy/notifications/reply', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      id: generateUUID(),                      // Generate unique ID
      parent_user_id: replyData.commentAuthorId, // 🔑 DIRECT: Comment author's user ID (recommended)
      content: replyData.content,              // Reply content
      comment_id: replyData.commentId,         // The new reply comment ID
      from_id: replyData.fromUserId,           // User who replied
      review_id: replyData.reviewId,           // Review context
      parent_id: replyData.parentCommentId,    // Parent comment ID
      from_name: replyData.fromUserName,       // Name of user who replied
      product_id: replyData.productId,         // Product context
      read: false
    })
  });
  return response.ok;
};

// Usage examples:
// Option 1: You have the comment author's user ID (recommended)
sendReplyNotification({
  commentId: "reply_123",
  commentAuthorId: "user_789",           // ← Direct user ID of comment author
  content: "Thanks for your comment!",
  fromUserId: "user_456", 
  reviewId: "review_789",
  parentCommentId: "comment_parent_111", // ← Parent comment ID
  fromUserName: "John Doe",
  productId: "product_222"
});

// Option 2: Let the system look up the user ID from comment ID (fallback)
sendReplyNotification({
  commentId: "reply_123",
  // commentAuthorId: undefined,         // ← Not provided
  content: "Thanks for your comment!",
  fromUserId: "user_456",
  reviewId: "review_789", 
  parentCommentId: "comment_parent_111", // ← System will look up user from this
  fromUserName: "John Doe",
  productId: "product_222"
});
```

### Like Notifications
**When:** Someone likes a comment or review

```javascript
const sendLikeNotification = async (likeData) => {
  const response = await fetch('https://notifications.reviewit.gy/notifications/like', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      target_type: likeData.targetType,      // "comment" or "review"
      target_id: likeData.targetId,          // 🔑 FLEXIBLE: Content ID OR User ID
      from_id: likeData.fromUserId,          // User who liked
      from_name: likeData.fromUserName,      // Name of user who liked
      product_id: likeData.productId,        // Product context
      read: false
    })
  });
  return response.ok;
};

// Usage examples:
// Option 1: You have the content ID (preferred)
sendLikeNotification({
  targetType: "comment",
  targetId: "comment_123",        // ← Comment ID
  fromUserId: "user_456",
  fromUserName: "Jane Smith",
  productId: "product_789"
});

// Option 2: You only know the target user ID (also works)
sendLikeNotification({
  targetType: "comment", 
  targetId: "user_owner_999",     // ← User ID (starts with "user_")
  fromUserId: "user_456",
  fromUserName: "Jane Smith",
  productId: "product_789"
});
```

### Product Owner Notifications
**When:** Someone reviews a product

```javascript
const sendProductOwnerNotification = async (reviewData) => {
  const response = await fetch('https://notifications.reviewit.gy/notifications/product-owner', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      id: generateUUID(),
      owner_id: reviewData.ownerId,          // Product owner user ID
      product_id: reviewData.productId,
      product_name: reviewData.productName,
      business_id: reviewData.businessId,
      review_title: reviewData.reviewTitle,
      from_name: reviewData.fromUserName,    // Reviewer name
      from_id: reviewData.fromUserId,        // Reviewer user ID
      read: false,
      comment_id: reviewData.commentId,      // Optional
      review_id: reviewData.reviewId         // Optional
    })
  });
  return response.ok;
};
```

---

## 📖 Getting Notifications

### Get All Notifications
```javascript
const getAllNotifications = async (userId) => {
  const response = await fetch(
    `https://notifications.reviewit.gy/notifications?user_id=${userId}`
  );
  const data = await response.json();
  
  return {
    userNotifications: data.user_notifications,      // Replies
    ownerNotifications: data.owner_notifications,    // Product owner alerts
    likeNotifications: data.like_notifications       // Likes
  };
};
```

### Get Unread Notifications
```javascript
const getUnreadNotifications = async (userId) => {
  const response = await fetch(
    `https://notifications.reviewit.gy/notifications/unread?user_id=${userId}`
  );
  return await response.json();
};
```

### Mark Notification as Read
```javascript
const markAsRead = async (notificationId, type) => {
  const response = await fetch(
    `https://notifications.reviewit.gy/notifications/${notificationId}/read?type=${type}`,
    { method: 'PUT' }
  );
  return response.ok;
};

// Usage:
markAsRead("notif_123", "user");   // For reply notifications
markAsRead("notif_456", "owner");  // For product owner notifications  
markAsRead("notif_789", "like");   // For like notifications
```

---

## 🎯 Data You Need to Collect

### For Reply Notifications:
- ✅ **Comment ID** of the new reply
- ✅ **User ID** of who replied
- ✅ **User Name** of who replied
- ✅ **Review ID** for context
- ✅ **Product ID** for context
- ⚠️ **Parent Comment ID** OR **Target User ID** (either works!)

### For Like Notifications:
- ✅ **Target Type** ("comment" or "review")
- ✅ **User ID** of who liked
- ✅ **User Name** of who liked
- ✅ **Product ID** for context
- ⚠️ **Content ID** OR **Target User ID** (either works!)

### For Product Owner Notifications:
- ✅ **Product Owner User ID**
- ✅ **Product ID** and **Product Name**
- ✅ **Business ID**
- ✅ **Review Title**
- ✅ **Reviewer User ID** and **Name**

---

## 🔧 Helper Functions

### Generate UUID
```javascript
const generateUUID = () => {
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function(c) {
    const r = Math.random() * 16 | 0;
    const v = c == 'x' ? r : (r & 0x3 | 0x8);
    return v.toString(16);
  });
};
```

### Complete Notification Handler
```javascript
class NotificationManager {
  constructor(userId) {
    this.userId = userId;
    this.eventSource = null;
  }
  
  connect() {
    this.eventSource = new EventSource(
      `https://notifications.reviewit.gy/notifications/stream?user_id=${this.userId}`
    );
    
    this.eventSource.onmessage = (event) => {
      const data = JSON.parse(event.data);
      this.handleNotification(data);
    };
    
    this.eventSource.onerror = (error) => {
      console.error('Notification connection error:', error);
      // Implement reconnection logic
    };
  }
  
  handleNotification(data) {
    switch(data.type) {
      case 'user':
        this.showReplyNotification(data.notification);
        break;
      case 'owner':
        this.showOwnerNotification(data.notification);
        break;
      case 'like':
        this.showLikeNotification(data.notification);
        break;
    }
  }
  
  showReplyNotification(notification) {
    this.showToast({
      title: `${notification.from_name} replied`,
      message: notification.content,
      type: 'reply',
      id: notification.id
    });
  }
  
  showLikeNotification(notification) {
    this.showToast({
      title: `${notification.from_name} liked your ${notification.target_type}`,
      message: `Someone appreciated your ${notification.target_type}!`,
      type: 'like',
      id: notification.id
    });
  }
  
  showOwnerNotification(notification) {
    this.showToast({
      title: `New review for ${notification.product_name}`,
      message: `${notification.from_name} reviewed your product`,
      type: 'owner',
      id: notification.id
    });
  }
  
  showToast(notification) {
    // Your toast implementation
    console.log('New notification:', notification);
  }
  
  disconnect() {
    if (this.eventSource) {
      this.eventSource.close();
    }
  }
}

// Usage:
const notifications = new NotificationManager('user_123');
notifications.connect();
```

---

## 🚨 Error Handling

### Common Errors:
- **400**: Missing required fields or invalid data
- **404**: User doesn't exist (create user first)
- **500**: Server error (check logs)

### Retry Logic:
```javascript
const sendNotificationWithRetry = async (url, data, maxRetries = 3) => {
  for (let i = 0; i < maxRetries; i++) {
    try {
      const response = await fetch(url, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(data)
      });
      
      if (response.ok) return true;
      
      if (response.status === 400) {
        console.error('Bad request:', await response.text());
        return false; // Don't retry 400 errors
      }
      
    } catch (error) {
      console.error(`Attempt ${i + 1} failed:`, error);
      if (i === maxRetries - 1) return false;
      await new Promise(resolve => setTimeout(resolve, 1000 * (i + 1)));
    }
  }
  return false;
};
```

---

## 🎉 Complete Integration Example

```javascript
// 1. Initialize when user logs in
const initNotifications = async (user) => {
  // Create user in notification system
  await createUser(user.id, user.username, user.fullName);
  
  // Connect to real-time notifications
  const notificationManager = new NotificationManager(user.id);
  notificationManager.connect();
  
  return notificationManager;
};

// 2. When someone replies to a comment
const onCommentReply = async (replyData) => {
  await sendReplyNotification({
    commentId: replyData.id,
    fromUserId: currentUser.id,
    reviewId: replyData.reviewId,
    parentCommentId: replyData.parentId, // Can be comment ID or user ID
    fromUserName: currentUser.name,
    productId: replyData.productId
  });
};

// 3. When someone likes content
const onContentLike = async (likeData) => {
  await sendLikeNotification({
    targetType: likeData.type,
    targetId: likeData.contentId, // Can be content ID or user ID
    fromUserId: currentUser.id,
    fromUserName: currentUser.name,
    productId: likeData.productId
  });
};

// 4. When someone reviews a product
const onProductReview = async (reviewData) => {
  await sendProductOwnerNotification({
    ownerId: reviewData.productOwnerId,
    productId: reviewData.productId,
    productName: reviewData.productName,
    businessId: reviewData.businessId,
    reviewTitle: reviewData.title,
    fromUserName: currentUser.name,
    fromUserId: currentUser.id,
    reviewId: reviewData.id
  });
};
```

---

## 🔑 Key Points

1. **Flexible ID Handling**: You can send either content IDs or user IDs - the service handles both
2. **Create Users First**: Always ensure users exist before sending notifications
3. **Real-Time**: Use SSE for instant notifications
4. **Error Handling**: Implement retry logic for network issues
5. **User Experience**: Show appropriate toast/popup notifications based on type

**Questions?** Check the logs or contact the backend team! 🚀