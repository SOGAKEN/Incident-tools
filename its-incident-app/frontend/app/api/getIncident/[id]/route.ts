import { NextRequest, NextResponse } from 'next/server'

// 環境変数の型チェック
if (!process.env.DBPILOT_URL) {
    throw new Error('DBPILOT_URL is not defined in environment variables')
}

export async function GET(request: NextRequest) {
    const sessionID = request.cookies.get('session_id')?.value

    // URL から ID を取得
    const idMatch = request.nextUrl.pathname.match(/\/api\/getIncident\/(.+)$/)
    const id = idMatch ? idMatch[1] : null

    if (!id) {
        return NextResponse.json({ error: 'ID parameter is missing' }, { status: 400 })
    }

    const response = await fetch(`${process.env.DBPILOT_URL}/email/${id}`, {
        method: 'GET',
        headers: {
            Authorization: `Bearer ${sessionID}`
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
