'use client'
import { useFetch } from '@/hooks/useFetch'
import { type Incident, type IncidentsApiResponse } from '@/typs/incident'

import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { AlertCircle, CheckCircle, Clock, Calendar as CalendarIcon, Wrench } from 'lucide-react'
import { format, fromUnixTime } from 'date-fns'
import { useEffect, useState } from 'react'
import CardCounter from '../parts/CardCounter'

import Loading from './Loading'
import PageNation from '../parts/PageNation'
import GetStatusIcon from '../parts/GetStatusIcon'
import { SearchComponent } from '../parts/Search'
import { DateRange } from 'react-day-picker'

interface TopTableProps {
    onIncidentClick: (incident: Incident) => void
}

const TopTable = ({ onIncidentClick }: TopTableProps) => {
    const [checkSatatus, setcheckStatas] = useState<string[]>(['未着手', '調査中', '失敗'])
    const [checkAssignees, setcheckAssignees] = useState<string[]>([])
    const [fromStatus, setFromStatus] = useState<Date | undefined>(undefined)
    const [toStatus, setToStatus] = useState<Date | undefined>(undefined)
    const [queryParam, setQueryParam] = useState<string>('未着手%2C調査中%2C失敗')
    const [queryAssignee, setAssignee] = useState<string>('')
    const [dateParam, setDateParam] = useState<string>('')

    const [page, setPage] = useState(1)
    const [limit, setLimit] = useState(100)
    const { data, isLoading } = useFetch<IncidentsApiResponse>(`/api/getIncidentAll?page=${page}&limit=${limit}&status=${queryParam}${dateParam}&assignee=${queryAssignee}`, {
        useSWR: true,
        swrOptions: {
            refreshInterval: 5000
        }
    })

    const handleSearch = async (selectedStatuses: string[], dateRange: DateRange | undefined, selectUniqueAssignees: string[]): Promise<void> => {
        // ステータスの処理
        const query = selectedStatuses.length > 0 ? selectedStatuses.join('%2C') : ''
        setQueryParam(query)
        setcheckStatas(selectedStatuses)
        setPage(1)

        const assignee = selectUniqueAssignees.length > 0 ? selectUniqueAssignees.join('%2C') : ''
        setAssignee(assignee)
        setcheckAssignees(selectUniqueAssignees)

        // 日付範囲の処理
        if (dateRange?.from && dateRange?.to) {
            setFromStatus(dateRange.from)
            setToStatus(dateRange.to)
            setDateParam(`&from=${format(dateRange.from, 'yyyy-MM-dd 00:00')}&to=${format(dateRange.to, 'yyyy-MM-dd 23:59')}`)
        } else {
            // 日付範囲がない場合はリセット
            setFromStatus(undefined)
            setToStatus(undefined)
            setDateParam('')
        }
    }

    const initialDateRange: DateRange | undefined =
        fromStatus && toStatus
            ? {
                  from: fromStatus,
                  to: toStatus
              }
            : undefined

    if (!data) {
        return null
    }

    const handlers = {
        first: () => setPage(1),
        previous: () => setPage((prev) => Math.max(prev - 1, 1)),
        next: () => setPage((prev) => Math.min(prev + 1, data.meta.pages)),
        last: () => setPage(data.meta.pages),
        onSelectChange: (value: string) => {
            setLimit(Number(value))
            setPage(1)
        }
    }

    const statusOrder = ['未着手', '調査中', '解決済み', '失敗']

    const sortDate = data.status_counts.sort((a, b) => {
        return statusOrder.indexOf(a.status) - statusOrder.indexOf(b.status)
    })

    if (isLoading) return <Loading />

    return (
        <div className="grid gap-4  md:grid-cols-2 lg:grid-cols-4">
            {sortDate.map((status, index) => {
                return status.status !== '解決済み' ? <CardCounter key={index} title={status.status} countSum={sortDate[index].count} /> : ''
            })}

            {!isLoading ? (
                <Card className="col-span-2 md:col-span-2 lg:col-span-4">
                    <SearchComponent
                        initialSelectedStatuses={checkSatatus}
                        initialSelectAssignees={checkAssignees}
                        uniqueAssignees={data.unique_assignees}
                        initialDateRange={initialDateRange}
                        onSearchAction={handleSearch}
                    />
                    <CardHeader>
                        <CardTitle></CardTitle>

                        <CardDescription></CardDescription>
                    </CardHeader>
                    <CardContent>
                        <Table>
                            <TableHeader className="sticky top-0 z-10000">
                                <TableRow>
                                    <TableHead className="w-[70px]">ID</TableHead>
                                    <TableHead className="w-[130px]">ステータス</TableHead>
                                    <TableHead className="w-[10px]"></TableHead>
                                    <TableHead className="w-[100px]">判定</TableHead>
                                    <TableHead>日時</TableHead>
                                    <TableHead>内容</TableHead>
                                    <TableHead className="w-[150px]">担当者</TableHead>
                                </TableRow>
                            </TableHeader>
                            <TableBody>
                                {data.data.map((incident) => {
                                    if (!incident.Incident) return null
                                    return (
                                        <TableRow key={incident.ID} onClick={() => onIncidentClick(incident.Incident)} className="cursor-pointer hover:bg-gray-100 dark:hover:bg-gray-700">
                                            <TableCell>{incident.ID}</TableCell>
                                            <TableCell>
                                                <div className="flex items-center gap-2">
                                                    {GetStatusIcon(incident.Incident.Status.Code)}
                                                    <Badge
                                                        variant={
                                                            incident.Incident.Status.Code === 0
                                                                ? 'red'
                                                                : incident.Incident.Status.Code === 1
                                                                  ? 'yellow'
                                                                  : incident.Incident.Status.Code === 99
                                                                    ? 'blue'
                                                                    : 'green'
                                                        }
                                                    >
                                                        {incident.Incident?.Status?.Name}
                                                    </Badge>
                                                </div>
                                            </TableCell>
                                            <TableCell> {incident.Incident.Vender !== 0 ? <Wrench /> : ''}</TableCell>
                                            <TableCell>
                                                <Badge variant={incident.Incident.APIData.Judgment === '静観' ? 'green' : 'red'}>{incident.Incident.APIData.Judgment}</Badge>
                                            </TableCell>
                                            <TableCell>{format(fromUnixTime(incident.Incident.APIData.CreatedAt), 'yyyy-MM-dd HH:mm')}</TableCell>
                                            <TableCell>
                                                <div className="font-medium">{incident.subject}</div>
                                                <div className="text-sm text-muted-foreground">{incident.Incident.APIData.Sender}</div>
                                            </TableCell>
                                            <TableCell>{incident.Incident.Assignee || '-'}</TableCell>
                                        </TableRow>
                                    )
                                })}
                            </TableBody>
                        </Table>
                    </CardContent>

                    <PageNation props={data.meta} handlers={handlers} displayLimit={limit} />
                </Card>
            ) : (
                <Loading />
            )}
        </div>
    )
}
export default TopTable
