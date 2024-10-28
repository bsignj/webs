# Go WebSocket Project

This project is a WebSocket-based application written in Go, with performance testing handled by k6.

## Prerequisites

Before running the app, ensure the following software is installed on your system:

1. **k6**: Install the latest version of k6. You can find the installation instructions [here](https://grafana.com/docs/k6/latest/set-up/install-k6/).

2. **Go**: Make sure the latest version of Go is installed. Installation instructions can be found [here](https://go.dev/doc/install).

## Setup

1. Clone the repository to your local machine.

```bash
git clone <repository-url>
cd <project-directory>
```

2. In the root directory of the project, rename the configuration file:

```bash
mv config.yaml.example config.yaml
```

## Running the Application

To run the application, follow these steps:

1. Run the Go application:

```bash
go run main.go
```

2. Run the performance test using k6:

```bash
k6 run perf-test.js
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
