import './globals.css'
import { ThemeProvider } from '@/components/template/theme-provider'
import type { Metadata } from 'next'
import localFont from 'next/font/local'

const geistSans = localFont({
    src: './fonts/GeistVF.woff',
    variable: '--font-geist-sans',
    weight: '100 900',
    preload: false
})
const geistMono = localFont({
    src: './fonts/GeistMonoVF.woff',
    variable: '--font-geist-mono',
    weight: '100 900',
    preload: false
})

export const metadata: Metadata = {
    title: 'おたすけ丸 | 管理ツール',
    description: 'BC+ itサポート 管理ツール'
}

export default function RootLayout({ children }: { children: React.ReactNode }) {
    return (
        <html lang="ja" suppressHydrationWarning>
            <body className={`${geistSans.variable} ${geistMono.variable} antialiased`}>
                <ThemeProvider attribute="class" defaultTheme="system" enableSystem disableTransitionOnChange>
                    {children}
                </ThemeProvider>
            </body>
        </html>
    )
}
