'use client'

import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import * as z from 'zod'
import { Eye, EyeOff } from 'lucide-react'

import { Button } from '@/components/ui/button'
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage, FormDescription } from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { useFetch } from '@/hooks/useFetch'
import { redirect, useRouter } from 'next/navigation'

type Email = {
    email: string
}

type Value = {
    email: string
    name: string
    password: string
}

const passwordRegex = /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[!@#$%^&*])[A-Za-z\d!@#$%^&*]{10,}$/

const formSchema = z
    .object({
        email: z.string().email('有効なメールアドレスを入力してください'),
        name: z.string().min(1, '氏名は必須です'),
        password: z.string().min(10, 'パスワードは10文字以上である必要があります').regex(passwordRegex, 'パスワードは大文字、小文字、数字、記号(!@#$%^&*)を含める必要があります'),
        confirmPassword: z.string()
    })
    .refine((data) => data.password === data.confirmPassword, {
        message: 'パスワードが一致しません',
        path: ['confirmPassword']
    })

export default function SettingProfilePage({ email }: Email) {
    const router = useRouter()

    const [showPassword, setShowPassword] = useState(false)
    const [showConfirmPassword, setShowConfirmPassword] = useState(false)

    const form = useForm<z.infer<typeof formSchema>>({
        resolver: zodResolver(formSchema),
        defaultValues: {
            email: email,
            name: '',
            password: '',
            confirmPassword: ''
        }
    })
    const { execute, error, data } = useFetch<Value>('/api/settingProfile', {
        method: 'POST',
        onSuccess: () => {
            router.push('/login')
        },
        onError: (error) => {
            console.error('エラーが発生しました。', error.message)
        }
    })

    async function onSubmit(values: z.infer<typeof formSchema>) {
        console.log(values)
        // ここでバックエンドAPIにデータを送信する処理を実装します
        await execute({
            body: {
                email: values.email,
                name: values.name,
                password: values.password
            }
        })
    }

    return (
        <Form {...form}>
            <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6">
                <FormField
                    control={form.control}
                    name="email"
                    render={({ field }) => (
                        <FormItem>
                            <FormLabel>メールアドレス</FormLabel>
                            <FormControl>
                                <Input {...field} autoComplete="off" disabled />
                            </FormControl>
                            <FormMessage />
                        </FormItem>
                    )}
                />
                <FormField
                    control={form.control}
                    name="name"
                    render={({ field }) => (
                        <FormItem>
                            <FormLabel>氏名</FormLabel>
                            <FormControl>
                                <Input placeholder="山田 太郎" {...field} autoComplete="off" />
                            </FormControl>
                            <FormMessage />
                        </FormItem>
                    )}
                />
                <FormField
                    control={form.control}
                    name="password"
                    render={({ field }) => (
                        <FormItem>
                            <FormLabel>パスワード</FormLabel>
                            <FormControl>
                                <div className="relative">
                                    <Input type={showPassword ? 'text' : 'password'} placeholder="********" {...field} autoComplete="new-password" />
                                    <Button type="button" variant="ghost" size="icon" className="absolute right-2 top-1/2 -translate-y-1/2" onClick={() => setShowPassword(!showPassword)}>
                                        {showPassword ? <Eye className="h-4 w-4" /> : <EyeOff className="h-4 w-4" />}
                                    </Button>
                                </div>
                            </FormControl>
                            <FormDescription>パスワードは10文字以上で、大文字、小文字、数字、記号(!@#$%^&*)を含める必要があります。</FormDescription>
                            <FormMessage />
                        </FormItem>
                    )}
                />
                <FormField
                    control={form.control}
                    name="confirmPassword"
                    render={({ field }) => (
                        <FormItem>
                            <FormLabel>パスワード（確認）</FormLabel>
                            <FormControl>
                                <div className="relative">
                                    <Input type={showConfirmPassword ? 'text' : 'password'} placeholder="********" {...field} autoComplete="new-password" />
                                    <Button
                                        type="button"
                                        variant="ghost"
                                        size="icon"
                                        className="absolute right-2 top-1/2 -translate-y-1/2"
                                        onClick={() => setShowConfirmPassword(!showConfirmPassword)}
                                    >
                                        {showConfirmPassword ? <Eye className="h-4 w-4" /> : <EyeOff className="h-4 w-4" />}
                                    </Button>
                                </div>
                            </FormControl>
                            <FormMessage />
                        </FormItem>
                    )}
                />
                <Button type="submit" className="w-full">
                    登録
                </Button>
            </form>
        </Form>
    )
}
