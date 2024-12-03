import { Badge } from '@/components/ui/badge'

type Color = {
    text: string | undefined
    code: number | undefined
}

const StatusDisplay: React.FC<Color> = ({ text, code }) => {
    return (
        <div className="flex items-center gap-2">
            <Badge variant={code === 0 ? 'red' : code === 1 ? 'yellow' : code === 2 ? 'green' : code === 99 ? 'blue' : 'red'}>{text}</Badge>
        </div>
    )
}

export default StatusDisplay
