# Scopion

A single-binary observability tool for collecting and visualizing application telemetry data.

Scopion provides a simple way to ingest, store, and visualize events and traces from applications. It includes both a REST API for data ingestion and a web-based dashboard for monitoring and analysis.

## Features

*   **Single-Binary Deployment:** Self-contained executable with embedded web interface for easy distribution and deployment.
*   **Real-time Monitoring:** Dashboard for viewing live events, traces, and system metrics.
*   **Data Ingestion:** REST API endpoint for receiving telemetry data from applications.
*   **Demo Mode:** Optional sample data generation for testing and demonstration purposes.
*   **SQLite Storage:** Built-in database for storing events and traces.
*   **Web Interface:** Modern React-based dashboard for data visualization and exploration.

## Technology Stack

*   **Backend:** Go 1.25.5, SQLite, Server-Sent Events
*   **Frontend:** React, TypeScript, Vite, Tailwind CSS, Recharts, Zustand
*   **Build:** Make, npm

## Getting Started

### Prerequisites

*   Go (version 1.25.5 or higher)
*   Node.js and npm (or bun)

### Installation

1.  **Clone the repository:**
    ```bash
    git clone https://github.com/xonoxc/scopion.git && cd scopion
    ```
2.  **Install Go dependencies:**
    ```bash
    go mod tidy
    ```
3.  **Install Node.js dependencies:**
    ```bash
    cd ui && npm install
    # or use bun:
    # cd ui && bun install
    ```

### Running the Application

#### Development Mode

For development with hot-reloading of the frontend:

```bash
make dev
```

This starts the backend server and frontend development server separately. Access the UI at `http://localhost:5173`.

#### Production Mode

1. **Build the application:**
   ```bash
   make build
   ```

2. **Run the server:**
   ```bash
   ./bin/scopion start
   ```

   Or use the make target:
   ```bash
   make run
   ```

The production binary embeds the entire frontend application and serves it from the configured port.

## Usage

### Command Line Interface

Scopion provides a command-line interface for starting and managing the server.

#### Starting the Server

```bash
scopion start [flags]
```

**Flags:**
- `--port, -p`: Port to run the server on (default "8080")
- `--demo`: Enable demo data generation (default true)

**Examples:**

Start the server on the default port with demo data enabled:
```bash
scopion start
```

Start the server on a specific port with demo data disabled:
```bash
scopion start --port 3000 --demo=false
```

#### Other Commands

- `scopion version`: Display the version information
- `scopion help`: Display help information

#### Server Endpoints

Once running, the server provides the following endpoints:

- `GET /`: Main web interface
- `GET /api/events`: Recent events data
- `GET /api/stats`: System statistics
- `GET /api/services`: Service information
- `GET /api/traces`: Trace data
- `GET /api/errors-by-service`: Error data grouped by service
- `GET /api/search`: Search events
- `GET /api/status`: Server status and configuration
- `POST /ingest`: Ingest telemetry data

### Web Interface

Access the Scopion dashboard in your browser at `http://localhost:8080` (or the configured port) to monitor system events and traces. The interface provides real-time visualization of application behavior, performance metrics, and error tracking.

## Configuration

Scopion uses SQLite for data storage and creates a `scopion.db` file in the current working directory. The web interface is embedded in the binary and served on the configured port.

### Environment Variables

None required. All configuration is done via command-line flags.

## Contributing

Contributions are welcome. Please ensure code follows Go and TypeScript best practices.

## License

This project is licensed under the MIT License - see the `LICENSE` file for details. 
