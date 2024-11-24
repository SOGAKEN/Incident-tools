import SettingProfilePage from '@/components/template/SettingProfile'
import { Card } from '@/components/ui/card'
import { cookies } from 'next/headers'
import { redirect } from 'next/navigation'

export default async function DashboardPage() {
    const cookieStore = await cookies()
    const email = cookieStore.get('session_id')?.value || ''

    if (!email) {
        redirect('/login')
    }

    return (
        <div className="flex items-center justify-center min-h-screen">
            <Card className="w-full max-w-md p-10">
                <SettingProfilePage email={email} />
            </Card>
        </div>
    )
}
