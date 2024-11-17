'use client'

import { useEffect, useState } from 'react'
import { Switch } from '@/components/ui/switch'
import { Label } from '@/components/ui/label'

interface StatusUpdateProps {
    initialStatus: string
    initialVenderStatus: number
    onStatusUpdate: (newStatus: string) => void
    onVendorContactUpdate: (contacted: boolean) => void
}

export function StatusUpdate({ initialStatus, initialVenderStatus, onStatusUpdate, onVendorContactUpdate }: StatusUpdateProps) {
    const [isResolved, setIsResolved] = useState(initialStatus === '解決済み')
    const [isVendorContacted, setIsVendorContacted] = useState(initialVenderStatus === 1)

    useEffect(() => {
        setIsResolved(initialStatus === '解決済み')
    }, [initialStatus])

    const handleStatusUpdate = (checked: boolean) => {
        setIsResolved(checked)
        onStatusUpdate(checked ? '解決済み' : '調査中')
    }

    const handleVendorContactUpdate = (checked: boolean) => {
        setIsVendorContacted(checked)
        onVendorContactUpdate(checked)
    }

    return (
        <div className="space-y-6">
            <div className="flex flex-col space-y-4">
                <div className="flex items-center justify-between">
                    <Label htmlFor="status-update" className="font-semibold">
                        解決済み
                    </Label>
                    <Switch id="status-update" checked={isResolved} onCheckedChange={handleStatusUpdate} disabled={isResolved} />
                </div>
                <div className="flex items-center justify-between">
                    <Label htmlFor="vendor-contact" className="font-semibold">
                        ベンダーに問い合わせ済み
                    </Label>
                    <Switch id="vendor-contact" checked={isVendorContacted} onCheckedChange={handleVendorContactUpdate} disabled={isVendorContacted} />
                </div>
            </div>
        </div>
    )
}
