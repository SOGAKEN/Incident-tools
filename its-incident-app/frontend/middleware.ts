// middleware.ts
import { NextResponse } from 'next/server'
import type { NextRequest } from 'next/server'

async function verifyToken(token: string): Promise<{ valid: boolean; email?: string }> {
    try {
        // エラーハンドリングの強化
        if (!process.env.NEXT_PUBLIC_AUTH_SERVICE_URL) {
            // console.error('AUTH_SERVICE_URL is not defined')
            return { valid: false }
        }

        if (!token) {
            // console.error('Token is empty')
            return { valid: false }
        }

        const url = `${process.env.NEXT_PUBLIC_AUTH_SERVICE_URL}/verify-token?token=${token}`
        // console.log('Verification URL:', url)

        const response = await fetch(url, {
            method: 'GET',
            headers: {
                'Content-Type': 'application/json',
                Authorization: `Bearer ${process.env.AUTH_SERVICE_TOKEN || ''}`
            }
        })

        // レスポンスの詳細なログ
        // console.log('Verification response:', {
        //     status: response.status,
        //     statusText: response.statusText,
        //     headers: Object.fromEntries(response.headers)
        // })

        if (!response.ok) {
            const errorText = await response.text()
            // console.error('Verification failed:', errorText)
            return { valid: false }
        }

        const data = await response.json()
        return { valid: true, email: data.data.email }
    } catch (error) {
        console.error('Token verification error:', error)
        return { valid: false }
    }
}

export async function middleware(request: NextRequest) {
    try {
        // リクエスト情報の詳細なログ
        // console.log('Middleware execution started', {
        //     url: request.url,
        //     method: request.method,
        //     headers: Object.fromEntries(request.headers),
        //     cookies: Object.fromEntries(request.cookies)
        // })

        const { pathname, searchParams } = new URL(request.url)
        const token = searchParams.get('token')
        const redirectTo = searchParams.get('redirectTo')
        const sessionId = request.cookies.get('session_id')?.value

        // 環境変数の確認
        // console.log('Environment variables:', {
        //     NODE_ENV: process.env.NODE_ENV,
        //     AUTH_SERVICE_URL: process.env.NEXT_PUBLIC_AUTH_SERVICE_URL,
        //     IS_CLOUD_RUN: process.env.K_SERVICE ? 'yes' : 'no'
        // })

        if (token) {
            const verificationResult = await verifyToken(token)
            // console.log('Token verification result:', verificationResult)
            //
            if (verificationResult.valid && verificationResult.email) {
                const targetUrl = new URL(redirectTo || '/account', request.url)
                targetUrl.searchParams.set('email', verificationResult.email)

                const response = NextResponse.redirect(targetUrl)

                // Cookie設定のデバッグログ
                // console.log('Setting cookie for domain:', new URL(request.url).hostname)
                //
                // response.cookies.set('session_id', verificationResult.email, {
                //     httpOnly: true,
                //     secure: process.env.NODE_ENV === 'production',
                //     sameSite: 'lax',
                //     maxAge: 60 * 60 * 24,
                //     domain: new URL(request.url).hostname,
                //     path: '/'
                // })

                response.cookies.set('session_id', verificationResult.email, {
                    httpOnly: true,
                    secure: true,
                    sameSite: 'lax',
                    maxAge: 60 * 60 * 24,
                    domain: new URL(request.url).hostname,
                    path: '/'
                })
                return response
            }

            return NextResponse.redirect(new URL('/login?error=invalid_token', request.url))
        }

        if (!sessionId) {
            const loginUrl = new URL('/login', request.url)
            if (pathname !== '/login') {
                loginUrl.searchParams.set('redirectTo', pathname + (searchParams.toString() ? `?${searchParams.toString()}` : ''))
            }
            return NextResponse.redirect(loginUrl)
        }

        return NextResponse.next()
    } catch (error) {
        console.error('Middleware error:', error)
        // エラー時はログインページにリダイレクト
        return NextResponse.redirect(new URL('/login?error=server_error', request.url))
    }
}

export const config = {
    matcher: ['/((?!api|_next/static|_next/image|favicon.ico|login).*)']
}
