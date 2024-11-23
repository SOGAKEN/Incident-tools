import { NextResponse } from 'next/server'
import type { NextRequest } from 'next/server'

async function verifyToken(token: string): Promise<{ valid: boolean; email?: string }> {
    try {
        const response = await fetch(`${process.env.NEXT_PUBLIC_AUTH_SERVICE_URL}/verify-token?token=${token}`, {
            method: 'GET',
            headers: {
                'Content-Type': 'application/json',
                Authorization: `Bearer ${process.env.AUTH_SERVICE_TOKEN}`
            }
        })

        if (!response.ok) {
            return { valid: false }
        }

        const data = await response.json()
        return { valid: true, email: data.email }
    } catch (error) {
        console.error('Token verification error:', error)
        return { valid: false }
    }
}

export async function middleware(request: NextRequest) {
    const { pathname, searchParams } = new URL(request.url)
    const token = searchParams.get('token')
    const redirectTo = searchParams.get('redirectTo')

    // トークンがある場合（メールリンクからの遷移）
    if (token) {
        // トークンを検証
        const { valid, email } = await verifyToken(token)
        if (valid && email) {
            // リダイレクト先の決定
            const targetUrl = new URL(redirectTo || '/account', request.url)
            // emailパラメータを追加
            targetUrl.searchParams.set('email', email)
            const response = NextResponse.redirect(targetUrl)

            // セッションクッキーの設定
            response.cookies.set('session_id', email, {
                httpOnly: true,
                secure: process.env.NODE_ENV === 'production',
                sameSite: 'lax',
                maxAge: 60 * 60 * 24 // 24時間
            })

            return response
        } else {
            // 無効なトークンの場合
            return NextResponse.redirect(new URL('/login?error=invalid_token', request.url))
        }
    }

    // トークンがない場合は通常の認証チェック
    const sessionId = request.cookies.get('session_id')?.value

    // セッションIDがない場合はログインページにリダイレクト
    if (!sessionId) {
        const loginUrl = new URL('/login', request.url)
        // 現在のURLを保存
        loginUrl.searchParams.set('redirectTo', pathname + searchParams.toString())
        return NextResponse.redirect(loginUrl)
    }

    // URLパラメータを保持したまま次の処理へ
    const response = NextResponse.next()

    return response
}

// ミドルウェアを適用するルートの設定
export const config = {
    matcher: [
        // トークン検証が必要なパス
        '/account',
        // 保護されたルート
        '/dashboard/:path*',
        '/profile/:path*',
        // トークンパラメータを含む可能性のあるパス
        '/((?!api|_next/static|_next/image|favicon.ico).*)'
    ]
}
