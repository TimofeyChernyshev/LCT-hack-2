import type { TokenPair } from "./types";

const base = import.meta.env.VITE_API_BASE_URL ?? "";

export class ApiError extends Error {
  status: number;
  code?: string;

  constructor(status: number, message: string, code?: string) {
    super(message);
    this.status = status;
    this.code = code;
  }
}

const accessKey = "fsp_access";
const refreshKey = "fsp_refresh";

export function saveTokens(pair: TokenPair) {
  localStorage.setItem(accessKey, pair.accessToken);
  localStorage.setItem(refreshKey, pair.refreshToken);
}

export function clearTokens() {
  localStorage.removeItem(accessKey);
  localStorage.removeItem(refreshKey);
}

export function hasToken() {
  return Boolean(localStorage.getItem(accessKey));
}

let refreshing: Promise<boolean> | null = null;

async function refreshTokens() {
  const refreshToken = localStorage.getItem(refreshKey);
  if (!refreshToken) return false;
  if (!refreshing) {
    refreshing = (async () => {
      try {
        const pair = await api<TokenPair>("auth", "/auth/refresh", {
          method: "POST",
          auth: false,
          body: JSON.stringify({ refreshToken }),
        });
        saveTokens(pair);
        return true;
      } catch {
        clearTokens();
        return false;
      } finally {
        refreshing = null;
      }
    })();
  }
  return refreshing;
}

type Options = RequestInit & { auth?: boolean; retry?: boolean };

export async function api<T>(service: string, path: string, options: Options = {}): Promise<T> {
  const headers = new Headers(options.headers);
  const body = options.body;
  if (body && !(body instanceof FormData) && !headers.has("Content-Type")) {
    headers.set("Content-Type", "application/json");
  }
  if (options.auth !== false) {
    const token = localStorage.getItem(accessKey);
    if (token) headers.set("Authorization", `Bearer ${token}`);
  }

  let response: Response;
  try {
    response = await fetch(`${base}/api/${service}${path}`, { ...options, headers });
  } catch {
    throw new ApiError(0, "Шлюз не ответил. Проверьте, что API запущен на порту 8080.");
  }

  if (response.status === 401 && options.auth !== false && !options.retry) {
    const ok = await refreshTokens();
    if (ok) return api<T>(service, path, { ...options, retry: true });
  }

  if (response.status === 204) return undefined as T;

  const text = await response.text();
  let data: { code?: string; message?: string } | null = null;
  if (text) {
    try {
      data = JSON.parse(text) as { code?: string; message?: string };
    } catch {
      data = null;
    }
  }
  if (!response.ok) {
    throw new ApiError(
      response.status,
      data?.message || "Шлюз не ответил. Проверьте, что API запущен на порту 8080.",
      data?.code,
    );
  }
  return data as T;
}
