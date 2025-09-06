# Go Chatroom Application

A real-time chat application built with Go, featuring WebSocket communication, PostgreSQL persistence, and RabbitMQ integration.

## Features

- Real-time messaging using WebSocket
- User authentication with JWT
- Persistent message storage
- Chatbot integration for stock quotes
- Multiple chatroom support

## Tech Stack

- **Backend**: Go 1.19+, Makefile
- **Database**: PostgreSQL - Hosted using Railway
- **Message Broker**: RabbitMQ - Hosted using Railway
- **WebSocket**: Gorilla WebSocket
- **Router**: Gorilla Mux

## Setup & Installation

1. **Set Environment Variables**

Create `.env` file:
```env
PORT=8080
SECRET_KEY=SECRET_KEY
DB_URL=postgres://username:password@localhost:5432/chatroom?sslmode=disable
RABBIT_MQ_URL=amqp://guest:guest@localhost:5672/
CHATBOT_EMAIL=chatbot@example.com
```

2. **Database Setup**
```bash
make install_migration
make migration_up
```

3. **Run Application**
```bash
make run
```

## API Documentation

### Public Endpoints

| Method | Endpoint | Description | Request Body |
|--------|----------|-------------|--------------|
| POST | `/user/` | Register user | `{"email": "user@example.com", "password": "secret"}` |
| POST | `/login` | Login user | `{"email": "user@example.com", "password": "secret"}` |

### Protected Endpoints

| Method | Endpoint | Description | Authentication | Request Body | Response |
|--------|----------|-------------|----------------|--------------|-----------|
| POST | `/chatrooms/` | Create chatroom | Required | `{"chatroom_name": "My Chatroom"}` | `{"chatroom_id": "uuid"}` |
| GET | `/chatrooms` | List chatrooms | Required | - | `[{"chatroom_id": "uuid", "chatroom_name": "My Chatroom"}]` |
| GET | `/ws/chatroom/{id}` | WebSocket connection | Required | - | WebSocket Connection |

**Note**: For protected endpoints, include the JWT token in the request header:
```
Authorization: Bearer <your_token>
```

Example Response for `/chatrooms`:
```json
[
  {
    "chatroom_id": "uuid-string",
    "chatroom_name": "Super Chatroom"
  }
]
```

Example Request for creating a chatroom:
```json
POST /chatrooms/
{
    "chatroom_name": "My New Chatroom"
}
```

Example Response:
```json
{
    "chatroom_id": "uuid-string"
}
```

### Running Tests
```bash
# Run all tests with verbose output
make run_test_verbose

# Run all tests with coverage
make run_test
```

### Project Structure
```
chatroom/
├── .env                # Environment variables
├── .env.example       # Environment variables template
├── .gitignore        # Git ignore rules
├── Makefile          # Build and run commands
├── go.mod            # Go module definition
├── go.sum            # Go module checksums
├── cmd/              # Application entry points
│   └── chatroom/     # Main application
│       └── main.go   # Entry point
├── chatbot/          # Chatbot implementation
│   └── chatbot.go    # Chatbot logic
├── chatroom/         # Main application logic
│   ├── handlers/     # HTTP request handlers
│   │   └── handler.go # Handler implementations
│   └── chatroom.go   # Service implementation
├── interfaces/       # Interface definitions
│   ├── chatbot.go   # Chatbot interfaces
│   └── db.go        # Database interfaces
├── migrations/       # Database migrations
│   ├── 000001_init.up.pgsql   # Initial schema
│   └── 000001_init.down.pgsql # Rollback schema
├── models/          # Data models
│   ├── client.go    # WebSocket client
│   ├── db.go        # Database models
│   ├── error.go     # Error definitions
│   └── hub.go       # WebSocket hub
├── repos/           # Database repositories
│   ├── db.go        # Database operations
│   └── db_test.go   # Database tests
├── utils/          # Utility functions
│   ├── encrypt.go  # Password encryption
│   └── http.go     # HTTP utilities
└── README.md       # Project documentation
```