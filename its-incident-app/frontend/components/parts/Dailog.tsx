'use client'

import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from '../ui/accordion'
import { Button } from '../ui/button'
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '../ui/dialog'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { UserContext } from '@/lib/UserContext'
import { Separator } from '../ui/separator'
import { Textarea } from '../ui/textarea'
import { Calendar as CalendarIcon, MailIcon, AlertTriangle, ChevronDown, ChevronUp } from 'lucide-react'
import { format, isWithinInterval, endOfDay, fromUnixTime } from 'date-fns'
import { type Incident, type IncidentsApiResponse } from '@/typs/incident'
import { useContext, useState, useRef, useEffect } from 'react'
import { useFetch } from '@/hooks/useFetch'
import Loading from '../template/Loading'
import { useToast } from '@/hooks/use-toast'
import ParamCard from '../template/ParamCard'
import WorkLog from './WorkLog'
import { StatusUpdate } from './statusUpdate'

interface DialogProps {
    isOpen: boolean
    onClose: () => void
    incident: Incident | null
}

interface ResponseResponse {
    message: string
}

const DialogWindow = ({ isOpen, onClose, incident }: DialogProps) => {
    const { toast } = useToast()

    const tableContainerRef = useRef<HTMLDivElement>(null)
    const userData = useContext(UserContext)
    const [newResponse, setNewResponse] = useState('')
    const [isWorkflowLogExpanded, setIsWorkflowLogExpanded] = useState(false)
    const { data, isLoading } = useFetch<Incident>(incident?.ID && isOpen ? `/api/getIncident/${incident.ID}` : `/api/getIncident/1`, {
        useSWR: true,
        swrOptions: {
            refreshInterval: isOpen ? 1000 : undefined
        }
    })

    const {
        execute,
        data: resposeFetch,
        error
    } = useFetch<ResponseResponse>('/api/response', {
        method: 'POST',
        body: {
            incident_id: data?.ID,
            responder: userData?.name,
            content: newResponse.replace(/\r\n/g, '\n'),
            status: '調査中'
        },
        onSuccess: () => {
            setNewResponse('')
            toast({
                description: resposeFetch ? resposeFetch.message : 'Success'
            })
        },
        onError: () => {
            toast({
                variant: 'destructive',
                description: error.message
            })
        }
    })
    const {
        execute: complete,
        data: completeData,
        error: completeError
    } = useFetch<ResponseResponse>('/api/response', {
        method: 'POST',
        body: {
            incident_id: data?.ID,
            responder: userData?.name,
            content: 'インシデントが解決しました。',
            status: '解決済み'
        },
        onSuccess: (completeData) => {
            setNewResponse('')
            toast({
                description: completeData ? completeData.message : 'Success'
            })
        },
        onError: () => {
            toast({
                variant: 'destructive',
                description: completeError.message
            })
        }
    })
    const {
        execute: escaretion,
        data: escaretionData,
        error: cescaretionError
    } = useFetch<ResponseResponse>('/api/notification', {
        method: 'POST',
        body: {
            incident_id: data?.ID,
            responder: 'システム',
            content: 'エスカレーションされました',
            status: '調査中',
            title: data?.APIData.Subject
        },
        onSuccess: (escaretionData) => {
            setNewResponse('')
            toast({
                description: escaretionData ? escaretionData.message : 'Success'
            })
        },
        onError: () => {
            toast({
                variant: 'destructive',
                description: cescaretionError.message
            })
        }
    })

    const {
        execute: vender,
        data: venderData,
        error: venderError
    } = useFetch<ResponseResponse>('/api/response', {
        method: 'POST',
        body: {
            incident_id: data?.ID, // タイプミスを修正
            responder: userData?.name,
            content: 'ベンダーへ問い合わせ済み',
            status: '調査中',
            vender: 1
        },
        onSuccess: () => {
            setNewResponse('')
            toast({
                description: venderData ? venderData.message : 'Success'
            })
        },
        onError: () => {
            // エラーハンドリングが必要な場合はここに実装
            toast({
                variant: 'destructive',
                description: venderError.message
            })
        }
    })

    useEffect(() => {
        if (tableContainerRef.current) {
            tableContainerRef.current.scrollTo({
                top: tableContainerRef.current.scrollHeight,
                behavior: 'smooth'
            })
        }
    }, [data?.Responses])

    const handleUpdateResponse = {
        complete: () => complete(),
        escaretion: () => escaretion(),
        vender: () => vender(),
        execute: () => {
            if (newResponse === '') return false
            execute()
        }
    }

    const handlWorkLogOpen = () => {
        setIsWorkflowLogExpanded(!isWorkflowLogExpanded)
    }

    if (!data) return null
    if (isLoading) return <Loading />

    return (
        <div>
            <Dialog open={isOpen} onOpenChange={onClose}>
                <DialogContent className="max-w-[80vw] w-full p-0 h-[95vh] flex flex-col dark:bg-black border-b border-white">
                    <DialogHeader className="p-[20px] flex flex-col bg-black dark dark:border-b-2 dark: border-b border-white">
                        <DialogTitle className="text-white">
                            【ID:{incident?.ID}】&nbsp;&nbsp;{incident?.APIData.Subject}
                        </DialogTitle>
                        <DialogDescription className="flex gap-5 pt-4">
                            <ParamCard title="ステータス" icon={true} status={data?.Status} content={''} />
                            <Separator orientation="vertical" />
                            <ParamCard title="判定" badge={true} status={data?.APIData.Judgment} content={''} />
                            <Separator orientation="vertical" />
                            <ParamCard title="発生日時" content={format(data.Datetime, 'yyyy-MM-dd HH:mm')} />
                            <Separator orientation="vertical" />
                            <ParamCard title="差出人" content={`${data?.APIData.From.replace(/<[^>]*>/g, '')}`} />
                            <Separator orientation="vertical" />
                            <ParamCard title="担当者" content={data?.Assignee || '-'} />
                        </DialogDescription>
                    </DialogHeader>
                    {/* <div className="grid grid-cols-2 h-full"> */}
                    <div className="flex overflow-y-auto h-full">
                        <div className="p-6 bg-gray-100 flex flex-col overflow-hidden dark:bg-black w-1/2">
                            <div className="bg-white rounded-lg shadow-md overflow-hidden flex flex-col h-full dark:bg-black dark:border dark:border-white">
                                <div className="flex-grow overflow-y-auto h-full">
                                    <Accordion type="single" collapsible className="w-full" defaultValue="item-0">
                                        <AccordionItem value="item-0" key={data.ID}>
                                            <AccordionTrigger className="px-4 py-2 hover:bg-gray-50 dark:hover:bg-gray-700">
                                                <div className="flex items-center space-x-2 text-left w-full">
                                                    <MailIcon className="h-5 w-5 text-gray-500 flex-shrink-0 dark:text-white" />
                                                    <div className="flex-grow min-w-0">
                                                        <div className="font-semibold truncate break-words whitespace-pre-wrap">{data.APIData.Subject}</div>
                                                        <div className="text-sm text-gray-500 flex justify-between">
                                                            <span className="truncate">{data.APIData.From}</span>
                                                            <span className="flex-shrink-0 ml-2"> {format(fromUnixTime(data.APIData.CreatedAt), 'yyyy-MM-dd HH:mm')}</span>
                                                        </div>
                                                    </div>
                                                </div>
                                            </AccordionTrigger>
                                            <AccordionContent>
                                                <div className="p-4 space-y-2">
                                                    <div>
                                                        <span className="font-semibold">From:</span> {data.APIData.From}
                                                    </div>
                                                    {/* <div>
                                                        <span className="font-semibold">To:</span> {data.APIData.From}
                                                    </div> */}
                                                    <div>
                                                        <span className="font-semibold">Date:</span> {format(fromUnixTime(data.APIData.CreatedAt), 'yyyy-MM-dd HH:mm')}
                                                    </div>
                                                    <Separator className="my-4" />
                                                    <div className="whitespace-pre-wrap">{data.APIData.Body}</div>
                                                </div>
                                            </AccordionContent>
                                        </AccordionItem>
                                        {data.Relations && data.Relations.length > 0 && (
                                            <div>
                                                <div className="px-4 py-2 font-semibold text-gray-700 bg-gray-100">関連メール</div>
                                                {data.Relations.map((relatedIncident, index) => (
                                                    <AccordionItem value={`item-${index + 1}`} key={relatedIncident.RelatedIncident.ID}>
                                                        <AccordionTrigger className="px-4 py-2 hover:bg-gray-50">
                                                            <div className="flex items-center space-x-2 text-left w-full">
                                                                <MailIcon className="h-5 w-5 text-gray-500 flex-shrink-0 dark:text-white" />
                                                                <div className="flex-grow min-w-0">
                                                                    <div className="font-semibold truncate">{relatedIncident.RelatedIncident.Subject}</div>
                                                                    <div className="text-sm text-gray-500 flex justify-between">
                                                                        <span className="truncate">{relatedIncident.RelatedIncident.FromEmail}</span>
                                                                        <span className="flex-shrink-0 ml-2">{format(relatedIncident.RelatedIncident.Datetime, 'yyyy-MM-dd HH:mm')}</span>
                                                                    </div>
                                                                </div>
                                                            </div>
                                                        </AccordionTrigger>
                                                        <AccordionContent>
                                                            <div className="p-4 space-y-2">
                                                                <div>
                                                                    <span className="font-semibold">From:</span> {relatedIncident.RelatedIncident.FromEmail}
                                                                </div>
                                                                <div>
                                                                    <span className="font-semibold">To:</span> {relatedIncident.RelatedIncident.ToEmail}
                                                                </div>
                                                                <div>
                                                                    <span className="font-semibold">Date:</span>
                                                                    {format(relatedIncident.RelatedIncident.Datetime, 'yyyy-MM-dd HH:mm')}
                                                                </div>
                                                                <Separator className="my-4" />
                                                                <div className="whitespace-pre-wrap">{relatedIncident.RelatedIncident.Content}</div>
                                                            </div>
                                                        </AccordionContent>
                                                    </AccordionItem>
                                                ))}
                                            </div>
                                        )}
                                    </Accordion>
                                </div>
                            </div>
                        </div>
                        <div className="p-6 overflow-y-auto h-full w-1/2">
                            <div className="space-y-6 flex flex-col h-full">
                                <StatusUpdate
                                    initialStatus={data.Status}
                                    initialVenderStatus={data.Vender}
                                    onStatusUpdate={handleUpdateResponse.complete}
                                    onVendorContactUpdate={handleUpdateResponse.vender}
                                />
                                <Separator className="dark:bg-white" />
                                <div className="flex-grow">
                                    <div className="flex justify-between items-center mb-2">
                                        <h4 className="font-semibold">対応履歴</h4>
                                        <Button onClick={handleUpdateResponse.escaretion} variant="destructive" className="flex items-center">
                                            <AlertTriangle className="mr-2 h-4 w-4" />
                                            エスカレーション
                                        </Button>
                                    </div>
                                    <div className="max-h-[350px] overflow-y-auto min-h-[100px]" ref={tableContainerRef}>
                                        <Table>
                                            <TableHeader className="sticky top-0 bg-white z-10 dark:bg-black dark:text-white">
                                                <TableRow>
                                                    <TableHead className="w-[120px]">日付</TableHead>
                                                    <TableHead>対応内容</TableHead>
                                                    <TableHead className="w-[120px]">名前</TableHead>
                                                </TableRow>
                                            </TableHeader>
                                            <TableBody>
                                                {data.Responses.map((response) => (
                                                    <TableRow key={response.ID}>
                                                        <TableCell>{format(response.Datetime, 'yyyy-MM-dd HH:mm')}</TableCell>
                                                        <TableCell className="whitespace-pre-wrap">{response.Content}</TableCell>
                                                        <TableCell>{response.Responder}</TableCell>
                                                    </TableRow>
                                                ))}
                                            </TableBody>
                                        </Table>
                                    </div>
                                </div>
                                <div className="h-[170px]">
                                    <h4 className="font-semibold mb-2">新規対応記録</h4>
                                    <div className="flex flex-col gap-2">
                                        <Textarea placeholder="対応内容を入力してください" value={newResponse} onChange={(e) => setNewResponse(e.target.value)} className="dark:border-white" />
                                        <Button onClick={handleUpdateResponse.execute}>記録を追加</Button>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                    <WorkLog isWorkflowLogExpanded={isWorkflowLogExpanded} onClick={handlWorkLogOpen} data={data} />
                </DialogContent>
            </Dialog>
        </div>
    )
}
export default DialogWindow
