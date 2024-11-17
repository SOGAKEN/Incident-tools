import Header from '@/components/template/Header'
import UserProvider from '@/components/template/UserProvider'
import { Toaster } from '@/components/ui/toaster'
import { cookies } from 'next/headers'
import { redirect } from 'next/navigation'

export default async function AppLayout({ children }: { children: React.ReactNode }) {
    const cookieStore = await cookies()
    const sessionID = cookieStore.get('session_id')?.value

    if (!sessionID) {
        redirect('/login')
    }

    const response = await fetch(`${process.env.DBPILOT_URL}/profiles`, {
        method: 'GET',
        headers: {
            Authorization: `Bearer ${sessionID}`
        },
        cache: 'no-store'
    })

    if (!response.ok) {
        redirect('/login')
    }

    const userData = await response.json()

    return (
        <UserProvider userData={userData}>
            <Header />
            <main className="p-[20px]">{children}</main>
            <Toaster />
        </UserProvider>
    )
}
