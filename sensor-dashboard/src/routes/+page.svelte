<script lang="ts">
  import { onMount } from 'svelte';
  import LineChart from '$lib/Components/LineChart.svelte';
  import type { NormalizedSensorData } from '$lib/models';
  import { getApiUrl } from '$lib/config';
  
  // Define data record type that matches the LineChart component
  type DataRecord = {
    x: number;
    y: number;
  };
  
  // State variables
  let sensorData: NormalizedSensorData[] = [];
  let tempChart: DataRecord[] = [];
  let co2Chart: DataRecord[] = [];
  let humidityChart: DataRecord[] = [];
  let vocChart: DataRecord[] = [];
  let pmassChart: DataRecord[] = [];
  let loading = true;
  let error: string | null = null;
  
  // Date range filters
  let startDate: string = '';
  let endDate: string = '';
  let startHour: number = 0;
  let endHour: number = 23;
  
  // Time formatter options for X-axis
  function getTimeFormat() {
    if (!tempChart.length) return undefined;
    
    // Calculate time span in hours
    const firstTimestamp = Math.min(...tempChart.map(d => d.x));
    const lastTimestamp = Math.max(...tempChart.map(d => d.x));
    const timeSpanHours = (lastTimestamp - firstTimestamp) / (1000 * 60 * 60);
    
    // Choose appropriate time format based on data range
    if (timeSpanHours <= 24) {
      // For 24 hours or less, show hour:minute
      return (d: any) => {
        const date = new Date(d);
        return date.toLocaleTimeString(undefined, { hour: '2-digit', minute: '2-digit' });
      };
    } else if (timeSpanHours <= 72) {
      // For 3 days or less, show day and time
      return (d: any) => {
        const date = new Date(d);
        return date.toLocaleString(undefined, { 
          weekday: 'short',
          hour: '2-digit', 
          minute: '2-digit'
        });
      };
    } else {
      // For longer periods, show month/day and hour
      return (d: any) => {
        const date = new Date(d);
        return date.toLocaleString(undefined, { 
          month: 'short', 
          day: 'numeric',
          hour: '2-digit'
        });
      };
    }
  }
  
  // Reactive axis options
  $: xAxisOptions = {
    tickFormat: getTimeFormat(),
    numTicks: 5,
    gridLine: true
  };
  
  $: tempAxisOptions = {
    minValue: tempChart.length ? Math.min(...tempChart.map(d => d.y).filter(v => !isNaN(v))) - 5 : 60,
    maxValue: tempChart.length ? Math.max(...tempChart.map(d => d.y).filter(v => !isNaN(v))) + 5 : 80,
  };
  
  $: co2AxisOptions = {
    minValue: 0,
    maxValue: co2Chart.length ? Math.max(...co2Chart.map(d => d.y).filter(v => !isNaN(v))) + 100 : 2000,
  };
  
  $: vocAxisOptions = {
    minValue: 0,
    maxValue: vocChart.length ? Math.max(...vocChart.map(d => d.y).filter(v => !isNaN(v))) * 1.2 : 1000,
  };
  
  $: pmAxisOptions = {
    minValue: 0,
    maxValue: pmassChart.length ? Math.max(...pmassChart.map(d => d.y).filter(v => !isNaN(v))) * 1.2 : 50,
  };
  
  // Set default date range to last 24 hours
  function setDefaultDateRange() {
    const now = new Date();
    const yesterday = new Date(now);
    yesterday.setDate(now.getDate() - 1);
    
    endDate = now.toISOString().split('T')[0];
    startDate = yesterday.toISOString().split('T')[0];
    startHour = 0;
    endHour = 23;
  }
  
  // Format hour for display
  function formatHour(hour: number): string {
    return hour.toString().padStart(2, '0') + ':00';
  }
  
  // Get hours array for selector
  function getHours(): {value: number, label: string}[] {
    const hours = [];
    for (let i = 0; i <= 23; i++) {
      hours.push({
        value: i,
        label: formatHour(i)
      });
    }
    return hours;
  }
  
  // Hours for selection
  const hours = getHours();
  
  // Fetch data with the current date range
  async function fetchData() {
    loading = true;
    error = null;
    
    try {
      // Build URL with date parameters - using the Go API endpoint
      const url = new URL(getApiUrl('sensor-data'));
      
      // Add date parameters with hour precision
      if (startDate) {
        const formattedStartDate = `${startDate}T${startHour.toString().padStart(2, '0')}:00:00`;
        url.searchParams.append('startDate', formattedStartDate);
      }
      
      if (endDate) {
        const formattedEndDate = `${endDate}T${endHour.toString().padStart(2, '0')}:59:59`;
        url.searchParams.append('endDate', formattedEndDate);
      }
      
      const response = await fetch(url);
      if (!response.ok) {
        throw new Error(`HTTP error: ${response.status}`);
      }
      
      // Get the normalized data from the API
      sensorData = await response.json();
      
      // Create chart datasets
      updateCharts();
      
      loading = false;
    } catch (err) {
      console.error('Error fetching data:', err);
      error = err instanceof Error ? err.message : 'Unknown error';
      loading = false;
    }
  }
  
  // Handle date input changes
  function handleDateChange() {
    fetchData();
  }
  
  // Reset date filters and fetch all data
  function resetDateFilters() {
    startDate = '';
    endDate = '';
    startHour = 0;
    endHour = 23;
    fetchData();
  }
  
  // Update all chart data
  function updateCharts() {
    tempChart = createTimeSeriesData(sensorData, 'temp');
    co2Chart = createTimeSeriesData(sensorData, 'CO2');
    humidityChart = createTimeSeriesData(sensorData, 'rH');
    vocChart = createTimeSeriesData(sensorData, 'VOC');
    pmassChart = createTimeSeriesData(sensorData, 'pmass25');
  }
  
  function createTimeSeriesData(data: NormalizedSensorData[], field: keyof NormalizedSensorData): DataRecord[] {
    return data.map(d => ({
      x: new Date(d.recTime).getTime(),
      y: Number(d[field])
    }));
  }
  
  // Initial load
  onMount(() => {
    setDefaultDateRange();
    fetchData();
  });
