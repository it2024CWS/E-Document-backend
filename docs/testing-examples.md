# API Testing Examples

## Testing with cURL

### Web App Login (No X-Client-Type header)

```bash
# Login as web client
curl -X POST http://localhost:5000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "usernameOrEmail": "admin@edocument.com",
    "password": "password"
  }' \
  -c cookies.txt \
  -v

# Response will NOT include tokens in body (only cookies):
# {
#   "success": true,
#   "message": "Login successful",
#   "data": {
#     "user": { ... }
#   }
# }

# Make authenticated request using cookies
curl -X GET http://localhost:5000/api/v1/users \
  -b cookies.txt
```

### Mobile App Login (With X-Client-Type: mobile header)

```bash
# Login as mobile client
curl -X POST http://localhost:5000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -H "X-Client-Type: mobile" \
  -d '{
    "usernameOrEmail": "admin@edocument.com",
    "password": "password"
  }' \
  -v

# Response WILL include tokens in body:
# {
#   "success": true,
#   "message": "Login successful",
#   "data": {
#     "user": { ... },
#     "accessToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
#     "refreshToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
#   }
# }

# Save the tokens and use in subsequent requests
ACCESS_TOKEN="your_access_token_here"

# Make authenticated request with Bearer token
curl -X GET http://localhost:5000/api/v1/users \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

### Refresh Token (Web)

```bash
# Refresh using cookies only
curl -X POST http://localhost:5000/api/v1/auth/refresh \
  -b cookies.txt \
  -c cookies.txt \
  -v

# Response will NOT include tokens in body (only update cookies)
```

### Refresh Token (Mobile)

```bash
REFRESH_TOKEN="your_refresh_token_here"

# Refresh with explicit token and mobile header
curl -X POST http://localhost:5000/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -H "X-Client-Type: mobile" \
  -d "{
    \"refreshToken\": \"$REFRESH_TOKEN\"
  }"

# Response WILL include new tokens in body
```

## Testing with Postman

### Setup Environment Variables

Create a Postman environment with these variables:
- `baseUrl`: `http://localhost:5000/api`
- `accessToken`: (will be set automatically)
- `refreshToken`: (will be set automatically)

### Collection: Web App Flow

#### 1. Login (Web)
```
POST {{baseUrl}}/v1/auth/login
Headers:
  Content-Type: application/json

Body (raw JSON):
{
  "usernameOrEmail": "admin@edocument.com",
  "password": "password"
}

Tests (to verify NO tokens in response):
pm.test("No access token in response body", function() {
    pm.expect(pm.response.json().data.accessToken).to.be.undefined;
});
pm.test("No refresh token in response body", function() {
    pm.expect(pm.response.json().data.refreshToken).to.be.undefined;
});
```

#### 2. Get Users (Web - uses cookies automatically)
```
GET {{baseUrl}}/v1/users
(Cookies sent automatically by Postman)
```

### Collection: Mobile App Flow

#### 1. Login (Mobile)
```
POST {{baseUrl}}/v1/auth/login
Headers:
  Content-Type: application/json
  X-Client-Type: mobile

Body (raw JSON):
{
  "usernameOrEmail": "admin@edocument.com",
  "password": "password"
}

Tests (to save tokens):
pm.test("Status code is 200", function() {
    pm.response.to.have.status(200);
});

var jsonData = pm.response.json();
pm.test("Has access token", function() {
    pm.expect(jsonData.data.accessToken).to.exist;
});
pm.test("Has refresh token", function() {
    pm.expect(jsonData.data.refreshToken).to.exist;
});

// Save tokens to environment
pm.environment.set("accessToken", jsonData.data.accessToken);
pm.environment.set("refreshToken", jsonData.data.refreshToken);
```

#### 2. Get Users (Mobile)
```
GET {{baseUrl}}/v1/users
Headers:
  Authorization: Bearer {{accessToken}}

Tests:
pm.test("Status code is 200", function() {
    pm.response.to.have.status(200);
});
pm.test("Returns user list", function() {
    var jsonData = pm.response.json();
    pm.expect(jsonData.success).to.be.true;
    pm.expect(jsonData.data).to.be.an('array');
});
```

#### 3. Refresh Token (Mobile)
```
POST {{baseUrl}}/v1/auth/refresh
Headers:
  Content-Type: application/json
  X-Client-Type: mobile

Body (raw JSON):
{
  "refreshToken": "{{refreshToken}}"
}

Tests (to update tokens):
var jsonData = pm.response.json();
if (jsonData.success) {
    pm.environment.set("accessToken", jsonData.data.accessToken);
    pm.environment.set("refreshToken", jsonData.data.refreshToken);
}
```

