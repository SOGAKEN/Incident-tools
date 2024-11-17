import { AlertCircle, CheckCircle, Activity } from 'lucide-react'

const GetStatusIcon = (status: string) => {
    switch (status) {
        case '未着手':
            return <AlertCircle className="h-5 w-5 text-red-500" />
        case '調査中':
            return <Activity className="h-5 w-5 text-yellow-500" />
        case '解決済み':
            return <CheckCircle className="h-5 w-5 text-green-500" />
        default:
            return null
    }
}

export default GetStatusIcon
