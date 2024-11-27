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

const WorkLog = ({ isWorkflowLogExpanded, onClick, data }: WorkLog) => {
    const [parsedAnswers, setParsedAnswers] = useState<ParsedAnswer[]>([])

    useEffect(() => {
        if (!data) return

        const workflowLogsArray: WorkflowLog[] = JSON.parse(data.APIData.WorkflowLogs)
        const result = workflowLogsArray.reduce((acc: ParsedAnswer[], item) => {
            const key = Object.keys(item)[0]
            if (item[key] !== null && key !== undefined && key !== '') {
                acc.push({ [key]: item[key] })
            }
            return acc
        }, [])

        setParsedAnswers(result)
    }, [data])
    return (
        <div className={`fixed left-0 bottom-0 w-[49.8%] bg-white shadow-lg transition-all duration-300 ease-in-out ${isWorkflowLogExpanded ? 'h-full' : 'h-20'} pr-1 dark:bg-black`}>
            <button
                className="w-full h-20 flex items-center justify-between px-4 text-left font-semibold bg-yellow-100 hover:bg-yellow-100 focus:outline-none dark:text-black"
                style={{
                    backgroundSize: '40px 40px'
                }}
                onClick={onClick}
            >
                <div className="text-m font-bold dark:text-black flex gap-3 items-center">
                    <div className="flex items-center gap-2">
                        <Brain className="inline-block" /> 判別結果：{data.APIData.Final}
                    </div>
                    <ChevronsRight className="inline-block w-[30px] " />
                    <div className="w-[150px] flex items-center gap-2">
                        <Bot className="inline-block" />
                        {data.APIData.Judgment}
                    </div>
                </div>
                {isWorkflowLogExpanded ? <ChevronDown className="h-6 w-6" /> : <ChevronUp className="h-6 w-6" />}
            </button>
            {data.APIData.WorkflowLogs && (
                <div className={`p-4 overflow-y-auto h-[calc(100%-4rem)] ${isWorkflowLogExpanded ? 'block' : 'hidden'}`}>
                    {parsedAnswers.map((logData, index) => (
                        <Card className="col-span-1 md:col-span-1 mb-1" key={index}>
                            <CardHeader>
                                <CardTitle>回答 {index + 1}</CardTitle>
                            </CardHeader>
                            <CardContent>
                                {Object.entries(logData).map(([key, value]) => (
                                    <div key={key}>{value}</div>
                                ))}
                            </CardContent>
                        </Card>
                    ))}
                </div>
            )}
        </div>
    )
}

export default WorkLog
