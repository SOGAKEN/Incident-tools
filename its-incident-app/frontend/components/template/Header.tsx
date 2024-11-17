'use client'

import Image from 'next/image'
import Link from 'next/link'
import React, { useContext } from 'react'
import { UserContext } from '@/lib/UserContext'
import logo from '@/assets/logo.png'
import { Button } from '@/components/ui/button'
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from '@/components/ui/dropdown-menu'
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'
import { Bell, Clipboard, LogOut } from 'lucide-react'
import { useFetch } from '@/hooks/useFetch'
import { useRouter } from 'next/navigation'
import { ModeToggle } from '../parts/DarkModeToggle'

interface LogoutResponse {
    success: boolean
}

const Header: React.FC = () => {
    const router = useRouter()
    const userData = useContext(UserContext)
    const { execute } = useFetch<LogoutResponse>('/api/logout', {
        method: 'POST',
        body: { email: userData?.email },
        onSuccess: () => {
            document.cookie = 'session_id=; expires=Thu, 01 Jan 1970 00:00:00 GMT; path=/; secure; samesite=strict'

            router.push('/login')
            router.refresh()
        },
        onError: () => {
            console.error('ログアウト失敗しました。')
        }
    })

    if (!userData) {
        return <div>ユーザーデータがありません</div>
    }

    const handleSignOut = () => {
        document.cookie = 'session_id=; max-age=0 path=/login'
        // ログアウト処理を実装
        execute()
    }

    return (
        <>
            {userData && (
                <header className="sticky top-0 z-50 w-full bg-black text-white">
                    <div className="flex h-16 items-center justify-between px-[50px]">
                        <Link href="/" className="flex items-center space-x-2">
                            <Image src={logo} alt="IncidentTolls Logo" height={40} priority={true} />
                            <span className="hidden font-bold sm:inline-block">{process.env.NEXT_PUBLIC_MAIN_NAME}</span>
                        </Link>
                        <nav className="flex items-center space-x-4">
                            <Link href="/dashboard">
                                <Button variant="ghost" size="sm" className="text-sm font-medium text-white hover:text-black dark:hover:text-white">
                                    <Bell className="mr-2 h-4 w-4" />
                                    アラート
                                </Button>
                            </Link>
                            <Link href="/work">
                                <Button variant="ghost" size="sm" className="text-sm font-medium text-white hover:text-black dark:hover:text-white">
                                    <Clipboard className="mr-2 h-4 w-4" />
                                    作業連絡
                                </Button>
                            </Link>
                        </nav>
                        <div className="flex space-x-2 justify-center items-center">
                            <ModeToggle />
                            <DropdownMenu>
                                <DropdownMenuTrigger asChild>
                                    <Button variant="ghost" className="relative h-8 w-8 rounded-full">
                                        <Avatar className="h-8 w-8 text-black hover:text-green-500 dark:text-white">
                                            <AvatarImage src={userData.image || ''} alt={userData.name || 'ユーザーアバター'} />
                                            <AvatarFallback>{`${userData.name?.slice(0, 1)}` || `${userData.email?.slice(0, 1).toUpperCase()}`}</AvatarFallback>
                                        </Avatar>
                                    </Button>
                                </DropdownMenuTrigger>
                                <DropdownMenuContent className="w-56" align="end" forceMount>
                                    <DropdownMenuItem className="flex-col items-start">
                                        <div className="text-sm font-medium">{userData.name}</div>
                                        <div className="text-xs text-muted-foreground">{userData.email}</div>
                                    </DropdownMenuItem>
                                    <DropdownMenuItem onClick={handleSignOut}>
                                        <LogOut className="mr-2 h-4 w-4" />
                                        <span>ログアウト</span>
                                    </DropdownMenuItem>
                                </DropdownMenuContent>
                            </DropdownMenu>
                        </div>
                    </div>
                </header>
            )}
        </>
    )
}

export default Header
