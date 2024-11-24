'use client'

import { useContext, useEffect, useState } from 'react'
import { UserContext } from '@/lib/UserContext'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { EyeIcon, EyeOffIcon } from 'lucide-react'
import { useFetch } from '@/hooks/useFetch'
import { redirect, useRouter } from 'next/navigation'
import { Alert, AlertDescription } from '@/components/ui/alert'
import Loading from './Loading'

type Profile = {
    name: string
    password: string
}

export function ProfileRegistrationComponent() {
    const router = useRouter()
    const [name, setName] = useState('')
    const [password, setPassword] = useState('')
    const [currentPassword, setCurrentPassword] = useState('')
    const [showCurrentPassword, setShowCurrentPassword] = useState(false)
    const [showPassword, setShowPassword] = useState(false)
    const [passwordError, setPasswordError] = useState('')
    const [isLoading, setIsLoading] = useState(true)
    const userData = useContext(UserContext)

    const { execute, error } = useFetch<Profile>('/api/createProfile', {
        method: 'POST',
        body: {
            name: name,
            password: password,
            current_password: currentPassword
        },
        onSuccess: () => {
            router.push('/login')
        },
        onError: () => {
            console.error('エラーが発生しました。')
        }
    })

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault()
        execute()
    }

    const validatePassword = (value: string) => {
        const regex = /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[!@#$%^&*])[A-Za-z\d!@#$%^&*]{10,}$/
        if (!regex.test(value)) {
            setPasswordError('パスワードは10文字以上で、大文字、小文字、数字、記号(!@#$%^&*)を含める必要があります。')
        } else {
            setPasswordError('')
        }
    }

    const handlePasswordChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        const newPassword = e.target.value
        setPassword(newPassword)
        validatePassword(newPassword)
    }
    const handlecurrentPasswordChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        const newPassword = e.target.value
        setCurrentPassword(newPassword)
    }

    if (isLoading) return <Loading />

    return (
        <div className="flex items-center justify-center min-h-[500px]">
            <Card className="w-full max-w-md mx-auto">
                <CardHeader>
                    <CardTitle className="text-2xl font-bold text-center">プロフィール登録</CardTitle>
                </CardHeader>
                <CardContent>
                    <form onSubmit={handleSubmit} className="space-y-6">
                        <div className="space-y-2">
                            <Label htmlFor="email">メールアドレス</Label>
                            <Input id="email" type="text" value={userData?.email} disabled />
                        </div>
                        <div className="space-y-2">
                            <Label htmlFor="name">名前</Label>
                            <Input id="name" type="text" value={name} onChange={(e) => setName(e.target.value)} required placeholder="山田 太郎" autoComplete="off" />
                        </div>
                        <div className="space-y-2">
                            <Label htmlFor="currentPassword">現在のパスワード</Label>
                            <div className="relative">
                                <Input
                                    id="currentPassword"
                                    type={showCurrentPassword ? 'text' : 'password'}
                                    value={currentPassword}
                                    onChange={handlecurrentPasswordChange}
                                    required
                                    placeholder="現在のパスワードを入力"
                                    aria-invalid={!!passwordError}
                                    aria-describedby="password-error"
                                />
                                <Button
                                    type="button"
                                    variant="ghost"
                                    size="icon"
                                    className="absolute right-2 top-1/2 -translate-y-1/2"
                                    onClick={() => setShowCurrentPassword(!showCurrentPassword)}
                                    aria-label={showCurrentPassword ? 'パスワードを隠す' : 'パスワードを表示'}
                                >
                                    {showCurrentPassword ? <EyeIcon className="h-4 w-4" /> : <EyeOffIcon className="h-4 w-4" />}
                                </Button>
                            </div>
                            {passwordError && (
                                <Alert variant="destructive" className="mt-2">
                                    <AlertDescription id="password-error">{passwordError}</AlertDescription>
                                </Alert>
                            )}
                        </div>
                        <div className="space-y-2">
                            <Label htmlFor="password">新しいパスワード</Label>
                            <div className="relative">
                                <Input
                                    id="password"
                                    type={showPassword ? 'text' : 'password'}
                                    value={password}
                                    onChange={handlePasswordChange}
                                    required
                                    placeholder="新しいパスワードを入力"
                                    aria-invalid={!!passwordError}
                                    aria-describedby="password-error"
                                />
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
                            {passwordError && (
                                <Alert variant="destructive" className="mt-2">
                                    <AlertDescription id="password-error">{passwordError}</AlertDescription>
                                </Alert>
                            )}
                        </div>
                        <Button type="submit" className="w-full" disabled={!!passwordError}>
                            登録
                        </Button>
                    </form>
                </CardContent>
            </Card>
        </div>
    )
}
