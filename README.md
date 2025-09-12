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
- **Database**: PostgreSQL - Currently hosted using Railway
- **Message Broker**: RabbitMQ - Currently hosted using Railway
- **WebSocket**: Gorilla WebSocket
- **Router**: Gorilla Mux

## Setup & Installation

1. **Set Environment**

Use the provided .env file and store it in the root of the repository. If no .env is found, feel free to reach out and i can provide it to you.
The .env has the following ENVs:

```env
# These are just example values.
PORT=PORT
SECRET_KEY=SECRET_KEY
DATABASE_URL=DATABASE_URL
RABBIT_MQ_URL=RABBIT_MQ_URL
CHATBOT_EMAIL=CHATBOT_EMAIL
```

Install Go and Makefile if you're planning to run it directly with Go. Then run the following command in the root of the repository.

```bash
go mod download
```

2. **Database Setup**

This step is optional since the DB is hosted in Railway with the necessary data

```bash
make install_migration
make migration_up
```

### Start the application

To run the application manually please follow the Setup & Installation section. Then write the following command in the CMD

```bash
make run
```

A less troublesome way to start the application is by using Docker. In the root of the repository run the following commands:

```bash
docker build -t chatroom-go .
docker run --env-file .env -p 8080:8080 chatroom-go
```

An even less troublesome way to start the application is by using the Makefile commands:

```bash
make build_image
make start_image
```

## API Documentation

### Public Endpoints

| Method | Endpoint | Description   | Request Body                                                                                           |
| ------ | -------- | ------------- | ------------------------------------------------------------------------------------------------------ |
| POST   | `/user/` | Register user | `{"user_email": "user@example.com", "user_password": "secret", "user_user_name": "user_name_example"}` |
| POST   | `/login` | Login user    | `{"user_email": "user@example.com", "user_password": "secret"}`                                        |

### Protected Endpoints

| Method | Endpoint            | Description          | Authentication | Request Body                       | Response                                                    |
| ------ | ------------------- | -------------------- | -------------- | ---------------------------------- | ----------------------------------------------------------- |
| POST   | `/chatrooms/`       | Create chatroom      | Required       | `{"chatroom_name": "My Chatroom"}` | `{"chatroom_id": "uuid"}`                                   |
| GET    | `/chatrooms`        | List chatrooms       | Required       | -                                  | `[{"chatroom_id": "uuid", "chatroom_name": "My Chatroom"}]` |
| GET    | `/ws/chatroom/{id}` | WebSocket connection | Required       | -                                  | WebSocket Connection                                        |

**Note**: For protected endpoints, include the JWT token in the request header:

```
Authorization: Bearer <your_token>
```

Example Response for `http://localhost:8080/chatrooms`:

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
