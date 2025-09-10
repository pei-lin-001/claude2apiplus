export interface FileUpload {
  id: string;
  name: string;
  type: string;
  size: number;
  url: string;
}

export interface Model {
  id: string;
  object: string;
  created: number;
  owned_by: string;
}

export interface SessionHealth {
  session_key: string;
  org_id: string;
  health_score: number;
  status: number; // 0: active, 1: cooling, 2: failed, 3: circuit_open
  last_used: string;
  last_error: string;
  cooldown_until: string;
  error_count: number;
  success_count: number;
  total_requests: number;
  avg_response_time: number;
  error_types: Record<string, number>;
  weights: number;
  circuit_breaker: {
    state: number;
    failure_count: number;
    last_failure: string;
    next_attempt: string;
    success_count: number;
    threshold: number;
    timeout: number;
  };
}

export interface CallRecord {
  session_key: string;
  timestamp: string;
  success: boolean;
  latency: number;
}

export interface SystemStats {
  total_sessions: number;
  active_sessions: number;
  cooling_sessions: number;
  failed_sessions: number;
  total_requests: number;
  successful_requests: number;
  failed_requests: number;
  average_latency: number;
  uptime: string;
  last_reset: string;
  errors_by_type: Record<string, number>;
  call_records: CallRecord[];
  call_count_by_hour: Record<string, number>;
}

// Helper function to convert status number to string
export function getStatusText(status: number): string {
  switch (status) {
    case 0: return 'active';
    case 1: return 'cooling';
    case 2: return 'failed';
    case 3: return 'circuit_open';
    default: return 'unknown';
  }
}

// Helper function to convert circuit breaker state to string
export function getCircuitBreakerState(state: number): string {
  switch (state) {
    case 0: return 'closed';
    case 1: return 'open';
    case 2: return 'half-open';
    default: return 'unknown';
  }
}