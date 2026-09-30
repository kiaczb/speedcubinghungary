import { unstable_rethrow } from "next/navigation";

const LOCAL_API_URL = "http://localhost:8080";

export function serverApiUrl(): string {
  return (
    process.env.SERVER_API_URL || process.env.NEXT_PUBLIC_API_URL || LOCAL_API_URL
  );
}

export function clientApiUrl(): string {
  return process.env.NEXT_PUBLIC_API_URL || LOCAL_API_URL;
}

export class ApiError extends Error {
  constructor(
    readonly status: number,
    message: string,
  ) {
    super(message);
    this.name = "ApiError";
  }
}

async function errorFrom(response: Response): Promise<ApiError> {
  if (response.status === 401 || response.status === 403) {
    return new ApiError(
      response.status,
      "Lejárt a bejelentkezésed — jelentkezz be újra.",
    );
  }

  if (response.status === 429) {
    return new ApiError(
      response.status,
      "Túl sok kérés — várj egy kicsit, és próbáld újra.",
    );
  }

  const message = await response.text().catch(() => "");
  return new ApiError(
    response.status,
    message.trim() || `A kérés sikertelen (${response.status})`,
  );
}

type ClientInit = Omit<RequestInit, "headers"> & {
  headers?: Record<string, string>;
  json?: unknown;
};

export async function clientFetch(
  path: string,
  init: ClientInit = {},
): Promise<Response> {
  const { json, headers, ...rest } = init;

  const requestInit: RequestInit = { credentials: "include", ...rest };
  if (json !== undefined) {
    requestInit.body = JSON.stringify(json);
    requestInit.headers = { "Content-Type": "application/json", ...headers };
  } else if (headers) {
    requestInit.headers = headers;
  }

  const response = await fetch(`${clientApiUrl()}${path}`, requestInit);

  if (!response.ok) throw await errorFrom(response);

  return response;
}

export async function clientJson<T>(
  path: string,
  init: ClientInit = {},
): Promise<T> {
  return (await clientFetch(path, init)).json() as Promise<T>;
}

export async function serverJson<T>(path: string, fallback: T): Promise<T> {
  try {
    const response = await fetch(`${serverApiUrl()}${path}`, {
      cache: "no-store",
    });
    if (!response.ok) return fallback;

    return (await response.json()) as T;
  } catch (error) {
    unstable_rethrow(error);
    console.error(`Nem sikerült elérni a backendet (${path}):`, error);
    return fallback;
  }
}
