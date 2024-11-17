import { NextResponse } from 'next/server'

export async function GET() {
    const response = await fetch(`https://bellits.net/v1/workflows/run/05795e88-c57d-416a-95b5-b1248ac30412`, {
        method: 'GET',
        headers: {
            Authorization: `Bearer app-QmNNTUfKyaUV7dmXJBcDuXid`
        }
    })

    if (!response.ok) {
        // 認証サーバーからのエラーレスポンスをそのまま返す
        const errorData = await response.json()
        return NextResponse.json({ error: errorData.message || 'Authentication failed' }, { status: response.status })
    }

    // 認証サーバーからのレスポンスを取得
    const data = await response.json()

    return NextResponse.json(data)
}
