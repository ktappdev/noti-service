# 🔔 Frontend Like Notification Integration Specification

## 🎯 Overview
This specification covers the complete integration of like notifications into your existing notification system. Like notifications are now fully integrated with all existing notification endpoints and SSE streams.

## 📋 What Changed in the Backend

### ✅ Completed Backend Updates
1. **GetAllNotifications** - Now includes `like_notifications` in response
2. **GetAllUnreadNotifications** - Now includes unread like notifications
3. **MarkNotificationAsRead** - Now supports `type: "like"`
4. **DeleteReadNotifications** - Now deletes read like notifications
5. **CreateLikeNotification** - Already working (no changes needed)

## 🔌 API Endpoint Changes

### 1. Get All Notifications
**Endpoint:** `GET /notifications/all?user_id={user_id}`

**NEW Response Format:**
```json
{
  "user_notifications": [
    {
      "id": "notif_123",
      "parent_user_id": "user_456",
      "content": "John replied to your comment",
      "notification_type": "reply",
      "created_at": "2024-01-15T10:30:00Z",
      "read": false,
      "comment_id": "comment_789",
      "from_id": "user_john",
      "from_name": "John Doe"
    }
  ],
  "owner_notifications": [
    {
      "id": "owner_notif_456",
      "owner_id": "user_456",
      "product_id": "prod_123",
      "notification_type": "review",
      "created_at": "2024-01-15T09:15:00Z",
      "read": false
    }
  ],
  "like_notifications": [
    {
      "id": "like_notif_789",
      "target_user_id": "user_456",
      "target_type": "comment",
      "target_id": "comment_123",
      "from_id": "user_john",
      "from_name": "John Doe",
      "product_id": "prod_789",
      "created_at": "2024-01-15T11:45:00Z",
      "read": false
    }
  ]
}
```

### 2. Get Unread Notifications
**Endpoint:** `GET /notifications/unread?user_id={user_id}`

**NEW Response Format:** Same structure as above, but only unread notifications

### 3. Mark Notification as Read
**Endpoint:** `PUT /notifications/{notification_id}/read?type={type}`

**NEW Supported Types:**
- `type=user` (existing)
- `type=owner` (existing)
- `type=like` ⭐ **NEW**

**Example:**
```javascript
// Mark like notification as read
fetch('/notifications/like_notif_789/read?type=like', { method: 'PUT' })
```

### 4. Delete Read Notifications
**Endpoint:** `DELETE /notifications/read?user_id={user_id}`

**NEW Response Format:**
```json
{
  "deleted_user_notifications": 5,
  "deleted_owner_notifications": 2,
  "deleted_like_notifications": 8,
  "user_notifications": [...],
  "owner_notifications": [...],
  "like_notifications": [...]
}
```

## 🔄 SSE Integration

### Like Notification SSE Message
When someone likes your content, you receive:

```json
{
  "user_id": "user_456",
  "type": "like",
  "event": "new_notification",
  "notification": {
    "id": "like_notif_789",
    "target_user_id": "user_456",
    "target_type": "comment",
    "target_id": "comment_123",
    "from_id": "user_john",
    "from_name": "John Doe",
    "product_id": "prod_789",
    "created_at": "2024-01-15T11:45:00Z",
    "read": false
  }
}
```

### Read Status SSE Message
When a like notification is marked as read:

```json
{
  "user_id": "user_456",
  "type": "like",
  "event": "notification_read",
  "notification": {
    "notification_id": "like_notif_789",
    "type": "like",
    "read": true,
    "timestamp": "2024-01-15T11:50:00Z"
  }
}
```

## 💻 Frontend Implementation Guide

### 1. Update Notification Fetching

```typescript
interface LikeNotification {
  id: string;
  target_user_id: string;
  target_type: 'comment' | 'review';
  target_id: string;
  from_id: string;
  from_name: string;
  product_id?: string;
  created_at: string;
  read: boolean;
}

interface NotificationResponse {
  user_notifications: UserNotification[];
  owner_notifications: OwnerNotification[];
  like_notifications: LikeNotification[]; // NEW!
}

// Updated fetch function
const fetchAllNotifications = async (userId: string): Promise<NotificationResponse> => {
  const response = await fetch(`/notifications/all?user_id=${userId}`);
  return response.json();
};
```

### 2. Update SSE Event Handling

```typescript
// Add to your existing SSE handler
eventSource.addEventListener('message', (event) => {
  const data = JSON.parse(event.data);
  
  // Handle new like notifications
  if (data.type === 'like' && data.event === 'new_notification') {
    const notification = data.notification as LikeNotification;
    
    // Add to your notification state
    setLikeNotifications(prev => [notification, ...prev]);
    
    // Show toast/popup
    showNotification({
      title: `${notification.from_name} liked your ${notification.target_type}`,
      message: `Someone appreciated your ${notification.target_type}!`,
      type: 'like',
      icon: '👍',
      timestamp: notification.created_at
    });
  }
  
  // Handle like notification read status
  if (data.type === 'like' && data.event === 'notification_read') {
    const readInfo = data.notification;
    setLikeNotifications(prev => 
      prev.map(notif => 
        notif.id === readInfo.notification_id 
          ? { ...notif, read: true }
          : notif
      )
    );
  }
});
```

### 3. Update Notification Display Components

