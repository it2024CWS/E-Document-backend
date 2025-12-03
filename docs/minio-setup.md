# MinIO File Storage Setup

## Overview

MinIO is an object storage system compatible with Amazon S3. This project uses MinIO for storing user profile pictures and other file uploads.

## Configuration

### Environment Variables

Add the following to your `.env` file:

```env
# MinIO Configuration
MINIO_ENDPOINT=localhost:9000
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin
MINIO_BUCKET=edocument-files
MINIO_USE_SSL=false
MINIO_PUBLIC_URL=http://localhost:9000
```

### Docker Compose Setup

The `docker-compose.yml` includes three services:

1. **MongoDB** - Database
2. **MinIO** - Object storage server
3. **CreateBuckets** - Helper to create initial bucket

## Getting Started

### 1. Start Services

```bash
docker-compose up -d
```

This will start:
- MongoDB on port `27017`
- MinIO API on port `9000`
- MinIO Console on port `9001`

### 2. Access MinIO Console

Open your browser and go to: `http://localhost:9001`

**Login credentials:**
- Username: `minioadmin`
- Password: `minioadmin`

### 3. Verify Bucket Creation

The `edocument-files` bucket should be automatically created. You can verify in the MinIO Console.

## API Endpoints

### Upload Profile Picture

**Endpoint:** `POST /api/v1/users/:id/profile-picture`

**Headers:**
- `Authorization: Bearer <token>`
- `Content-Type: multipart/form-data`

**Body:**
- `file`: Image file (max 5MB, jpg/jpeg/png/gif/webp)

**Example using cURL:**

```bash
curl -X POST http://localhost:5000/api/v1/users/{user_id}/profile-picture \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -F "file=@/path/to/image.jpg"
```

**Response:**

```json
{
  "success": true,
  "message": "Profile picture uploaded successfully",
  "data": {
    "id": "507f1f77bcf86cd799439011",
    "username": "johndoe",
    "email": "john@example.com",
    "profile_picture": "http://localhost:9000/edocument-files/profiles/1234567890_image.jpg",
    "first_name": "John",
    "last_name": "Doe",
    "role": "Employee",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

### Delete Profile Picture

**Endpoint:** `DELETE /api/v1/users/:id/profile-picture`

**Headers:**
- `Authorization: Bearer <token>`

**Example using cURL:**

```bash
curl -X DELETE http://localhost:5000/api/v1/users/{user_id}/profile-picture \
  -H "Authorization: Bearer YOUR_TOKEN"
```

**Response:**

```json
{
  "success": true,
  "message": "Profile picture deleted successfully",
  "data": {
    "id": "507f1f77bcf86cd799439011",
    "username": "johndoe",
    "email": "john@example.com",
    "profile_picture": "",
    "first_name": "John",
    "last_name": "Doe",
    "role": "Employee",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

## File Validation

### Supported Formats
- JPEG/JPG
- PNG
- GIF
- WebP

### File Size Limit
- Maximum: 5MB

### Validation Rules
1. File size must not exceed 5MB
2. Content-Type must be a valid image MIME type
3. File extension must be in the allowed list

## Storage Structure

Files are organized by folder:

```
edocument-files/
  └── profiles/
      ├── 1234567890_user1.jpg
      ├── 1234567891_user2.png
      └── ...
```

## Production Considerations

### 1. Security

For production, update MinIO credentials:

```env
MINIO_ACCESS_KEY=your-secure-access-key
MINIO_SECRET_KEY=your-secure-secret-key
```

### 2. SSL/TLS

Enable SSL for production:

```env
MINIO_USE_SSL=true
MINIO_ENDPOINT=minio.yourdomain.com:9000
MINIO_PUBLIC_URL=https://minio.yourdomain.com:9000
```

### 3. Bucket Policy

For public read access to profile pictures, set bucket policy in MinIO Console:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "AWS": ["*"]
      },
      "Action": ["s3:GetObject"],
      "Resource": ["arn:aws:s3:::edocument-files/profiles/*"]
    }
  ]
}
```

### 4. CDN Integration

Consider using a CDN (CloudFlare, AWS CloudFront) to serve files for better performance.

## Troubleshooting

### Connection Refused

If you get "connection refused" error:

1. Check if MinIO is running: `docker ps`
2. Check MinIO logs: `docker logs edocument-minio`
3. Verify environment variables are set correctly

### Bucket Not Found

If bucket doesn't exist:

1. Access MinIO Console: `http://localhost:9001`
2. Create bucket manually named `edocument-files`
3. Or run the createbuckets service again

### Upload Fails

Common issues:

1. **File too large**: Reduce file size below 5MB
2. **Invalid format**: Use supported image formats only
3. **Permission denied**: Check MinIO access credentials
4. **Network error**: Verify MinIO endpoint is reachable

## Development vs Production

### Development (Current Setup)

- Uses `localhost:9000`
- HTTP (not HTTPS)
- Default credentials
- Single instance

### Production Recommendations

- Use dedicated MinIO server or cloud storage (AWS S3, Google Cloud Storage)
- Enable HTTPS/SSL
- Use strong credentials
- Implement backup strategy
- Consider high availability setup
- Use reverse proxy (Nginx)
- Implement rate limiting on upload endpoints

## Mobile App Integration

When integrating with mobile apps:

1. Upload returns the full URL: `http://localhost:9000/edocument-files/profiles/...`
2. Mobile app can directly load images from this URL
3. For authenticated access, implement signed URLs
4. Consider image thumbnails for better performance

## Next Steps

1. ✅ Profile picture upload/delete implemented
2. Future enhancements:
   - Image resizing/optimization
   - Multiple file upload
   - Document storage for E-Document system
   - Thumbnail generation
   - Signed URLs for private files
   - File versioning
