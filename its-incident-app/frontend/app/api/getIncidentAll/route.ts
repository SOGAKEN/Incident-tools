import { cookies } from 'next/headers'
import { NextResponse } from 'next/server'
import type { NextRequest } from 'next/server'

// 環境変数の型チェック
if (!process.env.DBPILOT_URL) {
    throw new Error('DBPILOT_URL is not defined in environment variables')
}

export async function GET(request: NextRequest) {
    const searchParams = request.nextUrl.searchParams
    const cookieStore = await cookies()
    const sessionID = cookieStore.get('session_id')?.value
    const page = parseInt(searchParams.get('page') || '1')
    const limit = parseInt(searchParams.get('limit') || '10')
    const status = searchParams.get('status')
    const statusArray = status ? status.split(',').map((s) => s.trim()) : []

    const from = searchParams.get('from')
    const to = searchParams.get('to')

    const assignee = searchParams.get('assignee')
    const assigneeArray = assignee ? assignee.split(',').map((s) => s.trim()) : []

    const response = await fetch(`${process.env.DBPILOT_URL}/incidents-all`, {
        method: 'POST',
        headers: {
            Authorization: `Bearer ${sessionID}`
        },
        body: JSON.stringify({
            page: page,
            limit: limit,
            status: statusArray,
            from: from,
            to: to,
            assignee: assigneeArray
        })
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
