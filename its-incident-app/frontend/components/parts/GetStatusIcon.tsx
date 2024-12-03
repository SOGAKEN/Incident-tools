import { AlertCircle, CheckCircle, Activity, Ban } from 'lucide-react'

const GetStatusIcon = (status: number) => {
    switch (status) {
        case 0:
            return <AlertCircle className="h-5 w-5 text-red-500" />
        case 1:
            return <Activity className="h-5 w-5 text-yellow-500" />
        case 2:
            return <CheckCircle className="h-5 w-5 text-green-500" />
        case 99:
            return <Ban className="h-5 w-5 text-blue-500" />
        default:
            return null
    }
}

export default GetStatusIcon
