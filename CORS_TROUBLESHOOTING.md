# 🔧 CORS Troubleshooting Guide

## ✅ Fixed CORS Configuration

The notification service is now configured to work with your production domain `https://reviewit.gy`.

### Current CORS Settings:
```go
app.Use(cors.New(cors.Config{
    AllowOrigins:     "https://reviewit.gy,http://localhost:3000,http://localhost:3001,http://127.0.0.1:3000",
    AllowMethods:     "GET,POST,HEAD,PUT,DELETE,PATCH,OPTIONS",
    AllowHeaders:     "Origin, Content-Type, Accept, Cache-Control, Authorization, X-Requested-With",
    AllowCredentials: true,
    ExposeHeaders:    "Content-Length, Content-Type",
}))
```

## 🚀 Testing from https://reviewit.gy

### 1. Regular API Calls:
```javascript
// This should now work from https://reviewit.gy
fetch('http://your-noti-service-domain:3001/notifications?user_id=test123')
  .then(response => response.json())
  .then(data => console.log(data));
```

### 2. SSE Connection:
```javascript
// This should also work now
const eventSource = new EventSource(
  'http://your-noti-service-domain:3001/notifications/stream?user_id=test123'
);

eventSource.onmessage = function(event) {
  console.log('Notification:', JSON.parse(event.data));
};
```

## 🔍 If You Still Get CORS Errors:

### Check 1: Server Domain
Make sure you're calling the correct server URL. If your notification service is deployed, update the URL:

```javascript
// Replace with your actual notification service URL
const BASE_URL = 'https://your-noti-service.com'; // or whatever domain you're using
const eventSource = new EventSource(`${BASE_URL}/notifications/stream?user_id=${userId}`);
```

### Check 2: Add More Domains (if needed)
If you have multiple frontend domains, add them to the CORS config:

```go
AllowOrigins: "https://reviewit.gy,https://www.reviewit.gy,https://staging.reviewit.gy,http://localhost:3000",
```

### Check 3: Browser Developer Tools
1. Open browser dev tools (F12)
2. Go to Network tab
3. Try the request
4. Look for:
   - ❌ Red CORS error in console
   - ❌ OPTIONS preflight request failing
   - ✅ Successful OPTIONS response with proper headers

### Check 4: Preflight Request
For complex requests, browsers send an OPTIONS preflight. Make sure it succeeds:

```bash
# Test preflight manually
curl -X OPTIONS \
  -H "Origin: https://reviewit.gy" \
  -H "Access-Control-Request-Method: GET" \
  -H "Access-Control-Request-Headers: Content-Type" \
  http://your-server:3001/notifications/stream
```

## 🛠️ Quick Debug Steps:

### 1. Test from Same Domain (should work):
```javascript
// If your notification service is on the same domain as reviewit.gy
const eventSource = new EventSource('/notifications/stream?user_id=test123');
```

### 2. Test with Simple Fetch First:
```javascript
// Test basic API call before trying SSE
fetch('http://your-noti-service:3001/notifications?user_id=test123', {
  method: 'GET',
  headers: {
    'Content-Type': 'application/json',
  }
})
.then(response => {
  console.log('CORS working!', response);
  return response.json();
})
.then(data => console.log(data))
.catch(error => console.error('CORS Error:', error));
```

### 3. Check Server Logs:
Look for CORS-related log messages when making requests from your frontend.

## 🔧 Alternative Solutions:

### Option 1: Proxy Through Your Backend
If CORS continues to be problematic, proxy the requests through your main reviewit.gy backend:

```javascript
// Instead of calling notification service directly
fetch('https://reviewit.gy/api/notifications/stream?user_id=test123')

// Your reviewit.gy backend then forwards to notification service
```

### Option 2: Use Server-Sent Events with Credentials
```javascript
const eventSource = new EventSource(
  'http://your-noti-service:3001/notifications/stream?user_id=test123',
  { withCredentials: true }  // This requires AllowCredentials: true (which we set)
);
```

## 📝 Need More Help?

If you're still getting CORS errors, please share:
1. The exact error message from browser console
2. Your notification service deployment URL
3. The specific frontend code that's failing

The current configuration should work for `https://reviewit.gy` → notification service calls!