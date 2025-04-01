# Sensor API - Rust Version

This is a Rust implementation of a RESTful API for sensor data management. The application connects to a MySQL database to store and retrieve IAQ_SEN55 sensor readings.

## Features

- RESTful API for sensor data
- MySQL database integration using SQLx
- JSON serialization/deserialization with Serde
- CORS support
- Configurable via environment variables
- Structured logging

## API Endpoints

- `GET /api/sensor-data` - Get sensor data with optional filtering (startDate, endDate, limit)
- `POST /api/sensor-data` - Upload new sensor data
- `GET /api/sensor-data/latest` - Get the latest sensor data (with optional limit)

## Requirements

- Rust 1.70+
- MySQL database
- MySQL client libraries installed on your system

## Setup

1. Create a `.env` file in the project root with the following configuration (or set environment variables):

```
DATABASE_URL=mysql://data:mysqldatapassword@localhost/buildingData
PORT=8080
RUST_LOG=info
```

2. Set up the MySQL database with the required table:

```sql
CREATE TABLE IF NOT EXISTS IAQ_SEN55 (
  id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
  location VARCHAR(50) NOT NULL,
  recTime TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  temp FLOAT,
  rH FLOAT,
  VOC FLOAT,
  NOx FLOAT,
  pmass1 FLOAT,
  pmass25 FLOAT,
  pmass4 FLOAT,
  pmass10 FLOAT,
  HCHO FLOAT,
  CO2 FLOAT,
  indoorTd FLOAT
);
```

3. Build the project:

```
cargo build --release
```

## Running the Application

```
cargo run --release
```

The API will be available at `http://localhost:8080` (or the configured port).

## Example Requests

### Get Sensor Data

```
curl -X GET "http://localhost:8080/api/sensor-data?limit=10&startDate=2023-01-01&endDate=2023-12-31"
```

### Get Latest Sensor Data

```
curl -X GET "http://localhost:8080/api/sensor-data/latest"
```

### Post Sensor Data

```
curl -X POST "http://localhost:8080/api/sensor-data" \
  -H "Content-Type: application/json" \
  -d '{
    "location": "1",
    "temp": 25.4,
    "rH": 45.2,
    "VOC": 12.5,
    "NOx": 0.8,
    "pmass1": 0.2,
    "pmass25": 0.5,
    "pmass4": 0.7,
    "pmass10": 1.2,
    "HCHO": 0.05,
    "CO2": 850,
    "indoorTd": 18.3
  }' 