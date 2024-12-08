'use client'

import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from '@/components/ui/accordion'
import { Button } from '../ui/button'
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { UserContext } from '@/lib/UserContext'
import { Separator } from '@/components/ui/separator'
import { Textarea } from '@/components/ui/textarea'
import { Calendar as CalendarIcon, MailIcon, AlertTriangle, ChevronDown, ChevronUp } from 'lucide-react'
import { format, isWithinInterval, endOfDay, fromUnixTime } from 'date-fns'
import { type Incident, type IncidentsApiResponse, type Data, type EmailData } from '@/typs/incident'
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

interface APIResponse {
    message: string
}

interface ActionConfig {
    endpoint: string
    body: Record<string, any>
    successMessage?: string
}

const useAPIAction = (defaultConfig: Partial<ActionConfig>) => {
    const { toast } = useToast()

    // Fetchの状態と手動実行関数を初期化
    const { execute, isLoading, data, error } = useFetch<APIResponse>(defaultConfig.endpoint || '', {
        method: 'POST',
        onSuccess: (data) => {
            toast({
                description: data?.message || defaultConfig.successMessage || 'Success',
                variant: 'green'
            })
        },
        onError: (error) => {
            console.log(error)
            toast({
                variant: 'destructive',
                description: error.message
            })
        }
    })

    // 実行関数を返す
    const executeAction = async (config: Partial<ActionConfig> = {}) => {
        const mergedConfig = { ...defaultConfig, ...config }

        if (!mergedConfig.endpoint) {
            throw new Error('Endpoint is required for API action')
        }

        return await execute({ body: mergedConfig.body })
    }

    return {
        execute: executeAction,
        isLoading,
        data,
        error
    }
}

