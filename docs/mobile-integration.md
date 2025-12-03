# Mobile App Integration Guide

## Authentication Flow for Mobile Apps

This API uses a **Unified Authentication Approach** that supports both web browsers (cookies) and mobile apps (token-based).

## Key Points

- ✅ **Same endpoints** for both web and mobile
- ✅ **Tokens returned in response body ONLY for mobile** (when sending `X-Client-Type: mobile` header)
- ✅ **Cookies set automatically** for web browsers
- ✅ **Secure by default** - web apps never expose tokens to JavaScript (XSS protection)
- ✅ **Flexible authentication** - use either Bearer token or cookies

## Authentication Endpoints

### 1. Login

**Endpoint:** `POST /api/v1/auth/login`

**Headers (Required for Mobile):**
```
X-Client-Type: mobile
Content-Type: application/json
```

**Request Body:**
```json
{
  "usernameOrEmail": "admin@edocument.com",
  "password": "password123"
}
```

**Response (with X-Client-Type: mobile):**
```json
{
  "success": true,
  "message": "Login successful",
  "data": {
    "user": {
      "id": "507f1f77bcf86cd799439011",
      "username": "admin",
      "email": "admin@edocument.com",
      "role": "Director",
      "phone": "+66812345678",
      "first_name": "John",
      "last_name": "Doe",
      "department_id": "dept123",
      "sector_id": "sector456",
      "created_at": "2024-01-15T10:30:00Z",
      "updated_at": "2024-01-15T10:30:00Z"
    },
    "accessToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refreshToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }
}
```

**Response (without X-Client-Type header - for web):**
```json
{
  "success": true,
  "message": "Login successful",
  "data": {
    "user": {
      "id": "507f1f77bcf86cd799439011",
      "username": "admin",
      "email": "admin@edocument.com",
      ...
    }
    // Note: accessToken and refreshToken are NOT included in response
    // Tokens are only available via httpOnly cookies for security
  }
}
```

### 2. Refresh Token

**Endpoint:** `POST /api/v1/auth/refresh`

**Headers (Required for Mobile):**
```
X-Client-Type: mobile
Content-Type: application/json
```

**Request Body:**
```json
{
  "refreshToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Response (with X-Client-Type: mobile):**
```json
{
  "success": true,
  "message": "Token refreshed successfully",
  "data": {
    "user": { ... },
    "accessToken": "new_access_token...",
    "refreshToken": "new_refresh_token..."
  }
}
```

**Response (without X-Client-Type header - for web):**
```json
{
  "success": true,
  "message": "Token refreshed successfully",
  "data": {
    "user": { ... }
    // Tokens only in httpOnly cookies
  }
}
```

### 3. Get Profile

**Endpoint:** `GET /api/v1/auth/profile`

**Headers:**
```
Authorization: Bearer <accessToken>
```

**Response:**
```json
{
  "success": true,
  "message": "Profile retrieved successfully",
  "data": {
    "id": "507f1f77bcf86cd799439011",
    "username": "admin",
    "email": "admin@edocument.com",
    ...
  }
}
```

### 4. Logout

**Endpoint:** `POST /api/v1/auth/logout`

**Response:**
```json
{
  "success": true,
  "message": "Logged out successfully",
  "data": null
}
```

## Mobile Implementation Guide

### Step 1: Store Tokens Securely

After successful login, store tokens in secure storage:

**iOS (Swift):**
```swift
import Security

class TokenManager {
    func saveToken(token: String, key: String) {
        let data = token.data(using: .utf8)!
        let query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrAccount as String: key,
            kSecValueData as String: data
        ]
        SecItemDelete(query as CFDictionary)
        SecItemAdd(query as CFDictionary, nil)
    }

    func getToken(key: String) -> String? {
        let query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrAccount as String: key,
            kSecReturnData as String: true
        ]
        var result: AnyObject?
        let status = SecItemCopyMatching(query as CFDictionary, &result)

        guard status == errSecSuccess,
              let data = result as? Data,
              let token = String(data: data, encoding: .utf8) else {
            return nil
        }
        return token
    }
}

// Usage
let tokenManager = TokenManager()
tokenManager.saveToken(token: accessToken, key: "accessToken")
tokenManager.saveToken(token: refreshToken, key: "refreshToken")
```

**Android (Kotlin):**
```kotlin
import androidx.security.crypto.EncryptedSharedPreferences
import androidx.security.crypto.MasterKey

class TokenManager(context: Context) {
    private val masterKey = MasterKey.Builder(context)
        .setKeyScheme(MasterKey.KeyScheme.AES256_GCM)
        .build()

    private val sharedPreferences = EncryptedSharedPreferences.create(
        context,
        "secure_prefs",
        masterKey,
        EncryptedSharedPreferences.PrefKeyEncryptionScheme.AES256_SIV,
        EncryptedSharedPreferences.PrefValueEncryptionScheme.AES256_GCM
    )

