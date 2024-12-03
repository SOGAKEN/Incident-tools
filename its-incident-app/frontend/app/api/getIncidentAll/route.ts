// /frontend/app/api/getIncidentAll/route.ts

import { cookies } from 'next/headers'
import { NextResponse } from 'next/server'
import type { NextRequest } from 'next/server'

// 新しい型定義を追加
interface IncidentStatus {
    id: number
    code: number
    name: string
    description: string
    is_active: boolean
    display_order: number
}

interface Incident {
    id: number
    datetime: string
    status_id: number
    status: IncidentStatus // ステータスが複合型になりました
    assignee: string
    vender: number
    message_id: string
    created_at: string
    updated_at: string
}

interface APIResponse {
    data: Incident[]
    meta: {
        total: number
        page: number
        limit: number
        pages: number
    }
    status_counts: Array<{
        status: string
        count: number
    }>
    unique_assignees: string[]
}

// 環境変数の型チェック
if (!process.env.DBPILOT_URL) {
    throw new Error('DBPILOT_URL is not defined in environment variables')
}

export async function GET(request: NextRequest) {
    try {
        const searchParams = request.nextUrl.searchParams
        const cookieStore = await cookies()
        const sessionID = cookieStore.get('session_id')?.value
        const page = parseInt(searchParams.get('page') || '1')
        const limit = parseInt(searchParams.get('limit') || '10')
        const status = searchParams.get('status')
        // ステータス名の配列をそのまま送信（バックエンドで対応）
        const statusArray = status ? status.split(',').map((s) => s.trim()) : []

        const from = searchParams.get('from')
        const to = searchParams.get('to')

        const assignee = searchParams.get('assignee')
        const assigneeArray = assignee ? assignee.split(',').map((s) => s.trim()) : []

        const response = await fetch(`${process.env.DBPILOT_URL}/email-all`, {
            method: 'POST',
            headers: {
                Authorization: `Bearer ${sessionID}`,
                'Content-Type': 'application/json' // Content-Typeヘッダーを追加
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
            // エラーレスポンスの詳細な処理
            const errorData = await response.json().catch(() => ({
                message: 'Invalid response from server'
            }))
            return NextResponse.json({ error: errorData.message || 'Request failed' }, { status: response.status })
        }

        // レスポンスのパースを try-catch で囲む
        const data: APIResponse = await response.json()

        // レスポンスの検証
        if (!data || typeof data !== 'object') {
            throw new Error('Invalid response format')
        }

        return NextResponse.json(data)
    } catch (error) {
        // エラーハンドリングの改善
        console.error('Error in getIncidentAll:', error)
        return NextResponse.json({ error: 'Internal server error' }, { status: 500 })
    }
}
