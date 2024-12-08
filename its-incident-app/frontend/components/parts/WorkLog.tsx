import { Incident } from '@/typs/incident'
import { Calendar as CalendarIcon, MailIcon, AlertTriangle, ChevronDown, ChevronUp, Bot, Brain, ChevronsRight } from 'lucide-react'
import { useEffect, useState } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '../ui/card'

interface WorkflowLog {
    [key: string]: string | null
}

interface ParsedAnswer {
    [key: string]: string
}

type WorkLog = {
    isWorkflowLogExpanded: boolean
    onClick: () => void
    data: Incident
}

type ParseData = {
    answer: string
    judgment: string
    evidence: string
    attention?: string
    time?: string
    pastResponseHitsoty: string
}

const WorkLog = ({ isWorkflowLogExpanded, onClick, data }: WorkLog) => {
    const [parsedAnswers, setParsedAnswers] = useState<ParsedAnswer[]>([])

    useEffect(() => {
        if (!data) return
        if (!data.APIData.WorkflowLogs) return

        const workflowLogsArray: WorkflowLog[] = JSON.parse(data.APIData.WorkflowLogs)
        // nullや空文字列のエントリーをフィルタリング
        const result = workflowLogsArray.reduce((acc: ParsedAnswer[], item) => {
            const key = Object.keys(item)[0]
            const value = item[key]
            // キーが存在し、値がnullでなく、空文字列でもない場合のみ追加
            if (key && value && value.trim() !== '') {
                acc.push({ [key]: value })
            }
            return acc
        }, [])

        setParsedAnswers(result)
    }, [data])

    return (
        <div className={`fixed left-0 bottom-0 w-[49.8%] bg-white shadow-lg transition-all duration-300 ease-in-out ${isWorkflowLogExpanded ? 'h-[calc(100%-10%)]' : 'h-20'} pr-1 dark:bg-background`}>
            <button
                className="w-full h-20 flex items-center justify-between px-4 text-left font-semibold bg-yellow-100 hover:bg-yellow-100 focus:outline-none dark:text-black"
                style={{
                    backgroundSize: '40px 40px'
                }}
                onClick={onClick}
            >
                <div className="text-m font-bold dark:text-black flex gap-3 items-center">
                    <div className="flex items-center gap-2">
                        <Brain className="inline-block w-[35px]" />
                        <span className="text-sm">判別結果：{data.APIData.Final}</span>
                    </div>
                    <ChevronsRight className="inline-block w-[30px] " />
                    <div className="w-[150px] flex items-center gap-2">
                        <Bot className="inline-block" />
                        <span className="text-sm">{data.APIData.Judgment}</span>
                    </div>
                </div>
                {isWorkflowLogExpanded ? <ChevronDown className="h-6 w-6" /> : <ChevronUp className="h-6 w-6" />}
            </button>
            {data.APIData.WorkflowLogs && (
                <div className={`p-4 overflow-y-auto h-[calc(100%-4rem)] ${isWorkflowLogExpanded ? 'block' : 'hidden'}`}>
                    {parsedAnswers.map((logData, index) => (
                        <Card className="col-span-1 md:col-span-1 mb-1" key={index}>
                            {Object.entries(logData).map(([key, value]) => {
                                try {
                                    // ```json と ``` を取り除いて JSON 文字列を抽出
                                    const jsonString = value.replace(/```json|```/g, '').trim()
                                    // JSON をパース
                                    const parsedData: ParseData = JSON.parse(jsonString)

                                    // パースしたデータを表示
                                    return (
                                        <div key={index}>
                                            <CardHeader>
                                                <CardTitle className="text-base">
                                                    判断 {index + 1}：{parsedData.answer}
                                                </CardTitle>
                                            </CardHeader>
                                            <CardContent>
                                                {/* <div key={key} className="p-2"> */}
                                                <div className="p-1">
                                                    <div className="text-sm font-bold pb-2">根拠：{parsedData.evidence}</div>
                                                    <div className="text-sm font-bold pb-2">判断：{parsedData.judgment}</div>

                                                    {parsedData.time ? <div className="text-sm font-bold pb-2">障害検知時刻：{parsedData.judgment}</div> : ''}
                                                    {parsedData.pastResponseHitsoty ? <div className="text-sm font-bold pb-2">過去対応：{parsedData.time}</div> : ''}
                                                    {parsedData.attention ? <div className="text-sm font-bold pb-2 text-red-500">※ {parsedData.attention}</div> : ''}
                                                </div>
                                            </CardContent>
                                        </div>
                                    )
                                } catch (error) {
                                    // パースに失敗した場合は元の文字列を表示
                                    return (
                                        <div key={index}>
                                            <CardHeader>
                                                <CardTitle>回答 {index + 1}</CardTitle>
                                            </CardHeader>

                                            <CardContent>
                                                <div key={key}>{value.replace(/```/g, '').trim()}</div>
                                            </CardContent>
                                        </div>
                                    )
                                }
                            })}
                        </Card>
                    ))}
                    <div className="text-white dark:text-black">{data.MessageID}</div>
                </div>
            )}
        </div>
    )
}

export default WorkLog
