export const BASE_URL = "/api/v1";

export interface ApiResponse<T> {
  data: T;
  ok: boolean;
  error?: string;
}

export async function apiGet<T>(path: string): Promise<ApiResponse<T>> {
  try {
    const res = await fetch(`${BASE_URL}${path}`);
    if (!res.ok) {
      return { data: null as T, ok: false, error: `HTTP ${res.status}` };
    }
    const data = await res.json();
    return { data, ok: true };
  } catch (err) {
    return { data: null as T, ok: false, error: String(err) };
  }
}

export async function apiPost<T>(
  path: string,
  body: unknown,
): Promise<ApiResponse<T>> {
  try {
    const res = await fetch(`${BASE_URL}${path}`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(body),
    });
    if (!res.ok) {
      const text = await res.text();
      return {
        data: null as T,
        ok: false,
        error: text || `HTTP ${res.status}`,
      };
    }
    const data = await res.json();
    return { data, ok: true };
  } catch (err) {
    return { data: null as T, ok: false, error: String(err) };
  }
}

export async function apiDelete(path: string): Promise<ApiResponse<void>> {
  try {
    const res = await fetch(`${BASE_URL}${path}`, { method: "DELETE" });
    if (!res.ok && res.status !== 204) {
      return {
        data: undefined as void,
        ok: false,
        error: `HTTP ${res.status}`,
      };
    }
    return { data: undefined as void, ok: true };
  } catch (err) {
    return { data: undefined as void, ok: false, error: String(err) };
  }
}
