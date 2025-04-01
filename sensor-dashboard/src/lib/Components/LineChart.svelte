<script lang="ts">
  import { VisXYContainer, VisLine, VisAxis, VisTooltip } from '@unovis/svelte';
  import { Line } from '@unovis/ts';
  
  // Define the data type expected
  type DataRecord = {
    x: number;
    y: number;
  };
  
  // Define axis options type
  type AxisOptions = {
    minValue?: number;
    maxValue?: number;
    tickFormat?: (tick: any) => string;
    numTicks?: number;
    gridLine?: boolean;
    tickLine?: boolean;
    tickTextColor?: string;
    tickTextFontSize?: string;
  };
  
  // Accept data as a prop
  export let data: DataRecord[] = [];
  export let height: number = 300;
  export let label: string = '';
  export let color: string = '#1d4ed8';
  export let showLegend: boolean = true;
  
  // Axis configuration
  export let yAxisOptions: AxisOptions = {};
  export let xAxisOptions: AxisOptions = {};
  
  // Set default axis options based on label
  $: {
    // Default Y-axis options based on metric type
    if (label.includes('Temperature')) {
      yAxisOptions = {
        ...{
          minValue: undefined,
          maxValue: undefined,
          numTicks: 5,
          gridLine: true,
          tickFormat: (d) => `${d}°F`
        },
        ...yAxisOptions
      };
    } else if (label.includes('CO2')) {
      yAxisOptions = {
        ...{
          minValue: 0,
          maxValue: undefined,
          numTicks: 5,
          gridLine: true,
          tickFormat: (d) => `${d} ppm`
        },
        ...yAxisOptions
      };
    } else if (label.includes('Humidity')) {
      yAxisOptions = {
        ...{
          minValue: 0,
          maxValue: 100,
          numTicks: 5,
          gridLine: true,
          tickFormat: (d) => `${d}%`
        },
        ...yAxisOptions
      };
    } else if (label.includes('VOC')) {
      yAxisOptions = {
        ...{
          minValue: 0,
          maxValue: undefined,
          numTicks: 5,
          gridLine: true,
          tickFormat: (d) => `${d} ppb`
        },
        ...yAxisOptions
      };
    } else if (label.includes('PM2.5')) {
      yAxisOptions = {
        ...{
          minValue: 0,
          maxValue: undefined,
          numTicks: 5,
          gridLine: true,
          tickFormat: (d) => `${d} μg/m³`
        },
        ...yAxisOptions
      };
    }
    
    // Default X-axis options (time-based)
    xAxisOptions = {
      ...{
        numTicks: 5,
        gridLine: true,
        tickFormat: (d) => {
          const date = new Date(d);
          return date.toLocaleString(undefined, { 
            month: 'short', 
            day: 'numeric', 
            hour: '2-digit', 
            minute: '2-digit'
          });
        }
      },
      ...xAxisOptions
    };
  }
  
  // Define accessors
  const x = (d: DataRecord) => d.x;
  const y = (d: DataRecord) => d.y;
  
  // Configure tooltip
  const tooltipConfig = {
    trigger: 'hover',
    content: (d: any) => {
      if (!d || !d[0]) return '';
      const datum = d[0].datum;
      let formattedValue = datum.y.toFixed(2);
      
      // Format the value based on the metric type
      if (label.includes('Temperature')) {
        formattedValue = `${formattedValue}°F`;
      } else if (label.includes('CO2')) {
        formattedValue = `${formattedValue} ppm`;
      } else if (label.includes('Humidity')) {
        formattedValue = `${formattedValue}%`;
      } else if (label.includes('VOC')) {
        formattedValue = `${formattedValue} ppb`;
      } else if (label.includes('PM2.5')) {
        formattedValue = `${formattedValue} μg/m³`;
      }
      
      return `
        <div style="padding: 8px;">
          <div style="font-weight: bold;">${new Date(datum.x).toLocaleString()}</div>
          <div>${label}: ${formattedValue}</div>
        </div>
      `;
    }
  };
</script>

<div class="chart-container">
  <div style="height: {height}px; width: 100%;">
    {#if data.length === 0}
      <div class="flex justify-center items-center h-full">
        <p class="text-gray-500">No data available</p>
      </div>
    {:else}
      <VisXYContainer {data} height={height}>
        <VisLine 
          {x} 
          {y} 
          curveType="monotoneX"
          lineWidth={2}
          {color}
          highlightOnHover={true}
          attributes={{
            [Line.selectors.line]: {
              'data-testid': 'line-chart',
              'data-metric': label
            }
          }}
        />
        <VisAxis 
          position="bottom" 
          label="Time" 
          type="x"
          numTicks={xAxisOptions.numTicks}
          gridLine={xAxisOptions.gridLine}
          tickFormat={(x: number) => {
            const date = new Date(x);
            return date.toLocaleString(undefined, { 
              month: 'short', 
              day: 'numeric', 
              hour: '2-digit', 
            });
          }}
          tickTextColor={xAxisOptions.tickTextColor}
          tickTextFontSize={xAxisOptions.tickTextFontSize}
        />
        <VisAxis 
          position="left" 
          label={label}
          type="y"
          numTicks={yAxisOptions.numTicks}
          gridLine={yAxisOptions.gridLine}
          tickFormat={yAxisOptions.tickFormat}
          tickTextColor={yAxisOptions.tickTextColor}
          tickTextFontSize={yAxisOptions.tickTextFontSize}
        />
        <VisTooltip config={tooltipConfig} />
      </VisXYContainer>
    {/if}
  </div>
  
  {#if showLegend && label}
    <div class="legend-container">
      <div class="legend-item">
        <span class="legend-color" style="background-color: {color}"></span>
        <span class="legend-name">{label}</span>
      </div>
    </div>
  {/if}
</div>

<style>
  .chart-container {
    position: relative;
  }
  
  .legend-container {
    display: flex;
    justify-content: center;
    margin-top: 8px;
  }
  
  .legend-item {
    display: flex;
    align-items: center;
    margin: 0 8px;
  }
  
  .legend-color {
    display: inline-block;
    width: 12px;
    height: 12px;
    border-radius: 2px;
    margin-right: 4px;
  }
  
  .legend-name {
    font-size: 0.875rem;
    color: #4b5563;
  }
  
  :global(.unovis .vis-axis text) {
    fill: #4b5563;
  }
  
  :global(.unovis .vis-axis line) {
    stroke: #e5e7eb;
  }
  
  :global(.unovis .vis-axis path) {
    stroke: #d1d5db;
  }
  
  :global(.unovis .vis-tooltip) {
    background-color: rgba(255, 255, 255, 0.95);
    border: 1px solid #e5e7eb;
    border-radius: 4px;
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  }
</style>