    fun saveToken(token: String, key: String) {
        sharedPreferences.edit()
            .putString(key, token)
            .apply()
    }

    fun getToken(key: String): String? {
        return sharedPreferences.getString(key, null)
    }

    fun clearTokens() {
        sharedPreferences.edit()
            .remove("accessToken")
            .remove("refreshToken")
            .apply()
    }
}

// Usage
val tokenManager = TokenManager(context)
tokenManager.saveToken(accessToken, "accessToken")
tokenManager.saveToken(refreshToken, "refreshToken")
```

**React Native:**
```javascript
import * as SecureStore from 'expo-secure-store';

// Save tokens
await SecureStore.setItemAsync('accessToken', accessToken);
await SecureStore.setItemAsync('refreshToken', refreshToken);

// Get tokens
const accessToken = await SecureStore.getItemAsync('accessToken');
const refreshToken = await SecureStore.getItemAsync('refreshToken');

// Delete tokens
await SecureStore.deleteItemAsync('accessToken');
await SecureStore.deleteItemAsync('refreshToken');
```

**Flutter:**
```dart
import 'package:flutter_secure_storage/flutter_secure_storage.dart';

class TokenManager {
  final storage = FlutterSecureStorage();

  Future<void> saveTokens(String accessToken, String refreshToken) async {
    await storage.write(key: 'accessToken', value: accessToken);
    await storage.write(key: 'refreshToken', value: refreshToken);
  }

  Future<String?> getAccessToken() async {
    return await storage.read(key: 'accessToken');
  }

  Future<String?> getRefreshToken() async {
    return await storage.read(key: 'refreshToken');
  }

  Future<void> clearTokens() async {
    await storage.delete(key: 'accessToken');
    await storage.delete(key: 'refreshToken');
  }
}

// Usage
final tokenManager = TokenManager();
await tokenManager.saveTokens(accessToken, refreshToken);
```

### Step 2: Login and Get Tokens

**IMPORTANT:** Always include `X-Client-Type: mobile` header when logging in:

**Example (JavaScript/Fetch):**
```javascript
const response = await fetch('http://localhost:5000/api/v1/auth/login', {
  method: 'POST',
  headers: {
    'X-Client-Type': 'mobile',  // REQUIRED for mobile apps
    'Content-Type': 'application/json',
  },
  body: JSON.stringify({
    usernameOrEmail: 'admin@example.com',
    password: 'password123',
  }),
});

const data = await response.json();
if (data.success) {
  const { accessToken, refreshToken } = data.data;
  // Store tokens securely
  await SecureStore.setItemAsync('accessToken', accessToken);
  await SecureStore.setItemAsync('refreshToken', refreshToken);
}
```

### Step 3: Make Authenticated Requests

Always include the access token in the `Authorization` header:

**Example (JavaScript/Fetch):**
```javascript
const accessToken = await getAccessToken(); // from secure storage

const response = await fetch('http://localhost:5000/api/v1/users', {
  method: 'GET',
  headers: {
    'Authorization': `Bearer ${accessToken}`,
    'Content-Type': 'application/json',
  },
});

const data = await response.json();
```

**Example (iOS/Swift with Alamofire):**
```swift
import Alamofire

let accessToken = tokenManager.getToken(key: "accessToken")

let headers: HTTPHeaders = [
    "Authorization": "Bearer \(accessToken ?? "")",
    "Content-Type": "application/json"
]

AF.request("http://localhost:5000/api/v1/users",
           method: .get,
           headers: headers)
    .responseJSON { response in
        // Handle response
    }
```

**Example (Android/Kotlin with Retrofit):**
```kotlin
interface ApiService {
    @GET("v1/users")
    suspend fun getUsers(
        @Header("Authorization") token: String
    ): Response<UsersResponse>
}

