import { ChevronLeftIcon, ChevronRightIcon, DoubleArrowLeftIcon, DoubleArrowRightIcon, CaretSortIcon } from '@radix-ui/react-icons'

import { Button } from '@/components/ui/button'
import { Select, SelectContent, SelectGroup, SelectItem, SelectLabel, SelectTrigger, SelectValue } from '@/components/ui/select'

type Meta = {
    total: number
    pages: number
    page: number
    limit: number
}

type OnClick = {
    handlers: {
        first: () => void
        previous: () => void
        next: () => void
        last: () => void
        onSelectChange: (value: string) => void
    }
    props: Meta
    displayLimit: number
}

const PageNation = ({ props, handlers, displayLimit }: OnClick) => {
    return (
        <div className="w-full flex gap-8 justify-end items-center p-[5px]">
            <div className="flex items-center gap-2">
                <div className="text-sm font-semibold">Rows per page</div>
                <Select onValueChange={handlers.onSelectChange}>
                    <SelectTrigger className="w-[80px] h-[35px]">
                        <SelectValue placeholder={displayLimit} />
                    </SelectTrigger>
                    <SelectContent>
                        <SelectItem value="10">10</SelectItem>
                        <SelectItem value="20">20</SelectItem>
                        <SelectItem value="30">30</SelectItem>
                        <SelectItem value="40">40</SelectItem>
                        <SelectItem value="50">50</SelectItem>
                    </SelectContent>
                </Select>
            </div>
            <div className="flex items-center gap-1">
                <span className="text-sm font-semibold">Page</span>
                <span className="text-sm font-semibold">{props.page}</span>
                <span className="text-sm font-semibold">of</span>
                <span className="text-sm font-semibold">{props.pages}</span>
            </div>
            <div className="flex items-center gap-1">
                <Button variant="outline" disabled={props.page === 1 ? true : false} onClick={handlers.first}>
                    <DoubleArrowLeftIcon className="h-4 w-4" />
                </Button>
                <Button variant="outline" disabled={props.page === 1 ? true : false} onClick={handlers.previous}>
                    <ChevronLeftIcon className="h-4 w-4" />
                </Button>
                <Button variant="outline" disabled={props.page === props.pages ? true : false} onClick={handlers.next}>
                    <ChevronRightIcon className="h-4 w-4" />
                </Button>
                <Button variant="outline" disabled={props.page === props.pages ? true : false} onClick={handlers.last}>
                    <DoubleArrowRightIcon className="h-4 w-4" />
                </Button>
            </div>
        </div>
    )
}

export default PageNation
