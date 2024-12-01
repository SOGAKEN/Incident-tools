'use client'

import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { Calendar } from '@/components/ui/calendar'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover'
import { Checkbox } from '@/components/ui/checkbox'
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from '@/components/ui/dropdown-menu'
import { format } from 'date-fns'
import { Calendar as CalendarIcon, Plus, Activity, CheckCircle, AlertCircle } from 'lucide-react'
import { ja } from 'date-fns/locale'
import { DateRange } from 'react-day-picker'

const statusItems = [
    { icon: AlertCircle, label: '未着手', count: 0 },
    { icon: Activity, label: '調査中', count: 0 },
    { icon: CheckCircle, label: '解決済み', count: 0 }
] as const

interface SearchComponentProps {
    initialSelectedStatuses?: string[]
    initialSelectAssignees?: string[]
    uniqueAssignees?: string[]
    initialDateRange?: DateRange
    onSearchAction: (selectedStatuses: string[], dateRange: DateRange | undefined, selectUniqueAssignees: string[]) => Promise<void>
}

export function SearchComponent({ initialSelectedStatuses = [], initialSelectAssignees = [], uniqueAssignees = [], initialDateRange, onSearchAction }: SearchComponentProps) {
    const [selectedStatuses, setSelectedStatuses] = useState<string[]>(initialSelectedStatuses)
    const [selectUniqueAssignees, setUniqueAssignees] = useState<string[]>(initialSelectAssignees)
    const [dateRange, setDateRange] = useState<DateRange | undefined>(initialDateRange)

    const assigneeList = uniqueAssignees

    const handleStatusChange = (status: string) => {
        setSelectedStatuses((prev) => (prev.includes(status) ? prev.filter((s) => s !== status) : [...prev, status]))
    }

    const handleAssigneeChange = (status: string) => {
        setUniqueAssignees((prev) => (prev.includes(status) ? prev.filter((s) => s !== status) : [...prev, status]))
    }
    const handleSearch = () => {
        onSearchAction(selectedStatuses, dateRange, selectUniqueAssignees)
    }

    return (
        <div className="flex items-center space-x-4 p-4 bg-background">
            <DropdownMenu>
                <DropdownMenuTrigger asChild>
                    <Button variant="outline" className="border-dashed">
                        <Plus className="w-4 h-4 mr-2" />
                        ステータス
                        {selectedStatuses.length > 0 && `(${selectedStatuses.length})`}
                    </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent className="w-56" align="start">
                    {statusItems.map((item, index) => (
                        <DropdownMenuItem key={index} className="flex items-center justify-between" onSelect={(e) => e.preventDefault()}>
                            <div className="flex items-center">
                                <Checkbox id={`status-${index}`} checked={selectedStatuses.includes(item.label)} onCheckedChange={() => handleStatusChange(item.label)} />
                                <label htmlFor={`status-${index}`} className="flex items-center ml-2 cursor-pointer">
                                    <item.icon className="w-4 h-4 mr-2" />
                                    {item.label}
                                </label>
                            </div>
                        </DropdownMenuItem>
                    ))}
                </DropdownMenuContent>
            </DropdownMenu>

            <DropdownMenu>
                <DropdownMenuTrigger asChild>
                    <Button variant="outline" className="border-dashed">
                        <Plus className="w-4 h-4 mr-2" />
                        担当者
                        {selectUniqueAssignees.length > 0 && `(${selectUniqueAssignees.length})`}
                    </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent className="w-56" align="start">
                    {assigneeList.map((item, index) => (
                        <DropdownMenuItem key={index} className="flex items-center justify-between" onSelect={(e) => e.preventDefault()}>
                            <div className="flex items-center">
                                <Checkbox id={`assignee-${index}`} checked={selectUniqueAssignees.includes(item)} onCheckedChange={() => handleAssigneeChange(item)} />
                                <label htmlFor={`assignee-${index}`} className="flex items-center ml-2 cursor-pointer">
                                    {item}
                                </label>
                            </div>
                        </DropdownMenuItem>
                    ))}
                </DropdownMenuContent>
            </DropdownMenu>

            <Popover>
                <PopoverTrigger asChild>
                    <Button variant="outline" className="w-[300px] justify-start text-left font-normal">
                        <CalendarIcon className="mr-2 h-4 w-4" />
                        {dateRange?.from ? (
                            dateRange.to ? (
                                <>
                                    {format(dateRange.from, 'yyyy年MM月dd日', { locale: ja })} - {format(dateRange.to, 'yyyy年MM月dd日', { locale: ja })}
                                </>
                            ) : (
                                format(dateRange.from, 'yyyy年MM月dd日', { locale: ja })
                            )
                        ) : (
                            <span>日付範囲を選択</span>
                        )}
                    </Button>
                </PopoverTrigger>
                <PopoverContent className="w-auto p-0" align="start">
                    <Calendar initialFocus mode="range" defaultMonth={dateRange?.from} selected={dateRange} onSelect={setDateRange} numberOfMonths={2} />
                </PopoverContent>
            </Popover>

            <Button onClick={handleSearch}>検索</Button>
        </div>
    )
}
