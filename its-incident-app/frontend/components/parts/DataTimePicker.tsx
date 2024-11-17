'use client'
import { format } from 'date-fns'
import { useState } from 'react'
import { FormControl, FormItem, FormLabel, FormMessage } from '@/components/ui/form'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover'
import { Button } from '@/components/ui/button'
import { Calendar } from '@/components/ui/calendar'
import { CalendarIcon, ClockIcon } from 'lucide-react'
import { Input } from '@/components/ui/input'
import { cn } from '@/lib/utils'

const DateTimePicker = ({ field, label }: { field: any; label: string }) => {
    const [date, setDate] = useState<Date | undefined>(field.value)
    const [time, setTime] = useState(format(field.value || new Date(), 'HH:mm'))

    const handleDateSelect = (newDate: Date | undefined) => {
        if (newDate) {
            const [hours, minutes] = time.split(':').map(Number)
            const updatedDate = new Date(newDate)
            updatedDate.setHours(hours)
            updatedDate.setMinutes(minutes)
            setDate(updatedDate)
            field.onChange(updatedDate)
        }
    }

    const handleTimeChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        const newTime = e.target.value
        setTime(newTime)
        if (date) {
            const [hours, minutes] = newTime.split(':').map(Number)
            const updatedDate = new Date(date)
            updatedDate.setHours(hours)
            updatedDate.setMinutes(minutes)
            field.onChange(updatedDate)
        }
    }

    return (
        <FormItem className="flex flex-col">
            <FormLabel>{label}</FormLabel>
            <Popover>
                <PopoverTrigger asChild>
                    <FormControl>
                        <Button variant={'outline'} className={cn('w-full pl-3 text-left font-normal', !date && 'text-muted-foreground')}>
                            {date ? format(date, 'yyyy年MM月dd日 HH:mm') : <span>日時を選択</span>}
                            <CalendarIcon className="ml-auto h-4 w-4 opacity-50" />
                        </Button>
                    </FormControl>
                </PopoverTrigger>
                <PopoverContent className="w-auto p-0" align="start">
                    <Calendar mode="single" selected={date} onSelect={handleDateSelect} initialFocus />
                    <div className="p-3 border-t flex items-center">
                        <ClockIcon className="mr-2 h-4 w-4 opacity-50" />
                        <Input type="time" value={time} onChange={handleTimeChange} className="w-full" />
                    </div>
                </PopoverContent>
            </Popover>
            <FormMessage />
        </FormItem>
    )
}

export default DateTimePicker
