// Parse Go duration format and convert to seconds with 2 decimal places
export function formatDuration(duration: string | null | undefined): string {
  if (!duration) return '-';
  
  // Handle Go duration formats: "1.5s", "100ms", "1m30s", "456.72ms", etc.
  if (duration.includes('ms')) {
    // Convert milliseconds to seconds
    const ms = parseFloat(duration.replace('ms', ''));
    return `${(ms / 1000).toFixed(2)}s`;
  } else if (duration.includes('µs') || duration.includes('us')) {
    // Convert microseconds to seconds
    const us = parseFloat(duration.replace(/[µu]s/, ''));
    return `${(us / 1000000).toFixed(2)}s`;
  } else if (duration.includes('m') && duration.includes('s')) {
    // Handle "1m30s" format
    const parts = duration.match(/(\d+\.?\d*)m(\d+\.?\d*)s/);
    if (parts) {
      const minutes = parseFloat(parts[1]);
      const seconds = parseFloat(parts[2]);
      return `${(minutes * 60 + seconds).toFixed(2)}s`;
    }
  } else if (duration.endsWith('s')) {
    // Already in seconds
    return `${parseFloat(duration.replace('s', '')).toFixed(2)}s`;
  }
  
  return duration; // Fallback to original
}

