/**
 * IAQSen55 sensor data model that matches exactly what comes from the API
 */
export interface RawSensorData {
  id: number;
  location: string;
  recTime: string;
  timestamp: string;
  temp: string | number;
  rH: number;
  VOC: number;
  NOx: string | number;
  pmass1: number;
  pmass25: number;
  pmass4: number;
  pmass10: number;
  HCHO: string | number;
  CO2: number;
  indoorTd: string | number;
}

/**
 * Normalized sensor data with consistent types
 */
export interface NormalizedSensorData {
  id: number;
  location: string;
  recTime: string;
  timestamp: string;
  temp: number;
  rH: number;
  VOC: number;
  NOx: number;
  pmass1: number;
  pmass25: number;
  pmass4: number;
  pmass10: number;
  HCHO: number;
  CO2: number;
  indoorTd: number;
} 