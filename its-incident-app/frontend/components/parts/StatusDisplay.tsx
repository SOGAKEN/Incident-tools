import { Badge } from '@/components/ui/badge'

type Color = {
    text: string | undefined
    red: string
    yellow: string
    green: string
}

const StatusDisplay: React.FC<Color> = ({ text, red, yellow, green }) => {
    return (
        <div className="flex items-center gap-2">
            <Badge variant={text === red ? 'red' : text === yellow ? 'yellow' : text === green ? 'green' : 'red'}>{text}</Badge>
        </div>
    )
}

export default StatusDisplay
