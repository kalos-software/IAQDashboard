/**
 * Configuration file for the sensor dashboard application
 * This allows for easily changing API endpoints when deploying as a static site
 */

// The base URL for the Go API
// Change this to match your Go API server address
export const API_BASE_URL = 'http://localhost:8080';

// Function to get a full API URL for a given endpoint
export function getApiUrl(endpoint: string): string {
  // Remove leading slash if present to avoid double slashes
  const cleanEndpoint = endpoint.startsWith('/') ? endpoint.substring(1) : endpoint;
  
  // Add the /api/ prefix to match the Go server routes
  const url = `${API_BASE_URL}/api/${cleanEndpoint}`;
  console.log('API URL:', url); // Log the URL being accessed
  return url;
} 