use actix_web::{web, App, HttpServer, HttpResponse, Responder, middleware};
use actix_cors::Cors;
use serde::{Serialize, Deserialize};
use sqlx::{mysql::{MySqlPool, MySqlPoolOptions}, FromRow};
use chrono::{DateTime, Utc};
use dotenv::dotenv;
use std::env;
use log::{info, error};
use std::str::FromStr;

// SensorData represents the IAQ_SEN55 sensor data
#[derive(Debug, Serialize, Deserialize, FromRow)]
struct SensorData {
    #[serde(skip_serializing_if = "Option::is_none")]
    id: Option<i32>,
    location: String,
    #[serde(skip)]
    rec_time: Option<DateTime<Utc>>,
    #[serde(rename = "timestamp")]
    #[sqlx(skip)]
    timestamp: Option<String>,
    temp: f64,
    #[serde(rename = "rH")]
    rh: f64,
    #[serde(rename = "VOC")]
    voc: f64,
    #[serde(rename = "NOx")]
    nox: f64,
    pmass1: f64,
    pmass25: f64,
    pmass4: f64,
    pmass10: f64,
    #[serde(rename = "HCHO")]
    hcho: f64,
    #[serde(rename = "CO2")]
    co2: f64,
    indoor_td: f64,
    #[serde(skip_serializing_if = "Option::is_none")]
    tags: Option<Vec<String>>,
}

// Database configuration
struct DbConfig {
    host: String,
    user: String,
    password: String,
    database: String,
    port: u16,
}

impl Default for DbConfig {
    fn default() -> Self {
        Self {
            host: "localhost".to_string(),
            user: "data".to_string(),
            password: "mysqldatapassword".to_string(),
            database: "buildingData".to_string(),
            port: 3306,
        }
    }
}

// App state
struct AppState {
    db_pool: MySqlPool,
}

// Initialize database connection
async fn init_db() -> Result<MySqlPool, sqlx::Error> {
    let config = DbConfig::default();
    
    let connection_string = format!(
        "mysql://{}:{}@{}:{}/{}",
        config.user, config.password, config.host, config.port, config.database
    );
    
    MySqlPoolOptions::new()
        .max_connections(25)
        .min_connections(5)
        .max_lifetime(std::time::Duration::from_secs(300))
        .connect(&connection_string)
        .await
}

// Sanitize sensor data to handle NaN and Infinity values
fn sanitize_sensor_data(data: &mut SensorData) {
    // Helper function to sanitize float values
    let sanitize_float = |val: &mut f64| {
        if !val.is_finite() {
            *val = 0.0;
        }
    };
    
    sanitize_float(&mut data.temp);
    sanitize_float(&mut data.rh);
    sanitize_float(&mut data.voc);
    sanitize_float(&mut data.nox);
    sanitize_float(&mut data.pmass1);
    sanitize_float(&mut data.pmass25);
    sanitize_float(&mut data.pmass4);
    sanitize_float(&mut data.pmass10);
    sanitize_float(&mut data.hcho);
    sanitize_float(&mut data.co2);
    sanitize_float(&mut data.indoor_td);
}

// Get sensor data with optional filtering
async fn get_sensor_data(
    db_pool: &MySqlPool,
    limit: i64,
    start_date: Option<String>,
    end_date: Option<String>,
) -> Result<Vec<SensorData>, sqlx::Error> {
    // Build query dynamically
    let mut query = String::from(
        "SELECT id, location, recTime, temp, rH, VOC, NOx, pmass1, pmass25, pmass4, pmass10, HCHO, CO2, indoorTd \
        FROM IAQ_SEN55"
    );
    
    let mut conditions = Vec::new();
    let mut args = Vec::new();
    
    if let Some(start) = &start_date {
        conditions.push("recTime >= ?");
        args.push(start.clone());
    }
    
    if let Some(end) = &end_date {
        conditions.push("recTime <= ?");
        args.push(end.clone());
    }
    
    if !conditions.is_empty() {
        query.push_str(" WHERE ");
        query.push_str(&conditions.join(" AND "));
    }
    
    query.push_str(" ORDER BY recTime DESC LIMIT ?");
    
    // Execute query
    let mut query_builder = sqlx::query_as::<_, SensorData>(&query);
    
    for arg in args {
        query_builder = query_builder.bind(arg);
    }
    query_builder = query_builder.bind(limit);
    
    let mut data = query_builder.fetch_all(db_pool).await?;
    
    // Process the data
    for item in &mut data {
        // Sanitize data
        sanitize_sensor_data(item);
        
        // Add timestamp from rec_time
        if let Some(rec_time) = item.rec_time {
            item.timestamp = Some(rec_time.to_rfc3339());
        }
        
        // Add tags based on sensor values
        let mut tags = Vec::new();
        if item.temp > 25.0 {
            tags.push("high-temperature".to_string());
        } else if item.temp < 18.0 {
            tags.push("low-temperature".to_string());
        }
        
        if item.co2 > 1000.0 {
            tags.push("high-co2".to_string());
        }
        
        if !tags.is_empty() {
            item.tags = Some(tags);
        }
    }
    
    Ok(data)
}

