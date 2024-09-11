# Otto: Automatic Load Balancer for Localhost Applications

Otto is a simple, efficient load balancer designed for localhost applications, written in Go. It uses a round-robin algorithm to distribute incoming requests across multiple local servers, ensuring balanced load and improved performance.

## Features

- Round-robin load balancing
- Automatic health checks
- Easy configuration via JSON file
- Support for multiple backend servers

## Getting Started

### Prerequisites

- Go 1.15 or higher

### Installation

1. Clone the repository:
   ```
   git clone https://github.com/yourusername/otto.git
   ```
2. Navigate to the project directory:
   ```
   cd otto
   ```

### Configuration

Create a `config.json` file in the project root with the following structure:

```json
{
  "port": ":8080",
  "healthCheckInterval": "5s",
  "servers": [
    "http://localhost:5001",
    "http://localhost:5002",
    "http://localhost:5003",
    "http://localhost:5004",
    "http://localhost:5005"
  ]
}
```

Adjust the port, health check interval, and server list as needed.

### Running Otto

Execute the following command in the project root:

```
go run otto.go
```

Otto will start and begin load balancing requests to your configured localhost applications.

## How It Works

1. Otto reads the configuration from `config.json`.
2. It sets up reverse proxies for each configured backend server.
3. A health check routine runs periodically for each server.
4. Incoming requests are distributed among healthy servers using a round-robin algorithm.

## Potential Improvements

1. **Weighted Round Robin**: Implement a weighted algorithm to allow some servers to receive more traffic than others.
2. **Least Connections**: Add an option to route traffic to the server with the least active connections.
3. **IP Hash**: Implement IP-based routing to ensure requests from the same client always go to the same server.
4. **Dynamic Server Addition/Removal**: Allow adding or removing servers without restarting Otto.
5. **HTTPS Support**: Add support for SSL/TLS termination.
6. **Logging and Metrics**: Implement detailed logging and expose metrics for monitoring.
7. **Rate Limiting**: Add the ability to limit the number of requests per client.
8. **Caching**: Implement a caching layer to reduce load on backend servers.
9. **Configuration Hot Reload**: Allow reloading configuration without restarting the service.
10. **Web UI**: Create a simple web interface for monitoring and configuration.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
