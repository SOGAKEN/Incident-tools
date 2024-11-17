"use client";
import { fetcher } from "@/lib/fetcher";
import { FetchState } from "@/typs/api";
import { useState, useEffect } from "react";

export function useBasicFetch<T>(url: string, init?: RequestInit) {
  const [state, setState] = useState<FetchState<T>>({
    data: null,
    error: null,
    isLoading: true,
  });

  useEffect(() => {
    let mounted = true;

    const fetchData = async () => {
      setState((prev) => ({ ...prev, isLoading: true }));

      try {
        const data = await fetcher<T>(url, init);
        if (mounted) {
          setState({ data, error: null, isLoading: false });
        }
      } catch (error) {
        if (mounted) {
          setState({
            data: null,
            error: error instanceof Error ? error : new Error("Unknown error"),
            isLoading: false,
          });
        }
      }
    };

    fetchData();

    return () => {
      mounted = false;
    };
  }, [url]);

  return state;
}
