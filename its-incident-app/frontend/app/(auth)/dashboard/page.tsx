'use client'
import TopTable from '@/components/template/TopTable'
import { useContext, useEffect, useState } from 'react'
import { type Incident, type IncidentsApiResponse } from '@/typs/incident'
import DialogWindow from '@/components/parts/Dailog'
import Loading from '@/components/template/Loading'

const Dashboard = () => {
    const [selectedIncident, setSelectedIncident] = useState<Incident | null>(null)
    const [isDialogOpen, setIsDialogOpen] = useState(false)
    const [isLoading, setIsLoading] = useState(true)

    const handleIncidentClick = (incident: Incident) => {
        setSelectedIncident(incident)
        setIsDialogOpen(true)
    }

    useEffect(() => {
        setIsLoading(false)
    }, [])

    if (isLoading) return <Loading />

    return (
        <>
            <TopTable onIncidentClick={handleIncidentClick} />
            <DialogWindow isOpen={isDialogOpen} onClose={() => setIsDialogOpen(false)} incident={selectedIncident} />
        </>
    )
}

export default Dashboard
