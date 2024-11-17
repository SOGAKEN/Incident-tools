'use client'

import { useState } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Label } from '@/components/ui/label'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { useRouter } from 'next/navigation'
import { useFetch } from '@/hooks/useFetch'
import Loading from './Loading'
import { EyeIcon, EyeOffIcon } from 'lucide-react'

interface LoginResponse {
    token: string
    user: {
        id: string
        name: string
        email: string
    }
    status?: number
}

export default function LoginForm() {
    const router = useRouter()
    const [email, setEmail] = useState('')
    const [password, setPassword] = useState('')
    const [loading, setLoading] = useState(false)
    const [faild, setFaild] = useState(false)
    const [showPassword, setShowPassword] = useState(false)

    const { execute, data, isLoading, error } = useFetch<LoginResponse>('/api/login', {
        method: 'POST',
        onSuccess: (data) => {
            if (data.status != 401) {
                router.push('/dashboard')
            } else {
                setFaild(true)
                setLoading(false)
            }
        },
        onError: () => {
            setLoading(false)
            console.error('エラーが発生しました。')
        }
    })

    if (isLoading || loading) return <Loading />
    if (error) return <div>エラーが発生しました。</div>

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault()
        setLoading(true)
        await execute({
            body: {
                email,
                password
            }
        })
    }

    return (
        <div className="flex items-center justify-center min-h-screen">
            <Card className="w-full max-w-md">
                <CardHeader>
                    <CardTitle>ログイン</CardTitle>
                    <CardDescription>{!faild ? 'アカウントにログインしてください。' : <span className="text-red-600 text-bold">ログイン失敗</span>}</CardDescription>
                </CardHeader>
                <CardContent>
                    <form onSubmit={handleSubmit} className="space-y-4">
                        <div className="space-y-2">
                            <Label htmlFor="email">メールアドレス</Label>
                            <Input id="email" type="email" placeholder="メールアドレスを入力" value={email} onChange={(e) => setEmail(e.target.value)} required />
                        </div>
                        <div className="space-y-2">
                            <Label htmlFor="password">パスワード</Label>

                            <div className="relative">
                                <Input id="password" type={showPassword ? 'text' : 'password'} placeholder="パスワードを入力" value={password} onChange={(e) => setPassword(e.target.value)} required />
                                <Button
                                    type="button"
                                    variant="ghost"
                                    size="icon"
                                    className="absolute right-2 top-1/2 -translate-y-1/2"
                                    onClick={() => setShowPassword(!showPassword)}
                                    aria-label={showPassword ? 'パスワードを隠す' : 'パスワードを表示'}
                                >
                                    {showPassword ? <EyeIcon className="h-4 w-4" /> : <EyeOffIcon className="h-4 w-4" />}
                                </Button>
                            </div>
                        </div>
                        <Button type="submit" className="w-full">
                            ログイン
                        </Button>
                    </form>
                </CardContent>
            </Card>
        </div>
    )
}
