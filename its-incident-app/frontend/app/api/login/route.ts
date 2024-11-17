import { NextResponse } from 'next/server'
import type { NextRequest } from 'next/server'

// 環境変数の型チェック
if (!process.env.AUTH_URL) {
    throw new Error('AUTH_URL is not defined in environment variables')
}

export async function POST(request: NextRequest) {
    try {
        // リクエストボディの取得
        const body = await request.json()

        // 認証サーバーへのリクエスト
        const response = await fetch(`${process.env.AUTH_URL}/login`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(body)
        })

        if (response.status === 401) {
            return NextResponse.json({ status: response.status })
        }

        if (!response.ok) {
            const errorData = await response.json()
            return NextResponse.json({ error: errorData.message || 'Authentication failed' }, { status: response.status })
        }

        // 認証サーバーからのレスポンスを取得
        const data = await response.json()

        // 認証サーバーからのSet-Cookieヘッダーを取得
        const setCookieHeader = response.headers.get('set-cookie')

        // クライアントへのレスポンスを作成
        const res = NextResponse.json(data)

        // Set-Cookieヘッダーをクライアントに送信
        if (setCookieHeader) {
            res.headers.set('set-cookie', setCookieHeader)
        }

        return res
    } catch (error) {
        // エラーハンドリング
        console.error('Login error:', error)
        return NextResponse.json({ error: 'Internal server error' }, { status: 500 })
    }
}
