"use client";

import { useState, useEffect } from "react";
import useSWR, { SWRConfiguration } from "swr";

export type FetchState<T> = {
  data: T | null;
  error: Error | null;
  isLoading: boolean;
};

export interface UseFetchOptions<T = any> {
  method?: "GET" | "POST" | "PUT" | "DELETE";
  body?: any;
  useSWR?: boolean;
  swrOptions?: SWRConfiguration;
  fetchOptions?: RequestInit;
  onSuccess?: (data: T) => void;
  onError?: (error: Error) => void;
}

export class FetchError extends Error {
  constructor(
    message: string,
    public status?: number,
    public statusText?: string,
  ) {
    super(message);
    this.name = "FetchError";
  }
}

const defaultOptions: UseFetchOptions = {
  method: "GET",
  useSWR: false,
};

export function useFetch<T>(
  url: string,
  options: UseFetchOptions<T> = defaultOptions,
) {
  const {
    method = "GET",
    body,
    useSWR: useSWRFlag = false,
    swrOptions,
    fetchOptions,
    onSuccess,
    onError,
  } = options;

  // Basic Fetchの状態管理
  const [state, setState] = useState<FetchState<T>>({
    data: null,
    error: null,
    isLoading: method === "GET", // GETの場合は初期ローディング
  });

  // 共通のfetch処理
  const fetchData = async (requestBody?: any): Promise<T> => {
    const response = await fetch(url, {
      method,
      headers: {
        "Content-Type": "application/json",
        ...fetchOptions?.headers,
      },
      credentials: "include",
      body: requestBody ? JSON.stringify(requestBody) : undefined,
      ...fetchOptions,
    });

    if (!response.ok) {
      const errorText = await response.text();
      throw new FetchError(
        errorText || `HTTP error! status: ${response.status}`,
        response.status,
        response.statusText,
      );
    }

    return response.json();
  };

  // 手動実行用の関数
  const execute = async (executeOptions?: {
    body?: any;
  }): Promise<T | null> => {
    setState((prev) => ({ ...prev, isLoading: true }));

    try {
      const data = await fetchData(executeOptions?.body ?? body);
      setState({ data, error: null, isLoading: false });
      onSuccess?.(data);
      return data;
    } catch (error) {
      const errorObj =
        error instanceof Error ? error : new Error("Unknown error");
      setState({ data: null, error: errorObj, isLoading: false });
      onError?.(errorObj);
      throw errorObj;
    }
  };

  // GETリクエストの自動実行
  useEffect(() => {
    if (useSWRFlag || method !== "GET") return;
    execute();
  }, [url, method, useSWRFlag]);

  // SWRの使用（GETリクエストのみ）
  const swr = useSWR<T>(
    useSWRFlag && method === "GET" ? url : null,
    () => fetchData(body),
    {
      revalidateOnFocus: false,
      revalidateOnReconnect: false,
      ...swrOptions,
    },
  );

  // レスポンスの統一
  if (useSWRFlag && method === "GET") {
    return {
      data: swr.data ?? null,
      error: swr.error,
      isLoading: swr.isLoading,
      mutate: swr.mutate,
      execute,
    };
  }

  return {
    ...state,
    execute,
    mutate: async () => {
      setState((prev) => ({ ...prev, isLoading: true }));
      const data = await fetchData(body);
      setState({ data, error: null, isLoading: false });
      return data;
    },
  };
}

// 型の補助
export type UseFetchResult<T> = {
  data: T | null;
  error: Error | null;
  isLoading: boolean;
  execute: (options?: { body?: any }) => Promise<T | null>;
  mutate: () => Promise<T>;
};