```typescript
// Add like notification rendering
const NotificationList = ({ notifications }: { notifications: NotificationResponse }) => {
  // Combine all notifications for unified display
  const allNotifications = [
    ...notifications.user_notifications.map(n => ({ ...n, type: 'user' })),
    ...notifications.owner_notifications.map(n => ({ ...n, type: 'owner' })),
    ...notifications.like_notifications.map(n => ({ ...n, type: 'like' }))
  ].sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime());

  return (
    <div className="notification-list">
      {allNotifications.map(notification => (
        <NotificationItem 
          key={notification.id} 
          notification={notification}
          onMarkAsRead={handleMarkAsRead}
        />
      ))}
    </div>
  );
};

// Updated notification item component
const NotificationItem = ({ notification, onMarkAsRead }) => {
  const getNotificationContent = () => {
    switch (notification.type) {
      case 'like':
        return {
          icon: '👍',
          title: `${notification.from_name} liked your ${notification.target_type}`,
          message: `Someone appreciated your ${notification.target_type}!`,
          color: '#ff6b6b'
        };
      case 'user':
        return {
          icon: '💬',
          title: notification.content,
          message: `From ${notification.from_name}`,
          color: '#4ecdc4'
        };
      case 'owner':
        return {
          icon: '⭐',
          title: `New review: ${notification.review_title}`,
          message: `From ${notification.from_name}`,
          color: '#45b7d1'
        };
    }
  };

  const content = getNotificationContent();

  return (
    <div 
      className={`notification-item ${!notification.read ? 'unread' : ''}`}
      onClick={() => onMarkAsRead(notification.id, notification.type)}
    >
      <div className="notification-icon" style={{ color: content.color }}>
        {content.icon}
      </div>
      <div className="notification-content">
        <h4>{content.title}</h4>
        <p>{content.message}</p>
        <span className="timestamp">{formatTimestamp(notification.created_at)}</span>
      </div>
      {!notification.read && <div className="unread-indicator" />}
    </div>
  );
};
```

### 4. Update Mark as Read Function

```typescript
const markNotificationAsRead = async (notificationId: string, type: 'user' | 'owner' | 'like') => {
  try {
    await fetch(`/notifications/${notificationId}/read?type=${type}`, {
      method: 'PUT'
    });
    
    // Update local state based on type
    switch (type) {
      case 'like':
        setLikeNotifications(prev => 
          prev.map(n => n.id === notificationId ? { ...n, read: true } : n)
        );
        break;
      case 'user':
        setUserNotifications(prev => 
          prev.map(n => n.id === notificationId ? { ...n, read: true } : n)
        );
        break;
      case 'owner':
        setOwnerNotifications(prev => 
          prev.map(n => n.id === notificationId ? { ...n, read: true } : n)
        );
        break;
    }
  } catch (error) {
    console.error('Failed to mark notification as read:', error);
  }
};
```

### 5. Update Notification Counting

```typescript
const getUnreadCount = (notifications: NotificationResponse): number => {
  const userUnread = notifications.user_notifications.filter(n => !n.read).length;
  const ownerUnread = notifications.owner_notifications.filter(n => !n.read).length;
  const likeUnread = notifications.like_notifications.filter(n => !n.read).length; // NEW!
  
  return userUnread + ownerUnread + likeUnread;
};

// Update your notification bell badge
const NotificationBell = () => {
  const [notifications, setNotifications] = useState<NotificationResponse>({
    user_notifications: [],
    owner_notifications: [],
    like_notifications: [] // NEW!
  });

  const unreadCount = getUnreadCount(notifications);

  return (
    <div className="notification-bell">
      <BellIcon />
      {unreadCount > 0 && (
        <span className="notification-badge">{unreadCount}</span>
      )}
    </div>
  );
};
```

## 🎨 Visual Design Recommendations

### Like Notification Styling
```css
.notification-item.like {
  border-left: 4px solid #ff6b6b;
}

.notification-icon.like {
  background: linear-gradient(135deg, #ff6b6b, #ff8e8e);
  color: white;
}

.notification-item.like:hover {
  background-color: #fff5f5;
}
```

### Notification Type Icons
- **Like notifications**: 👍 (thumbs up)
- **Reply notifications**: 💬 (speech bubble)
- **Review notifications**: ⭐ (star)
- **Comment notifications**: 💭 (thought bubble)

## 🧪 Testing Checklist

### ✅ Frontend Testing Tasks
1. **Fetch notifications** - Verify `like_notifications` array is present
2. **SSE like events** - Test receiving new like notifications via SSE
3. **Mark like as read** - Test marking like notifications as read
4. **Notification counting** - Verify like notifications are included in unread count
5. **Visual display** - Ensure like notifications render with proper styling
6. **Delete read notifications** - Test bulk deletion includes like notifications
7. **Real-time updates** - Test SSE read status updates for like notifications

### Test Scenarios
```javascript
// Test 1: Create and receive like notification
await createLikeNotification('comment', 'comment_123', 'user_456', 'John Doe', 'prod_789');
// Verify: SSE message received, notification appears in UI, unread count increases

// Test 2: Mark like notification as read
await markNotificationAsRead('like_notif_789', 'like');
// Verify: Notification marked as read, unread count decreases, SSE read event received

// Test 3: Fetch all notifications
const notifications = await fetchAllNotifications('user_123');
// Verify: Response includes like_notifications array

// Test 4: Delete read notifications
await deleteReadNotifications('user_123');
// Verify: Read like notifications are deleted
```

## 🚀 Migration Steps

1. **Update API calls** - Modify existing notification fetching to handle new response format
2. **Update SSE handlers** - Add like notification event handling
3. **Update UI components** - Add like notification rendering and styling
4. **Update state management** - Include like notifications in your state
5. **Test thoroughly** - Verify all notification flows work with like notifications
6. **Deploy incrementally** - Test in staging before production

## 📞 Support

If you encounter any issues:
1. Check that like notifications appear in API responses
2. Verify SSE events are being received for type "like"
3. Ensure mark-as-read calls include `type=like` parameter
4. Test notification counting includes all three types

The backend now fully supports like notifications integrated with your existing notification system! 🎉