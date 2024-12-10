import { cookies } from 'next/headers'
import { NextResponse } from 'next/server'
import type { NextRequest } from 'next/server'

// 環境変数の型チェック
if (!process.env.DBPILOT_URL) {
    throw new Error('DBPILOT_URL is not defined in environment variables')
}

export async function POST(request: NextRequest) {
    // リクエストボディの取得
    const cookieStore = await cookies()
    const sessionID = cookieStore.get('session_id')?.value
    const body = await request.json()

    console.log('body', body)
    // 認証サーバーへのリクエスト
    const response = await fetch(`${process.env.DBPILOT_URL}/incidents/bulk-status`, {
        method: 'POST',
        headers: {
            Authorization: `Bearer ${sessionID}`
        },
        body: JSON.stringify(body)
    })

    if (!response.ok) {
        // 認証サーバーからのエラーレスポンスをそのまま返す
        const errorData = await response.json()
        return NextResponse.json({ error: errorData.message || 'Authentication failed' }, { status: response.status })
    }

    const data = await response.json()

    return NextResponse.json(data)
}
