import GetStatusIcon from '@/components/parts/GetStatusIcon'
import StatusDisplay from '@/components/parts/StatusDisplay'

type ParamCard = {
    title: string
    content: string | null | undefined | number
    status?: string
    badge?: boolean
    icon?: boolean
}

const ParamCard: React.FC<ParamCard> = ({ title, content, status, badge, icon }) => {
    return (
        <div className="flex items-center gap-2">
            <h4 className="font-semibold text-black mix-blend-normal dark:text-white">{title}:</h4>
            {icon ? (
                <div className="flex items-center gap-2">
                    {status && GetStatusIcon(status)}
                    <StatusDisplay text={status} red="未着手" yellow="調査中" green="解決済み" />
                </div>
            ) : badge ? (
                <div className="flex items-center gap-2">
                    <StatusDisplay text={status} red={'要対応'} yellow="" green="静観" />
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