// Usage
val accessToken = tokenManager.getToken("accessToken")
val response = apiService.getUsers("Bearer $accessToken")
```

### Step 4: Handle Token Expiration

Implement automatic token refresh when access token expires:

**Example Flow:**
```javascript
async function makeAuthenticatedRequest(url, options = {}) {
  let accessToken = await getAccessToken();

  // Try with current access token
  let response = await fetch(url, {
    ...options,
    headers: {
      ...options.headers,
      'Authorization': `Bearer ${accessToken}`,
    },
  });

  // If 401, refresh token and retry
  if (response.status === 401) {
    const refreshToken = await getRefreshToken();

    // Refresh tokens - IMPORTANT: Include X-Client-Type header
    const refreshResponse = await fetch('http://localhost:5000/api/v1/auth/refresh', {
      method: 'POST',
      headers: {
        'X-Client-Type': 'mobile',  // REQUIRED for mobile apps
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ refreshToken }),
    });

    if (refreshResponse.ok) {
      const data = await refreshResponse.json();
      const { accessToken: newAccessToken, refreshToken: newRefreshToken } = data.data;

      // Save new tokens
      await saveAccessToken(newAccessToken);
      await saveRefreshToken(newRefreshToken);

      // Retry original request with new token
      response = await fetch(url, {
        ...options,
        headers: {
          ...options.headers,
          'Authorization': `Bearer ${newAccessToken}`,
        },
      });
    } else {
      // Refresh failed, redirect to login
      await logout();
      throw new Error('Session expired. Please login again.');
    }
  }

  return response;
}
```

### Step 5: Implement Logout

Clear tokens from secure storage:

```javascript
async function logout() {
  // Call logout endpoint (optional, mainly clears server-side cookies)
  await fetch('http://localhost:5000/api/v1/auth/logout', {
    method: 'POST',
  });

  // Clear local tokens
  await deleteAccessToken();
  await deleteRefreshToken();

  // Navigate to login screen
  navigation.navigate('Login');
}
```

## Why X-Client-Type Header?

The `X-Client-Type: mobile` header is required for security reasons:

**For Web Apps:**
- Tokens should NEVER be exposed to JavaScript (XSS vulnerability)
- Should only use httpOnly cookies that JavaScript cannot access
- Cookies are automatically sent with every request

**For Mobile Apps:**
- Cannot use httpOnly cookies effectively
- Need tokens in response body to store in secure storage
- Manually include Authorization header in each request

By requiring the `X-Client-Type: mobile` header, we ensure:
- ✅ Web apps never receive tokens in response body (secure by default)
- ✅ Mobile apps explicitly request tokens
- ✅ Single API endpoint for both platforms
- ✅ Better XSS protection for web applications

## Security Best Practices

### ✅ DO

1. **Always include X-Client-Type: mobile header for login/refresh**
   ```javascript
   headers: {
     'X-Client-Type': 'mobile',
     'Content-Type': 'application/json'
   }
   ```

2. **Store tokens in secure storage only**
   - iOS: Keychain
   - Android: EncryptedSharedPreferences
   - React Native: expo-secure-store or react-native-keychain
   - Flutter: flutter_secure_storage

3. **Use HTTPS in production**
   ```javascript
   const API_BASE_URL = __DEV__
     ? 'http://localhost:5000/api'
     : 'https://api.yourdomain.com/api';
   ```

4. **Implement token refresh logic**
   - Automatically refresh when 401 occurs
   - Refresh before token expires (proactive)
   - Always include `X-Client-Type: mobile` header in refresh requests

5. **Clear tokens on logout**

6. **Handle network errors gracefully**

### ❌ DON'T

1. **Never store tokens in:**
   - AsyncStorage/SharedPreferences (unencrypted)
   - LocalStorage (web)
   - Plain text files
   - Redux/Vuex state (persisted)

2. **Never log tokens in console**
   ```javascript
   // ❌ Bad
   console.log('Token:', accessToken);

   // ✅ Good
   console.log('Token received:', !!accessToken);
   ```

3. **Never include tokens in URLs**
   ```javascript
   // ❌ Bad
   fetch(`/api/users?token=${accessToken}`)

   // ✅ Good
   fetch('/api/users', {
     headers: { 'Authorization': `Bearer ${accessToken}` }
   })
   ```

## Token Lifetimes

- **Access Token:** 15 minutes (configurable via JWT_ACCESS_EXPIRATION)
- **Refresh Token:** 7 days (configurable via JWT_REFRESH_EXPIRATION)

## API Response Format

All responses follow this structure:

**Success Response:**
```json
{
  "success": true,
  "message": "Operation successful",
  "data": { ... }
}
```

**Error Response:**
```json
{
  "success": false,
  "message": "Error message",
  "error": {
    "code": "ERROR_CODE",
    "detail": "Detailed error description"
  }
}
```

## Common Error Codes

| Code | Description | Action |
|------|-------------|--------|
| UNAUTHORIZED | Missing or invalid token | Redirect to login |
| INVALID_TOKEN | Token expired or malformed | Refresh token or login |
| USER_NOT_FOUND | User doesn't exist | Redirect to login |
| INVALID_CREDENTIALS | Wrong username/password | Show error message |

## Testing with Postman

1. **Login:**
   - POST `http://localhost:5000/api/v1/auth/login`
   - Copy `accessToken` from response

2. **Make Authenticated Request:**
   - Add header: `Authorization: Bearer <accessToken>`
   - Example: GET `http://localhost:5000/api/v1/users`

3. **Test Token Refresh:**
   - POST `http://localhost:5000/api/v1/auth/refresh`
   - Body: `{ "refreshToken": "<refreshToken>" }`

## Support

For issues or questions:
- Check API documentation: `http://localhost:5000/swagger/index.html`
- GitHub Issues: https://github.com/your-repo/issues
- Email: support@example.com

---

**Last Updated:** 2024-01-15
