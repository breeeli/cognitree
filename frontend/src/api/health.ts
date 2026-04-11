import { apiGet } from "./client";

export interface HealthStatus {
  status: string;
  timestamp: string;
  database: string;
}

export function getHealth() {
  return apiGet<HealthStatus>("/health");
}