</script>

<main class="container mx-auto p-4 max-w-4xl">
  <h1 class="text-3xl font-bold mb-6">Sensor Dashboard</h1>

  <div class="bg-white shadow rounded-lg p-4 mb-6">
    <h2 class="text-xl font-semibold mb-2">Filter by Date and Time Range</h2>
    <div class="grid grid-cols-1 md:grid-cols-2 gap-4 mb-4">
      <div>
        <div class="mb-2 font-medium">Start:</div>
        <div class="flex flex-wrap items-center gap-2">
          <div class="flex items-center">
            <label for="start-date" class="mr-2">Date:</label>
            <input 
              type="date" 
              id="start-date" 
              bind:value={startDate} 
              on:change={handleDateChange}
              class="border rounded px-2 py-1"
            />
          </div>
          <div class="flex items-center">
            <label for="start-hour" class="mr-2">Hour:</label>
            <select 
              id="start-hour" 
              bind:value={startHour} 
              on:change={handleDateChange}
              class="border rounded px-2 py-1"
            >
              {#each hours as hour}
                <option value={hour.value}>{hour.label}</option>
              {/each}
            </select>
          </div>
        </div>
      </div>
      
      <div>
        <div class="mb-2 font-medium">End:</div>
        <div class="flex flex-wrap items-center gap-2">
          <div class="flex items-center">
            <label for="end-date" class="mr-2">Date:</label>
            <input 
              type="date" 
              id="end-date" 
              bind:value={endDate} 
              on:change={handleDateChange}
              class="border rounded px-2 py-1"
            />
          </div>
          <div class="flex items-center">
            <label for="end-hour" class="mr-2">Hour:</label>
            <select 
              id="end-hour" 
              bind:value={endHour} 
              on:change={handleDateChange}
              class="border rounded px-2 py-1"
            >
              {#each hours as hour}
                <option value={hour.value}>{hour.label}</option>
              {/each}
            </select>
          </div>
        </div>
      </div>
    </div>
    
    <div class="flex flex-wrap gap-2">
      <button 
        on:click={resetDateFilters}
        class="bg-gray-200 hover:bg-gray-300 text-gray-800 px-3 py-1 rounded"
      >
        Reset
      </button>
      <button 
        on:click={() => { setDefaultDateRange(); handleDateChange(); }}
        class="bg-blue-500 hover:bg-blue-600 text-white px-3 py-1 rounded"
      >
        Last 24 Hours
      </button>
      <button 
        on:click={() => { 
          const now = new Date();
          endDate = now.toISOString().split('T')[0];
          endHour = now.getHours();
          startDate = endDate;
          startHour = Math.max(0, endHour - 6);
          handleDateChange(); 
        }}
        class="bg-green-500 hover:bg-green-600 text-white px-3 py-1 rounded"
      >
        Last 6 Hours
      </button>
      <button 
        on:click={() => { 
          const now = new Date();
          endDate = now.toISOString().split('T')[0];
          endHour = now.getHours();
          startDate = endDate;
          startHour = Math.max(0, endHour - 1);
          handleDateChange(); 
        }}
        class="bg-indigo-500 hover:bg-indigo-600 text-white px-3 py-1 rounded"
      >
        Last Hour
      </button>
    </div>
    
    {#if sensorData.length === 0 && !loading}
      <p class="mt-4 text-amber-600">No data available for the selected date range.</p>
    {/if}
  </div>

  {#if loading}
    <div class="flex justify-center items-center h-40 bg-white shadow rounded-lg">
      <div class="animate-pulse text-lg">Loading data...</div>
    </div>
  {:else if error}
    <div class="bg-red-50 border border-red-200 text-red-700 p-4 rounded mb-6">
      <p><strong>Error:</strong> {error}</p>
    </div>
  {:else if sensorData.length === 0}
    <div class="bg-amber-50 border border-amber-200 text-amber-700 p-4 rounded mb-6">
      <p>No data available for the selected date range.</p>
    </div>
  {:else}
    <div class="mb-4 text-sm text-gray-600">
      Showing {sensorData.length} data points from {sensorData.length > 0 ? new Date(sensorData[sensorData.length - 1].recTime).toLocaleString() : ''} to {sensorData.length > 0 ? new Date(sensorData[0].recTime).toLocaleString() : ''}
    </div>
    
    <div class="space-y-6">
      <section class="bg-white shadow rounded-lg p-4">
        <h2 class="text-xl font-semibold mb-2">Temperature</h2>
        <LineChart 
          data={tempChart} 
          label="Temperature (°F)" 
          color="#ef4444"
          yAxisOptions={tempAxisOptions}
          xAxisOptions={xAxisOptions}
        />
      </section>
      
      <section class="bg-white shadow rounded-lg p-4">
        <h2 class="text-xl font-semibold mb-2">CO2</h2>
        <LineChart 
          data={co2Chart} 
          label="CO2 (ppm)" 
          color="#10b981"
          yAxisOptions={co2AxisOptions}
          xAxisOptions={xAxisOptions}
        />
      </section>
      
      <section class="bg-white shadow rounded-lg p-4">
        <h2 class="text-xl font-semibold mb-2">Humidity</h2>
        <LineChart 
          data={humidityChart} 
          label="Humidity (%)" 
          color="#3b82f6"
          xAxisOptions={xAxisOptions}
        />
      </section>
      
      <section class="bg-white shadow rounded-lg p-4">
        <h2 class="text-xl font-semibold mb-2">VOC</h2>
        <LineChart 
          data={vocChart} 
          label="VOC (ppb)" 
          color="#8b5cf6"
          yAxisOptions={vocAxisOptions}
          xAxisOptions={xAxisOptions}
        />
      </section>
      
      <section class="bg-white shadow rounded-lg p-4">
        <h2 class="text-xl font-semibold mb-2">Particulate Matter</h2>
        <LineChart 
          data={pmassChart} 
          label="PM2.5 (μg/m³)" 
          color="#f59e0b"
          yAxisOptions={pmAxisOptions}
          xAxisOptions={xAxisOptions}
        />
      </section>
    </div>
  {/if}
</main>

<style>
  :global(body) {
    background-color: #f9fafb;
    color: #111827;
  }
  
  .animate-pulse {
    animation: pulse 2s cubic-bezier(0.4, 0, 0.6, 1) infinite;
  }
  
  @keyframes pulse {
    0%, 100% {
      opacity: 1;
    }
    50% {
      opacity: 0.5;
    }
  }
</style>