// Insert new sensor data
async fn insert_sensor_data(db_pool: &MySqlPool, mut data: SensorData) -> Result<(), sqlx::Error> {
    // Sanitize the data first
    sanitize_sensor_data(&mut data);
    
    // Convert location string to number if it's a valid number
    let loc_num = i32::from_str(&data.location).unwrap_or(0);
    
    // Insert data into database
    sqlx::query(
        "INSERT INTO IAQ_SEN55 (location, temp, rH, pmass1, pmass25, pmass4, pmass10, VOC, NOx, HCHO, CO2, indoorTd) \
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
    )
    .bind(loc_num)
    .bind(data.temp)
    .bind(data.rh)
    .bind(data.pmass1)
    .bind(data.pmass25)
    .bind(data.pmass4)
    .bind(data.pmass10)
    .bind(data.voc)
    .bind(data.nox)
    .bind(data.hcho)
    .bind(data.co2)
    .bind(data.indoor_td)
    .execute(db_pool)
    .await?;
    
    Ok(())
}

// API handler for getting sensor data
async fn handle_get_sensor_data(
    query: web::Query<GetSensorDataQuery>,
    data: web::Data<AppState>,
) -> impl Responder {
    // Get query parameters
    let limit = query.limit.unwrap_or(15000);
    
    info!(
        "Fetching data with range: {:?} to {:?}, limit: {}",
        query.start_date, query.end_date, limit
    );
    
    // Get data from database
    match get_sensor_data(&data.db_pool, limit, query.start_date.clone(), query.end_date.clone()).await {
        Ok(sensor_data) => {
            info!("Returned {} records", sensor_data.len());
            HttpResponse::Ok().json(sensor_data)
        }
        Err(e) => {
            error!("Error fetching sensor data: {}", e);
            HttpResponse::InternalServerError().body("Failed to fetch sensor data")
        }
    }
}

// API handler for inserting sensor data
async fn handle_post_sensor_data(
    sensor_data: web::Json<SensorData>,
    data: web::Data<AppState>,
) -> impl Responder {
    // Insert data into database
    match insert_sensor_data(&data.db_pool, sensor_data.into_inner()).await {
        Ok(_) => HttpResponse::Created().json(serde_json::json!({
            "message": "Sensor data inserted successfully"
        })),
        Err(e) => {
            error!("Error inserting sensor data: {}", e);
            HttpResponse::InternalServerError().body("Failed to insert sensor data")
        }
    }
}

// API handler for getting latest sensor data
async fn handle_get_latest_sensor_data(
    query: web::Query<LatestSensorDataQuery>,
    data: web::Data<AppState>,
) -> impl Responder {
    // Default limit to 1
    let limit = query.limit.unwrap_or(1);
    
    // Get latest data from database
    match get_sensor_data(&data.db_pool, limit, None, None).await {
        Ok(sensor_data) => HttpResponse::Ok().json(sensor_data),
        Err(e) => {
            error!("Error fetching latest sensor data: {}", e);
            HttpResponse::InternalServerError().body("Failed to fetch latest sensor data")
        }
    }
}

// Query parameters for GET /api/sensor-data
#[derive(Debug, Deserialize)]
struct GetSensorDataQuery {
    start_date: Option<String>,
    end_date: Option<String>,
    limit: Option<i64>,
}

// Query parameters for GET /api/sensor-data/latest
#[derive(Debug, Deserialize)]
struct LatestSensorDataQuery {
    limit: Option<i64>,
}

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    // Load environment variables
    dotenv().ok();
    
    // Initialize logger
    env_logger::init_from_env(env_logger::Env::default().default_filter_or("info"));
    
    // Initialize database connection
    let db_pool = init_db().await.expect("Failed to connect to database");
    
    // Create app state
    let app_state = web::Data::new(AppState { db_pool });
    
    // Get port from environment
    let port = env::var("PORT").unwrap_or_else(|_| "8080".to_string());
    let bind_address = format!("0.0.0.0:{}", port);
    
    info!("Starting sensor API server on {}...", bind_address);
    
    // Start server
    HttpServer::new(move || {
        // Set up CORS
        let cors = Cors::default()
            .allow_any_origin()
            .allowed_methods(vec!["GET", "POST", "PUT", "DELETE"])
            .allowed_headers(vec!["Content-Type", "Authorization"])
            .supports_credentials();
        
        App::new()
            .wrap(middleware::Logger::default())
            .wrap(cors)
            .app_data(app_state.clone())
            .route("/api/sensor-data", web::get().to(handle_get_sensor_data))
            .route("/api/sensor-data", web::post().to(handle_post_sensor_data))
            .route("/api/sensor-data/latest", web::get().to(handle_get_latest_sensor_data))
    })
    .bind(bind_address)?
    .run()
    .await
} 