import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'

type CardCounterProps = {
    title: string
    countSum: number
}

const CardCounter: React.FC<CardCounterProps> = ({ title, countSum }) => {
    return (
        <Card className="col-span-1 md:col-span-1">
            <CardHeader>
                <CardTitle>{title}</CardTitle>
            </CardHeader>
            <CardContent>
                <div className="text-4xl font-bold">{countSum}</div>
                <div className="text-sm text-muted-foreground">件</div>
            </CardContent>
        </Card>
    )
}

export default CardCounter
