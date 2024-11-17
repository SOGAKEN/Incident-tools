'use client'

import React from 'react'
import { UserContext } from '@/lib/UserContext'

type UserData = {
    name: string
    email: string
    image: string
    user_id: number
}

interface UserProviderProps {
    userData: UserData
    children: React.ReactNode
}

const UserProvider: React.FC<UserProviderProps> = ({ userData, children }) => {
    return <UserContext.Provider value={userData}>{children}</UserContext.Provider>
}

export default UserProvider
