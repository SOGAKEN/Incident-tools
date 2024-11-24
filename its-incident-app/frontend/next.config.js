/** @type {import('next').NextConfig} */
const nextConfig = {
    experimental: {
        serverActions: {}
    },
    async redirects() {
        return [
            {
                source: '/',
                destination: '/login',
                permanent: false
            }
        ]
    },
    output: 'standalone'
}

module.exports = nextConfig
