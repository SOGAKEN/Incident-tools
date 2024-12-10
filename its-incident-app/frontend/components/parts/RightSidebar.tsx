'use client'

import { Button } from '@/components/ui/button'
import { Sheet, SheetContent, SheetDescription, SheetHeader, SheetTitle, SheetTrigger } from '@/components/ui/sheet'
import { useToast } from '@/hooks/use-toast'
import { useFetch } from '@/hooks/useFetch'
import { UserContext } from '@/lib/UserContext'
import { useContext, useState } from 'react'
import { Terminal } from 'lucide-react'

import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'

type Balk = {
    mesasge_id: string[]
    uniqe_host: string[]
}

type BalkResponse = {
    message: string
}

export function RightSidebar({ mesasge_id, uniqe_host }: Balk) {
    const { toast } = useToast()
    const userData = useContext(UserContext)
    const [open, setOpen] = useState(false)
    // Fetchの状態と手動実行関数を初期化
    const { execute, isLoading, data, error } = useFetch<BalkResponse>('/api/balkChange', {
        method: 'POST',
        onSuccess: (data) => {
            toast({
                description: data?.message || 'Success',
                variant: 'green'
            })
            setOpen(false)
        },
        onError: (error) => {
            toast({
                variant: 'destructive',
                description: error.message
            })
        }
    })

    const balkStatusUpdateHandler = () => {
        execute({
            body: {
                message_ids: mesasge_id,
                status_id: 2,
                response: '一括でステータスを「解決済み」に更新しました',
                responder: userData?.name
            }
        })
    }

    console.log('seet', uniqe_host)
    return (
        <Sheet open={open} onOpenChange={setOpen}>
            <SheetTrigger asChild>
                <Button variant="outline">一括更新</Button>
            </SheetTrigger>
            <SheetContent side="right">
                <SheetHeader>
                    <SheetTitle>ステータス一括更新</SheetTitle>
                    <SheetDescription>一括で解決済みにします。</SheetDescription>
                </SheetHeader>
                <div>
                    {uniqe_host.map((host, index) => {
                        return (
                            <Alert key={index} className="my-1" variant="blue">
                                <Terminal className="h-4 w-4" />
                                <AlertTitle>Check Host!</AlertTitle>
                                <AlertDescription>{host}</AlertDescription>
                            </Alert>
                        )
                    })}
                </div>
                <div className="my-4">
                    選択<span className="text-2xl font-bold mx-2">{mesasge_id.length}</span>件
                </div>
                <Button onClick={balkStatusUpdateHandler}>更新</Button>
            </SheetContent>
        </Sheet>
    )
}
