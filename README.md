# Staffy-SSO

Authentication and authorization microservice for the Staffy platform. Handles user registration, login, JWT token management, and user data retrieval.

## API Endpoints

### 🔐 Register
Creates a new user account and returns authentication token.

**Request:**
```json
{
    "email": "user@example.com",
    "name": "John",
    "surname": "Doe",
    "is_recruiter": true,
    "password": "securepassword123"
}
```
**Response:**
```json
{
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
        "user_id": "9e868144-7b19-4675-ba77-ba9333c9b27f",
        "email": "user@example.com",
        "name": "John",
        "surname": "Doe",
        "is_recruiter": true
    }
}
```

### 🔑 Login
Authenticates existing user and returns JWT token.

**Request:**
```json
{
    "email": "user@example.com",
    "password": "securepassword123"
}
```
**Response:**
```json
{
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
        "user_id": "9e868144-7b19-4675-ba77-ba9333c9b27f",
        "email": "user@example.com",
        "name": "John",
        "surname": "Doe",
        "is_recruiter": true
    }
}
```

### 👤 GetUserByToken
Retrieves user information using JWT token.

**Request:**
```json
{
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```
**Response:**
```json
{
    "user_id": "9e868144-7b19-4675-ba77-ba9333c9b27f",
    "email": "user@example.com",
    "name": "John",
    "surname": "Doe",
    "is_recruiter": true
}
```

### 🗑️ Delete User
Deletes user account (requires authentication).

**Request:**
```json
{
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```
**Response:**
```json
{
    "timestamp": "1761402621",
    "status_code": "200",
    "status_message": "User successfully deleted"
}
```

### 🔄 Refresh Token
Generates new JWT token while maintaining user session.

**Request:**
```json
{
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```
**Response:**
```json
{
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

## Technology Stack

- **gRPC** - High-performance RPC framework
- **JWT** - JSON Web Tokens for authentication
- **PostgreSQL** - Primary data storage
- **Redis** - Caching layer for performance
- **Go** - Backend programming language

## Error Handling

All endpoints return appropriate gRPC status codes:
- `0 OK` - Successful operation
- `3 Invalid Arguments` - Invalid input parameters
- `5 Not Found` - User not found
- `6 Already Exists` - User already exists
- `13 Internal Server Error` - Server-side issues

## Security Features

- Password hashing with bcrypt
- JWT token expiration
- Input validation and sanitization