const DialogWindow = ({ isOpen, onClose, incident }: DialogProps) => {
    const tableContainerRef = useRef<HTMLDivElement>(null)
    const userData = useContext(UserContext)
    const [newResponse, setNewResponse] = useState('')
    const [isWorkflowLogExpanded, setIsWorkflowLogExpanded] = useState(false)
    const { data, isLoading } = useFetch<Data>(incident?.MessageID && isOpen ? `/api/getIncident/${incident.MessageID}` : null, {
        useSWR: true,
        swrOptions: {
            refreshInterval: isOpen ? 1000 : undefined
        }
    })
    const responseAction = useAPIAction({ endpoint: '/api/response' })
    const notificationAction = useAPIAction({ endpoint: '/api/notification' })

    const actions = {
        response: () =>
            responseAction.execute({
                body: {
                    incident_id: data?.Incident.ID,
                    responder: userData?.name,
                    content: newResponse.replace(/\r\n/g, '\n'),
                    status: '調査中'
                }
            }),
        complete: () =>
            responseAction.execute({
                body: {
                    incident_id: data?.Incident.ID,
                    responder: userData?.name,
                    content: 'インシデントが解決しました。',
                    status: '解決済み'
                }
            }),
        escalation: () =>
            notificationAction.execute({
                body: {
                    incident_id: data?.Incident.ID,
                    responder: 'システム',
                    content: 'エスカレーションされました',
                    status: '調査中',
                    title: data?.EmailData.subject
                }
            }),
        vender: () =>
            responseAction.execute({
                body: {
                    incident_id: data?.Incident.ID,
                    responder: userData?.name,
                    content: 'ベンダーへ問い合わせ済み',
                    status: '調査中',
                    vender: 1
                }
            })
    }

    useEffect(() => {
        if (tableContainerRef.current) {
            tableContainerRef.current.scrollTo({
                top: tableContainerRef.current.scrollHeight,
                behavior: 'smooth'
            })
        }
    }, [data?.Incident.Responses])

    const handleAction = {
        complete: () => actions.complete(),
        escalation: () => actions.escalation(),
        vender: () => actions.vender(),
        response: () => {
            if (newResponse === '') return false
            actions.response()
            setNewResponse('')
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
                <DialogContent className="max-w-[80vw] w-full p-0 h-[95vh] flex flex-col dark:bg-background border-b border-white">
                    <DialogHeader className="p-[20px] flex flex-col bg-black dark dark:border-b-2 dark: border-b border-white">
                        <DialogTitle className="text-white text-base">
                            【ID:{data.EmailData.ID}】&nbsp;&nbsp;{data.EmailData.subject}
                        </DialogTitle>
                        <DialogDescription className="flex gap-5 pt-4">
                            <ParamCard title="ステータス" icon={true} status={data.Incident.Status.Name} code={data.Incident.Status.Code} content={''} />
                            <Separator orientation="vertical" />
                            <ParamCard title="判定" badge={true} status={data.Incident.APIData.Judgment} code={data.Incident.APIData.Judgment === '静観' ? 2 : 0} content={''} />
                            <Separator orientation="vertical" />
                            <ParamCard title="発生日時" content={format(data.Incident.Datetime, 'yyyy-MM-dd HH:mm')} code={0} />
                            <Separator orientation="vertical" />
                            <ParamCard title="差出人" content={`${data.Incident.APIData.From.replace(/<[^>]*>/g, '')}`} code={0} />
                            <Separator orientation="vertical" />
                            <ParamCard title="担当者" content={data.Incident.Assignee || '-'} code={0} />
                        </DialogDescription>
                    </DialogHeader>
                    {/* <div className="grid grid-cols-2 h-full"> */}
                    <div className="flex overflow-y-auto h-full">
                        {/* 左ペイン: メール表示 */}
                        <EmailDisplay data={data.EmailData} />

                        {/* 右ペイン: アクション */}
                        <ActionPane data={data.Incident} tableContainerRef={tableContainerRef} newResponse={newResponse} setNewResponse={setNewResponse} handleAction={handleAction} />
                    </div>
                    <WorkLog isWorkflowLogExpanded={isWorkflowLogExpanded} onClick={handlWorkLogOpen} data={data.Incident} />
                </DialogContent>
            </Dialog>
        </div>
    )
}

const EmailDisplay: React.FC<{ data: EmailData }> = ({ data }) => {
    return (
        <div className="p-6 bg-gray-100 flex flex-col overflow-hidden  dark:bg-background w-1/2">
            <div className="bg-white rounded-lg shadow-md overflow-hidden flex flex-col h-[72vh] dark:bg-black dark:border dark:border-white">
                <div className="flex-grow overflow-y-auto h-full">
                    <Accordion type="single" collapsible className="w-full" defaultValue="item-0">
                        <AccordionItem value="item-0" key={data.ID}>
                            <AccordionTrigger className="px-4 py-2 hover:bg-opacity-30">
                                <div className="flex items-center space-x-2 text-left w-full">
                                    <MailIcon className="h-5 w-5 text-gray-500 flex-shrink-0 dark:text-white" />
                                    <div className="flex-grow min-w-0">
                                        <div className="font-semibold truncate break-words whitespace-pre-wrap text-sm">{data.subject}</div>
                                        <div className="text-sm text-gray-500 flex justify-between">
                                            <span className="truncate text-sm">{data.from}</span>
                                            <span className="flex-shrink-0 ml-2 text-sm"> {format(data.CreatedAt, 'yyyy-MM-dd HH:mm')}</span>
                                        </div>
                                    </div>
                                </div>
                            </AccordionTrigger>
                            <AccordionContent>
                                <div className="p-4 space-y-2 max-h-[300px] overflow-y-scroll">
                                    <div>
                                        <span className="font-semibold text-sm">From:</span> {data.from}
                                    </div>
                                    <div>
                                        <span className="font-semibold text-sm">Date:</span> {format(data.CreatedAt, 'yyyy-MM-dd HH:mm')}
                                    </div>
                                    <Separator className="my-4" />
                                    <div className="whitespace-pre-wrap text-sm">{data.body}</div>
                                </div>
                            </AccordionContent>
                        </AccordionItem>
                        {data.Incident.Relations && data.Incident.Relations.length > 0 && (
                            <div>
                                <div className="px-4 py-2 font-semibold text-gray-700 bg-gray-100">関連メール</div>
                                {data.Incident.Relations.map((relatedIncident, index) => (
                                    <AccordionItem value={`item-${index + 1}`} key={relatedIncident.ID}>
                                        <AccordionTrigger className="px-4 py-2 hover:bg-opacity-30">
                                            <div className="flex items-center space-x-2 text-left w-full">
                                                <MailIcon className="h-5 w-5 text-gray-500 flex-shrink-0 dark:text-white" />
                                                <div className="flex-grow min-w-0">
                                                    <div className="font-semibold text-xs">{relatedIncident.RelatedIncident?.APIData?.Subject}</div>
                                                    <div className="text-sm text-gray-500 flex justify-between">
                                                        <span className="truncate text-xs">{relatedIncident.RelatedIncident.APIData.From}</span>
                                                        <span className="flex-shrink-0 ml-2 text-xs">{format(relatedIncident.RelatedIncident.CreatedAt, 'yyyy-MM-dd HH:mm')}</span>
                                                    </div>
                                                </div>
                                            </div>
                                        </AccordionTrigger>
                                        <AccordionContent>
                                            <div className="p-4 space-y-2 max-h-[300px] overflow-y-scroll">
                                                <div className="whitespace-pre-wrap text-xs">{relatedIncident.RelatedIncident.APIData.Body}</div>
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
    )
}

const ActionPane: React.FC<{
    data: Incident
    tableContainerRef: React.RefObject<HTMLDivElement>
    newResponse: string
    setNewResponse: (value: string) => void
    handleAction: Record<string, () => void>
}> = ({ data, tableContainerRef, newResponse, setNewResponse, handleAction }) => {
    // アクション部分のコンポーネント実装
    return (
        <div className="p-6 overflow-y-auto h-full w-1/2">
            <div className="space-y-6 flex flex-col h-full">
                <StatusUpdate initialStatus={data.Status.Name} initialVenderStatus={data.Vender} onStatusUpdate={handleAction.complete} onVendorContactUpdate={handleAction.vender} />
                <Separator className="dark:bg-white" />
                <div className="flex-grow">
                    <div className="flex justify-between items-center mb-2">
                        <h4 className="font-semibold">対応履歴</h4>
                        <Button onClick={handleAction.escalation} variant="destructive" className="flex items-center">
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
                        <Button onClick={handleAction.response}>記録を追加</Button>
                    </div>
                </div>
            </div>
        </div>
    )
}
export default DialogWindow
