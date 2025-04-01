package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	"github.com/rs/cors"
)

// SensorData represents the IAQ_SEN55 sensor data
type SensorData struct {
	ID       int       `json:"id" db:"id"`
	Location string    `json:"location" db:"location"`
	RecTime  time.Time `json:"recTime" db:"recTime"`
	Timestamp string   `json:"timestamp" db:"recTime"` 
	Temp     float64   `json:"temp" db:"temp"`
	RH       float64   `json:"rH" db:"rH"`
	VOC      float64   `json:"VOC" db:"VOC"`
	NOx      float64   `json:"NOx" db:"NOx"`
	Pmass1   float64   `json:"pmass1" db:"pmass1"`
	Pmass25  float64   `json:"pmass25" db:"pmass25"`
	Pmass4   float64   `json:"pmass4" db:"pmass4"`
	Pmass10  float64   `json:"pmass10" db:"pmass10"`
	HCHO     float64   `json:"HCHO" db:"HCHO"`
	CO2      float64   `json:"CO2" db:"CO2"`
	IndoorTd float64   `json:"indoorTd" db:"indoorTd"`
	Tags     []string  `json:"tags,omitempty"` // Optional tags field
}

// Database configuration - same as in your TypeScript app
var dbConfig struct {
	Host     string
	User     string
	Password string
	Database string
	Port     int
}

var db *sqlx.DB
var err error
// Initialize database connection
func initDB() error {
	dbConfig.User = os.Getenv("MYSQL_USER")
	dbConfig.Password = os.Getenv("MYSQL_PASS")
	dbConfig.Host = os.Getenv("MYSQL_HOST")
	port := os.Getenv("MYSQL_PORT")
	dbConfig.Database = os.Getenv("MYSQL_DB")

	dbConfig.Port, err = strconv.Atoi(port)
	if err != nil {
		log.Printf("Error converting port to int: %v", err)
		log.Printf("Using default port 3306")
		dbConfig.Port = 3306
	}
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
		dbConfig.User, dbConfig.Password, dbConfig.Host, dbConfig.Port, dbConfig.Database)
	
	var err error
	db, err = sqlx.Connect("mysql", dsn)
	if err != nil {
		return err
	}
	
	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	
	return nil
}

// Sanitize sensor data to handle NaN and Infinity values
func sanitizeSensorData(data *SensorData) {
	// Helper function to sanitize float values
	sanitizeFloat := func(val *float64) {
		if val == nil || !isValidFloat(*val) {
			*val = 0.0
		}
	}
	
	sanitizeFloat(&data.Temp)
	sanitizeFloat(&data.RH)
	sanitizeFloat(&data.VOC)
	sanitizeFloat(&data.NOx)
	sanitizeFloat(&data.Pmass1)
	sanitizeFloat(&data.Pmass25)
	sanitizeFloat(&data.Pmass4)
	sanitizeFloat(&data.Pmass10)
	sanitizeFloat(&data.HCHO)
	sanitizeFloat(&data.CO2)
	sanitizeFloat(&data.IndoorTd)
}

// Check if float value is valid (not NaN or Infinity)
func isValidFloat(val float64) bool {
	return !isNaN(val) && !isInf(val, 0)
}

// Check if value is NaN
func isNaN(val float64) bool {
	return val != val
}

// Check if value is Infinity
func isInf(val float64, sign int) bool {
	return (sign >= 0 && val > 1.797693134862315708145274237317043567981e+308) ||
		(sign <= 0 && val < -1.797693134862315708145274237317043567981e+308)
}

// GetSensorData retrieves sensor data from the database with optional filtering
func GetSensorData(limit int, startDate, endDate string) ([]SensorData, error) {
	query := "SELECT id, location, recTime, temp, rH, VOC, NOx, pmass1, pmass25, pmass4, pmass10, HCHO, CO2, indoorTd FROM IAQ_SEN55"
	conditions := []string{}
	args := []interface{}{}
	
	// Add date filtering if provided
	if startDate != "" {
		conditions = append(conditions, "recTime >= ?")
		args = append(args, startDate)
	}
	
	if endDate != "" {
		conditions = append(conditions, "recTime <= ?")
		args = append(args, endDate)
	}
	
	// Build the WHERE clause if conditions exist
	if len(conditions) > 0 {
		query += " WHERE "
		for i, condition := range conditions {
			if i > 0 {
				query += " AND "
			}
			query += condition
		}
	}
	
	// Add sorting and limit
	query += " ORDER BY recTime DESC LIMIT ?"
	args = append(args, limit)
	
	// Execute the query
	var data []SensorData
	err := db.Select(&data, query, args...)
	if err != nil {
		return nil, err
	}
	
	// Sanitize and process the data
	for i := range data {
		// Convert location from numeric to string
		if _, err := strconv.Atoi(data[i].Location); err != nil {
			// If it's not a valid integer, keep it as is
		}
		
		// Sanitize the data
		sanitizeSensorData(&data[i])
		
		// Add timestamp field from recTime for consistency with TypeScript API
		data[i].Timestamp = data[i].RecTime.Format(time.RFC3339)
	}
	
	return data, nil
}

