'use client'
import { useFetch } from '@/hooks/useFetch'
import { EmailData, type Incident, type IncidentsApiResponse } from '@/typs/incident'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Wrench } from 'lucide-react'
import { Checkbox } from '@/components/ui/checkbox'
import { format, fromUnixTime } from 'date-fns'
import { useEffect, useState } from 'react'
import CardCounter from '../parts/CardCounter'
import Loading from './Loading'
import PageNation from '../parts/PageNation'
import GetStatusIcon from '../parts/GetStatusIcon'
import { SearchComponent } from '../parts/Search'
import { DateRange } from 'react-day-picker'

import { RightSidebar } from '@/components/parts/RightSidebar'

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
    const [selectedRows, setSelectedRows] = useState<Set<number>>(new Set())
    const [selectAll, setSelectAll] = useState(false)
    const [checkedMessageId, setCheckedMessageId] = useState<string[]>([])
    const [allInfo, setAllInfo] = useState<EmailData[]>([])
    const [uniqHost, setUniqHost] = useState<string[]>([])

    const { data, isLoading } = useFetch<IncidentsApiResponse>(`/api/getIncidentAll?page=${page}&limit=${limit}&status=${queryParam}${dateParam}&assignee=${queryAssignee}`, {
        useSWR: true,
        swrOptions: {
            refreshInterval: 5000
        }
    })

    // Reset selection when page changes
    useEffect(() => {
        setSelectedRows(new Set())
        setSelectAll(false)
    }, [page])

    const handleSearch = async (selectedStatuses: string[], dateRange: DateRange | undefined, selectUniqueAssignees: string[]): Promise<void> => {
        const query = selectedStatuses.length > 0 ? selectedStatuses.join('%2C') : ''
        setQueryParam(query)
        setcheckStatas(selectedStatuses)
        setPage(1)

        const assignee = selectUniqueAssignees.length > 0 ? selectUniqueAssignees.join('%2C') : ''
        setAssignee(assignee)
        setcheckAssignees(selectUniqueAssignees)

        if (dateRange?.from && dateRange?.to) {
            setFromStatus(dateRange.from)
            setToStatus(dateRange.to)
            setDateParam(`&from=${format(dateRange.from, 'yyyy-MM-dd 00:00')}&to=${format(dateRange.to, 'yyyy-MM-dd 23:59')}`)
        } else {
            setFromStatus(undefined)
            setToStatus(undefined)
            setDateParam('')
        }
    }
    const getSelectedMessageIds = (): string[] => {
        if (!data) return []
        return data.data.filter((incident) => selectedRows.has(incident.ID)).map((incident) => incident.message_id)
    }
    const getSelectedIncidents = () => {
        if (!data) return []
        return data.data.filter((incident) => selectedRows.has(incident.ID))
    }
    const getSelectedUniqueHosts = () => {
        if (!data) return []

        const uniqueHosts = new Set<string>()

        data.data
            .filter((incident) => selectedRows.has(incident.ID))
            .forEach((incident) => {
                const host = incident.Incident?.APIData?.Host
                if (host) {
                    uniqueHosts.add(host)
                }
            })

        return Array.from(uniqueHosts)
    }

    useEffect(() => {
        const messageIds = getSelectedMessageIds()
        const allinfomation = getSelectedIncidents()
        const uniqhost = getSelectedUniqueHosts()
        setCheckedMessageId(messageIds)
        setAllInfo(allinfomation)
        setUniqHost(uniqhost)
    }, [selectedRows, data])

    const handleAllCheckChange = (checked: boolean) => {
        setSelectAll(checked)
        if (checked && data) {
            const allIds = data.data.map((incident) => incident.ID)
            setSelectedRows(new Set(allIds))
        } else {
            setSelectedRows(new Set())
        }
    }

    const handleRowCheckChange = (checked: boolean, id: number) => {
        const newSelected = new Set(selectedRows)
        if (checked) {
            newSelected.add(id)
        } else {
            newSelected.delete(id)
        }
        setSelectedRows(newSelected)

        // Update selectAll state based on whether all visible rows are selected
        if (data) {
            setSelectAll(newSelected.size === data.data.length)
        }
    }

    const initialDateRange: DateRange | undefined =
        fromStatus && toStatus
            ? {
                  from: fromStatus,
                  to: toStatus
              }
            : undefined

    if (!data) return null
    if (isLoading) return <Loading />

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

    return (
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
            {sortDate.map((status, index) => (status.status !== '解決済み' ? <CardCounter key={index} title={status.status} countSum={sortDate[index].count} /> : null))}

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
                                <TableHead className="w-[30px]">
                                    <div className="flex items-center">
                                        <Checkbox checked={selectAll} onCheckedChange={handleAllCheckChange} />
                                    </div>
                                </TableHead>
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
                                const isSelected = selectedRows.has(incident.ID)

                                return (
                                    <TableRow
                                        key={incident.ID}
                                        onClick={() => onIncidentClick(incident.Incident)}
                                        className={`cursor-pointer hover:bg-gray-100 dark:hover:bg-gray-700 
                                            ${isSelected ? 'bg-gray-100 dark:bg-gray-700' : ''}`}
                                    >
                                        <TableCell onClick={(e) => e.stopPropagation()}>
                                            <div className="flex items-center">
                                                <Checkbox checked={selectedRows.has(incident.ID)} onCheckedChange={(checked) => handleRowCheckChange(!!checked, incident.ID)} />
                                            </div>
                                        </TableCell>
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
                                        <TableCell>{incident.Incident.Vender !== 0 ? <Wrench /> : ''}</TableCell>
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
                    {checkedMessageId.length === 0 ? (
                        ''
                    ) : (
                        <div className="mt-1">
                            <RightSidebar mesasge_id={checkedMessageId} uniqe_host={uniqHost} />
                        </div>
                    )}
                </CardContent>
                <PageNation props={data.meta} handlers={handlers} displayLimit={limit} />
            </Card>
        </div>
    )
}

export default TopTable
