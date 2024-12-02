'use client'

import { useState, useEffect, useCallback } from 'react'
import useSWR, { SWRConfiguration } from 'swr'

export type FetchState<T> = {
    data: T | null
    error: Error | null
    isLoading: boolean
}

export interface UseFetchOptions<T = any> {
    method?: 'GET' | 'POST' | 'PUT' | 'DELETE'
    body?: any
    useSWR?: boolean
    swrOptions?: SWRConfiguration
    fetchOptions?: RequestInit
    onSuccess?: (data: T) => void
    onError?: (error: Error) => void
}

export class FetchError extends Error {
    constructor(
        message: string,
        public status?: number,
        public statusText?: string
    ) {
        super(message)
        this.name = 'FetchError'
    }
}

type FetchUrl = string | null | undefined

const defaultOptions: UseFetchOptions = {
    method: 'GET',
    useSWR: false
}

export function useFetch<T>(url: FetchUrl, options: UseFetchOptions<T> = defaultOptions) {
    // オプションの分割代入を useCallback の外で行い、個別の変数として保持
    const method = options.method || 'GET'
    const useSWRFlag = options.useSWR || false
    const fetchOptions = options.fetchOptions
    const onSuccess = options.onSuccess
    const onError = options.onError
    const body = options.body

    const [state, setState] = useState<FetchState<T>>({
        data: null,
        error: null,
        isLoading: false
    })

    // fetchData の実装を useCallback でラップ
    const fetchData = useCallback(
        async (requestBody?: any): Promise<T> => {
            if (!url) throw new Error('URL is required')

            const response = await fetch(url, {
                method,
                headers: {
                    'Content-Type': 'application/json',
                    ...fetchOptions?.headers
                },
                credentials: 'include',
                body: requestBody ? JSON.stringify(requestBody) : undefined,
                ...fetchOptions
            })

            if (!response.ok) {
                const errorText = await response.text()
                throw new FetchError(errorText || `HTTP error! status: ${response.status}`, response.status, response.statusText)
            }

            return response.json()
        },
        [url, method, fetchOptions]
    )

    // execute の実装を useCallback でラップ
    const execute = useCallback(
        async (executeOptions?: { body?: any }): Promise<T | null> => {
            if (!url) return null

            setState((prev) => ({ ...prev, isLoading: true }))

            try {
                const data = await fetchData(executeOptions?.body ?? body)
                setState({ data, error: null, isLoading: false })
                onSuccess?.(data)
                return data
            } catch (error) {
                const errorObj = error instanceof Error ? error : new Error('Unknown error')
                setState({ data: null, error: errorObj, isLoading: false })
                onError?.(errorObj)
                throw errorObj
            }
        },
        [url, fetchData, body, onSuccess, onError]
    )

    // 自動実行の useEffect
    useEffect(() => {
        if (!url || useSWRFlag || method !== 'GET') return
        execute()
    }, [url, method, useSWRFlag, execute])

    // SWR の実装
    const {
        data: swrData,
        error: swrError,
        isLoading: swrIsLoading,
        mutate: swrMutate
    } = useSWR<T>(url && useSWRFlag && method === 'GET' ? url : null, () => fetchData(body), {
        revalidateOnFocus: false,
        revalidateOnReconnect: false,
        ...options.swrOptions
    })

    // レスポンスの統一
    if (useSWRFlag && method === 'GET' && url) {
        return {
            data: swrData ?? null,
            error: swrError,
            isLoading: swrIsLoading,
            execute,
            mutate: swrMutate
        }
    }

    return {
        ...state,
        execute,
        mutate: async () => {
            if (!url) throw new Error('Cannot mutate with invalid URL')
            return fetchData(body)
        }
    }
}

export type UseFetchResult<T> = {
    data: T | null
    error: Error | null
    isLoading: boolean
    execute: (options?: { body?: any }) => Promise<T | null>
    mutate: () => Promise<T>
}
