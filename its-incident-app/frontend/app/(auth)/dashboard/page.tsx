'use client'
import TopTable from '@/components/template/TopTable'
import { useContext, useEffect, useState } from 'react'
import { UserContext } from '@/lib/UserContext'
import { type Incident, type IncidentsApiResponse } from '@/typs/incident'
import { redirect, useRouter } from 'next/navigation'
import DialogWindow from '@/components/parts/Dailog'
import Loading from '@/components/template/Loading'

const Dashboard = () => {
    const router = useRouter()
    const [selectedIncident, setSelectedIncident] = useState<Incident | null>(null)
    const [isDialogOpen, setIsDialogOpen] = useState(false)
    const [isLoading, setIsLoading] = useState(true)

    const userData = useContext(UserContext)

    useEffect(() => {
        if (!userData?.name) {
            redirect('/profile')
        } else {
            setIsLoading(false)
        }
    }, [userData, router])

    const handleIncidentClick = (incident: Incident) => {
        setSelectedIncident(incident)
        setIsDialogOpen(true)
    }

    useEffect(() => {
        setIsLoading(false)
    })

    if (isLoading) return <Loading />

    return (
        <>
            <TopTable onIncidentClick={handleIncidentClick} />
            <DialogWindow isOpen={isDialogOpen} onClose={() => setIsDialogOpen(false)} incident={selectedIncident} />
        </>
    )
}

export default Dashboard
