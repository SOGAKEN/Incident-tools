'use client'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form'
import DateTimePicker from '@/components/parts/DataTimePicker'
import { Button } from '@/components/ui/button'
import { Textarea } from '@/components/ui/textarea'
import { Input } from '../ui/input'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'

import * as z from 'zod'
import { useEffect, useState } from 'react'
import Loading from './Loading'
import { useFetch } from '@/hooks/useFetch'
import { toast } from '@/hooks/use-toast'

interface ResponseResponse {
    message: string
}

const formSchema = z
    .object({
        title: z.string().min(2, {
            message: 'タイトルは2文字以上で入力してください。'
        }),
        startDateTime: z.date({
            required_error: '開始日時を選択してください。'
        }),
        endDateTime: z.date({
            required_error: '終了日時を選択してください。'
        }),
        worker: z.string().min(1, {
            message: '作業者を入力してください。'
        }),
        verifier: z.string().min(1, {
            message: '確認者を入力してください。'
        }),
        target: z.string().min(1, {
            message: '対象を入力してください。'
        }),
        client: z.string().min(1, {
            message: 'クライアントを入力してください。'
        }),
        content: z.string().min(10, {
            message: '内容は10文字以上で入力してください。'
        })
    })
    .refine((data) => data.endDateTime > data.startDateTime, {
        message: '終了日時は開始日時より後である必要があります。',
        path: ['endDateTime']
    })

const AddTaskForm = () => {
    const [isLoading, setIsLoading] = useState(true)

    useEffect(() => {
        setIsLoading(false)
    }, [])

    const form = useForm<z.infer<typeof formSchema>>({
        resolver: zodResolver(formSchema),
        defaultValues: {
            title: '',
            startDateTime: new Date(),
            endDateTime: new Date(new Date().getTime() + 60 * 60 * 1000), // 1時間後
            worker: '',
            verifier: '',
            target: '',
            client: '',
            content: ''
        }
    })

    const { execute, data, error } = useFetch<ResponseResponse>('/api/work', {
        method: 'POST',
        onSuccess: (data) => {
            toast({
                description: data ? data.message : 'Success'
            })
        },
        onError: (error) => {
            toast({
                variant: 'destructive',
                description: error.message
            })
        }
    })

    async function onSubmit(values: z.infer<typeof formSchema>) {
        console.log(values)
        await execute({
            body: values
        })
    }

    if (isLoading) return <Loading />

    return (
        <div className="container mx-auto py-10 px-4 md:px-6 lg:px-8">
            <Card className="max-w-2xl mx-auto">
                <CardHeader>
                    <CardTitle className="text-2xl font-bold">新規作業連絡</CardTitle>
                    <CardDescription>作業の詳細を入力してください。</CardDescription>
                </CardHeader>
                <CardContent>
                    <Form {...form}>
                        <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6">
                            <FormField
                                control={form.control}
                                name="title"
                                render={({ field }) => (
                                    <FormItem>
                                        <FormLabel>タイトル</FormLabel>
                                        <FormControl>
                                            <Input placeholder="作業タイトルを入力" {...field} />
                                        </FormControl>
                                        <FormMessage />
                                    </FormItem>
                                )}
                            />
                            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                                <FormField control={form.control} name="startDateTime" render={({ field }) => <DateTimePicker field={field} label="開始日時" />} />
                                <FormField control={form.control} name="endDateTime" render={({ field }) => <DateTimePicker field={field} label="終了日時" />} />
                            </div>
                            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                                <FormField
                                    control={form.control}
                                    name="worker"
                                    render={({ field }) => (
                                        <FormItem>
                                            <FormLabel>作業者</FormLabel>
                                            <FormControl>
                                                <Input placeholder="作業者名を入力" {...field} />
                                            </FormControl>
                                            <FormMessage />
                                        </FormItem>
                                    )}
                                />
                                <FormField
                                    control={form.control}
                                    name="verifier"
                                    render={({ field }) => (
                                        <FormItem>
                                            <FormLabel>確認者</FormLabel>
                                            <FormControl>
                                                <Input placeholder="確認者名を入力" {...field} />
                                            </FormControl>
                                            <FormMessage />
                                        </FormItem>
                                    )}
                                />
                            </div>
                            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                                <FormField
                                    control={form.control}
                                    name="target"
                                    render={({ field }) => (
                                        <FormItem>
                                            <FormLabel>対象ホスト</FormLabel>
                                            <FormControl>
                                                <Input placeholder="対象を入力" {...field} />
                                            </FormControl>
                                            <FormMessage />
                                        </FormItem>
                                    )}
                                />
                                <FormField
                                    control={form.control}
                                    name="client"
                                    render={({ field }) => (
                                        <FormItem>
                                            <FormLabel>クライアント</FormLabel>
                                            <FormControl>
                                                <Input placeholder="クライアント名を入力" {...field} />
                                            </FormControl>
                                            <FormMessage />
                                        </FormItem>
                                    )}
                                />
                            </div>
                            <FormField
                                control={form.control}
                                name="content"
                                render={({ field }) => (
                                    <FormItem>
                                        <FormLabel>内容</FormLabel>
                                        <FormControl>
                                            <Textarea placeholder="作業内容の詳細を入力してください" className="resize-none" {...field} />
                                        </FormControl>
                                        <FormMessage />
                                    </FormItem>
                                )}
                            />
                            <Button type="submit" className="w-full">
                                作業連絡を送信
                            </Button>
                        </form>
                    </Form>
                </CardContent>
            </Card>
        </div>
    )
}

export default AddTaskForm