// Insert new sensor data 
func InsertSensorData(data SensorData) error {
	// Sanitize the data first
	sanitizeSensorData(&data)
	
	// Convert location string to number if it's a valid number
	locNum, err := strconv.Atoi(data.Location)
	if err != nil {
		locNum = 0 // Default to 0 if not a valid number
	}
	
	// Insert data into database
	_, err = db.Exec(`
		INSERT INTO IAQ_SEN55 (
			location, temp, rH, pmass1, pmass25, pmass4, pmass10, 
			VOC, NOx, HCHO, CO2, indoorTd
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		locNum, data.Temp, data.RH, data.Pmass1, data.Pmass25, data.Pmass4, data.Pmass10,
		data.VOC, data.NOx, data.HCHO, data.CO2, data.IndoorTd,
	)
	
	return err
}

// API handler for getting sensor data
func handleGetSensorData(w http.ResponseWriter, r *http.Request) {
	// Get query parameters
	startDate := r.URL.Query().Get("startDate")
	endDate := r.URL.Query().Get("endDate")
	limitStr := r.URL.Query().Get("limit")
	
	// Default limit to 15000, as in your TypeScript API
	limit := 15000
	if limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}
	
	log.Printf("Fetching data with range: %s to %s, limit: %d", startDate, endDate, limit)
	
	// Get data from database
	data, err := GetSensorData(limit, startDate, endDate)
	if err != nil {
		log.Printf("Error fetching sensor data: %v", err)
		http.Error(w, "Failed to fetch sensor data", http.StatusInternalServerError)
		return
	}
	
	log.Printf("Returned %d records", len(data))
	
	// Add tag support - for demo, just add a tag based on temperature
	for i := range data {
		// Example of adding tags based on sensor values
		if data[i].Temp > 25 {
			data[i].Tags = append(data[i].Tags, "high-temperature")
		} else if data[i].Temp < 18 {
			data[i].Tags = append(data[i].Tags, "low-temperature")
		}
		
		if data[i].CO2 > 1000 {
			data[i].Tags = append(data[i].Tags, "high-co2")
		}
	}
	
	// Set content type header
	w.Header().Set("Content-Type", "application/json")
	
	// Return as JSON
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
	}
}

// API handler for inserting sensor data
func handlePostSensorData(w http.ResponseWriter, r *http.Request) {
	var data SensorData
	
	// Decode the request body
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// Insert data into database
	if err := InsertSensorData(data); err != nil {
		log.Printf("Error inserting sensor data: %v", err)
		http.Error(w, "Failed to insert sensor data", http.StatusInternalServerError)
		return
	}
	
	// Return success response
	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Sensor data inserted successfully",
	})
}

// API handler for getting latest sensor data
func handleGetLatestSensorData(w http.ResponseWriter, r *http.Request) {
	// Get query parameters
	limitStr := r.URL.Query().Get("limit")
	
	// Default limit to 1
	limit := 1
	if limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}
	
	// Get latest data from database
	data, err := GetSensorData(limit, "", "")
	if err != nil {
		log.Printf("Error fetching latest sensor data: %v", err)
		http.Error(w, "Failed to fetch latest sensor data", http.StatusInternalServerError)
		return
	}
	
	// Set content type header
	w.Header().Set("Content-Type", "application/json")
	
	// Return as JSON
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
	}
}

func main() {
	// Initialize database connection
	if err := initDB(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()
	
	// Create router
	r := mux.NewRouter()
	
	// Define API routes
	r.HandleFunc("/api/sensor-data", handleGetSensorData).Methods("GET")
	r.HandleFunc("/api/sensor-data", handlePostSensorData).Methods("POST")
	r.HandleFunc("/api/sensor-data/latest", handleGetLatestSensorData).Methods("GET")
	
	// Add CORS support
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"}, // Adjust this for production
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	})
	
	// Define server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	
	server := &http.Server{
		Addr:    ":" + port,
		Handler: c.Handler(r),
	}
	
	// Start server
	log.Printf("Starting sensor API server on port %s...", port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
