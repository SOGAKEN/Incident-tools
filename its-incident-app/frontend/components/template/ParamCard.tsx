import GetStatusIcon from '@/components/parts/GetStatusIcon'
import StatusDisplay from '@/components/parts/StatusDisplay'

type ParamCard = {
    title: string
    content: string | null | undefined | number
    status?: string
    badge?: boolean
    icon?: boolean
    code: number
}

const ParamCard: React.FC<ParamCard> = ({ title, content, status, badge, icon, code }) => {
    return (
        <div className="flex items-center gap-2">
            <h4 className="font-semibold text-black mix-blend-normal dark:text-white">{title}:</h4>
            {icon ? (
                <div className="flex items-center gap-2">
                    {status && GetStatusIcon(code)}
                    <StatusDisplay text={status} code={code} />
                </div>
            ) : badge ? (
                <div className="flex items-center gap-2">
                    <StatusDisplay text={status} code={code} />
                </div>
            ) : (
                <div className="flex items-center gap-2">
                    <div className="text-black dark:text-white">{content}</div>
                </div>
            )}
        </div>
    )
}

export default ParamCard
