# Chirpy - Social Media API

A lightweight social media platform API built with Go, featuring user authentication, message posting (chirps), and premium subscriptions. This project demonstrates modern Go web development practices with PostgreSQL, JWT authentication, and RESTful API design.

## 🚀 Features

- **User Management**: Registration, login, profile updates
- **Authentication**: JWT tokens with refresh token support
- **Chirps**: Create, read, and delete short messages (140 characters)
- **Premium Subscriptions**: Chirpy Red premium tier via webhook integration
- **Admin Panel**: Basic analytics and system management
- **Content Moderation**: Automatic profanity filtering

## 🛠 Tech Stack

- **Language**: Go 1.21+
- **Database**: PostgreSQL
- **Authentication**: JWT (HMAC-SHA256)
- **Password Hashing**: bcrypt
- **Database Queries**: sqlc (type-safe SQL)
- **HTTP Router**: Go standard library (net/http)
- **Environment**: godotenv for configuration

## 📋 Prerequisites

- Go 1.21 or higher
- PostgreSQL 12+
- sqlc (for database code generation)
- goose (for database migrations)

## 🔧 Installation

1. **Clone the repository**
   ```bash
   git clone https://github.com/yourusername/httpChirpy.git
   cd httpChirpy
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Set up environment variables**
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

4. **Configure your `.env` file**
   ```env
   DB_URL=postgres://username:password@localhost:5432/chirpy?sslmode=disable
   JWT_SECRET_KEY=your-super-secret-jwt-key
   PLATFORM=dev
   POLKA_KEY=your-webhook-secret-key
   ```

5. **Run database migrations**
   ```bash
   goose -dir sql/schema postgres $DB_URL up
   ```

6. **Generate database code**
   ```bash
   sqlc generate
   ```

7. **Start the server**
   ```bash
   go run .
   ```

The server will start on `http://localhost:8080`

## 📚 API Documentation

### Authentication Endpoints

#### Register User
```http
POST /api/users
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "securepassword"
}
```

#### Login
```http
POST /api/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "securepassword"
}
```

#### Refresh Token
```http
POST /api/refresh
Authorization: Bearer <refresh_token>
```

#### Revoke Token
```http
POST /api/revoke
Authorization: Bearer <refresh_token>
```

### Chirp Endpoints

#### Create Chirp
```http
POST /api/chirps
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "body": "This is my first chirp!"
}
```

#### Get All Chirps
```http
GET /api/chirps?sort=desc&author_id=<user_id>
```

#### Get Chirp by ID
```http
GET /api/chirps/{chirpID}
```

#### Delete Chirp
```http
DELETE /api/chirps/{chirpID}
Authorization: Bearer <access_token>
```

### User Management

#### Update User
```http
PUT /api/users
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "email": "newemail@example.com",
  "password": "newpassword"
}
```

### Webhook Endpoints

#### Polka Webhook (Premium Upgrades)
```http
POST /api/polka/webhooks
Content-Type: application/json

{
  "event": "user.upgraded",
  "data": {
    "user_id": "user-uuid-here"
  }
}
```

### Admin Endpoints

#### Health Check
```http
GET /api/healthz
```

#### Metrics
```http
GET /admin/metrics
```

#### Reset System (Development Only)
```http
POST /admin/reset
```

## 🏗 Project Structure

```
httpChirpy/
├── internal/
│   ├── auth/                 # Authentication utilities
│   │   ├── auth.go          # JWT and password handling
│   │   └── auth_test.go     # Authentication tests
│   └── database/            # Database layer
│       ├── db.go           # Database connection
│       ├── models.go       # Data models
│       └── *.sql.go        # Generated query code
├── sql/
│   ├── queries/            # SQL query definitions
│   │   ├── users.sql
│   │   ├── chirps.sql
│   │   └── refresh_tokens.sql
│   └── schema/             # Database migrations
│       ├── 001_users.sql
│       ├── 002_chirps.sql
│       └── ...
├── assets/                 # Static assets
├── main.go                # Application entry point
├── go.mod                 # Go module definition
└── README.md             # This file
```

## 🧪 Testing

Run the test suite:
```bash
go test ./...
```

Run tests with coverage:
```bash
go test -cover ./...
```

Run specific package tests:
```bash
go test ./internal/auth
```

## 🔒 Security Features

- **Password Security**: bcrypt hashing with configurable cost
- **JWT Authentication**: Secure token-based authentication
- **Token Refresh**: Long-lived refresh tokens for session management
- **Input Validation**: Request validation and sanitization
- **Content Filtering**: Automatic profanity filtering
- **SQL Injection Protection**: Parameterized queries via sqlc

## 🚀 Production Considerations

This is a learning/demonstration project. For production use, consider:

### Security Enhancements
- [ ] Rate limiting middleware
- [ ] HTTPS/TLS configuration
- [ ] CORS policy implementation
- [ ] Input validation middleware
- [ ] SQL injection prevention auditing
- [ ] Security headers middleware

### Performance Optimizations
- [ ] Database connection pooling
- [ ] Query result caching
- [ ] Response compression
- [ ] Database indexing optimization
- [ ] Pagination for large datasets

### Monitoring & Observability
- [ ] Structured logging (JSON format)
- [ ] Metrics collection (Prometheus)
- [ ] Distributed tracing
- [ ] Health check endpoints
- [ ] Error tracking (Sentry)

### Infrastructure
- [ ] Docker containerization
- [ ] Kubernetes deployment
- [ ] Database migrations in CI/CD
- [ ] Environment-specific configurations
- [ ] Graceful shutdown handling

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Built as part of the Boot.dev Go course
- Inspired by Twitter's early architecture
- Uses industry-standard Go practices and patterns

## 📞 Support

If you have questions or need help:
- Open an issue on GitHub
- Check the documentation in the `.go.example` files
- Review the test files for usage examples

---

**Note**: This project includes comprehensive documentation in `.go.example` files that demonstrate production-ready code patterns and best practices for Go web development.