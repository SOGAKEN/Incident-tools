import { NextResponse } from 'next/server'
import type { NextRequest } from 'next/server'

// 環境変数の型チェック
if (!process.env.AUTH_URL) {
    throw new Error('AUTH_URL is not defined in environment variables')
}

export async function POST(request: NextRequest) {
    // リクエストボディの取得
    const body = await request.json()

    // 認証サーバーへのリクエスト
    const response = await fetch(`${process.env.AUTH_URL}/accounts`, {
        method: 'POST',
        body: JSON.stringify(body)
    })

    if (!response.ok) {
        const errorData = await response.json()
        return NextResponse.json({ error: errorData.message || 'Authentication failed' }, { status: response.status })
    }

    const data = await response.json()

    return NextResponse.json(data)
}