#### 4. Get Profile (Mobile)
```
GET {{baseUrl}}/v1/auth/profile
Headers:
  Authorization: Bearer {{accessToken}}
```

#### 5. Logout (Mobile)
```
POST {{baseUrl}}/v1/auth/logout

Tests (to clear tokens):
pm.environment.unset("accessToken");
pm.environment.unset("refreshToken");
```

## Testing with HTTPie

### Web Login
```bash
http POST http://localhost:5000/api/v1/auth/login \
  usernameOrEmail=admin@edocument.com \
  password=password \
  --session=web
```

### Mobile Login
```bash
http POST http://localhost:5000/api/v1/auth/login \
  usernameOrEmail=admin@edocument.com \
  password=password \
  X-Client-Type:mobile
```

### Make Authenticated Request (Mobile)
```bash
TOKEN="your_access_token"
http GET http://localhost:5000/api/v1/users \
  "Authorization: Bearer $TOKEN"
```

## Expected Behavior Summary

### Web Client (No X-Client-Type header)
| Endpoint | Cookies Set? | Tokens in Body? |
|----------|-------------|-----------------|
| POST /auth/login | ✅ Yes | ❌ No |
| POST /auth/refresh | ✅ Yes (updated) | ❌ No |
| GET /auth/profile | - | - |
| GET /users | - | - |

### Mobile Client (X-Client-Type: mobile)
| Endpoint | Cookies Set? | Tokens in Body? |
|----------|-------------|-----------------|
| POST /auth/login | ✅ Yes (optional) | ✅ Yes |
| POST /auth/refresh | ✅ Yes (optional) | ✅ Yes |
| GET /auth/profile | - | - |
| GET /users | - | - |

## Security Verification

### Test 1: Web client should NOT get tokens in response

```bash
RESPONSE=$(curl -s -X POST http://localhost:5000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"usernameOrEmail":"admin","password":"password"}')

# Should output: NOT FOUND (secure)
echo $RESPONSE | grep -o "accessToken" || echo "✅ SECURE: No accessToken in web response"
```

### Test 2: Mobile client SHOULD get tokens in response

```bash
RESPONSE=$(curl -s -X POST http://localhost:5000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -H "X-Client-Type: mobile" \
  -d '{"usernameOrEmail":"admin","password":"password"}')

# Should output: FOUND
echo $RESPONSE | grep -o "accessToken" && echo "✅ CORRECT: accessToken in mobile response"
```

### Test 3: Bearer token authentication works

```bash
# Login and extract token
TOKEN=$(curl -s -X POST http://localhost:5000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -H "X-Client-Type: mobile" \
  -d '{"usernameOrEmail":"admin","password":"password"}' \
  | jq -r '.data.accessToken')

# Use token to access protected endpoint
curl -X GET http://localhost:5000/api/v1/users \
  -H "Authorization: Bearer $TOKEN"

# Should return user list
```

### Test 4: Missing X-Client-Type returns user data without tokens

```bash
curl -X POST http://localhost:5000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"usernameOrEmail":"admin","password":"password"}' \
  | jq '.data | has("accessToken")'

# Should output: false
```

## Common Issues & Solutions

### Issue 1: Not receiving tokens in mobile app

**Cause:** Missing `X-Client-Type: mobile` header

**Solution:**
```javascript
// ❌ Wrong
fetch('/api/v1/auth/login', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify(credentials)
})

// ✅ Correct
fetch('/api/v1/auth/login', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'X-Client-Type': 'mobile'  // Add this!
  },
  body: JSON.stringify(credentials)
})
```

### Issue 2: Web app can access tokens via JavaScript

**Cause:** This is actually prevented! The API doesn't return tokens in body for web clients.

**Verification:**
```javascript
// This will NOT work (tokens not in response body)
fetch('/api/v1/auth/login', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify(credentials)
})
.then(r => r.json())
.then(data => {
  console.log(data.data.accessToken); // undefined - SECURE!
});

// Tokens are only in httpOnly cookies (JavaScript cannot access)
```

### Issue 3: 401 Unauthorized on requests

**Cause:** Token expired or not sent correctly

**Solution for Mobile:**
```javascript
// Make sure to include Authorization header
fetch('/api/v1/users', {
  headers: {
    'Authorization': `Bearer ${accessToken}`
  }
})
```

**Solution for Web:**
```javascript
// Make sure to include credentials
fetch('/api/v1/users', {
  credentials: 'include'  // Sends cookies
})
```

---

**Last Updated:** 2024-01-15
