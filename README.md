# Microsaur (Micro dinosaur)

A lightweight and flexible Go-based application that enables you to quickly set up a remote SQLite database server accessible over HTTP. Simplify database management by interacting with SQLite databases remotely, executing queries, and performing updates seamlessly through HTTP requests. Please note that this is a tiny experiment and is not intended for production use.

## Features

- **Remote SQLite Database**: Spin up and manage SQLite databases remotely.
- **HTTP Interface**: Interact with the database using a straightforward HTTP API.

- **Query and Update**: Execute queries and update operations over HTTP, making it easy to integrate with various applications.

- **Secure Communication**: Communicate securely with the remote database server using standard HTTP protocols.

- **Flexible and Easy to Use**: Designed for simplicity and flexibility, allowing you to focus on building applications rather than managing databases.

## Getting Started

1. Clone the repository: `git clone https://github.com/oreoluwa-bs/microsaur.git`

2. Build and run the server: `go run main.go`

3. Access the API and start querying the remote SQLite database.

## Usage Examples

```http
POST /database
Content-Type: application/json

{
  "name": "myDatabase"
}


GET /database
Content-Type: application/json

{
  "id" : "1"
  "name": "myDatabase"
}

POST /database/{databaseId}
Content-Type: application/json

{
  "sql": "INSERT INTO users (name, age) VALUES (?, ?)",
  "params": ["John Doe", 30]
}
```

## Contributions

Contributions are welcome! Feel free to submit issues, pull requests, or provide feedback to improve the functionality and usability of this application.

<!-- ## License

This project is licensed under the [MIT License](LICENSE). -->
