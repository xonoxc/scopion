import { useState, useEffect } from "react"

export interface ServerStatus {
   demo_enabled: boolean
   version: string
}

export function useServerStatus() {
   const [status, setStatus] = useState<ServerStatus | null>(null)
   const [loading, setLoading] = useState(true)
   const [error, setError] = useState<string | null>(null)

   useEffect(() => {
      const fetchStatus = async () => {
         try {
            const response = await fetch("/api/status")
            if (!response.ok) {
               throw new Error("Failed to fetch server status")
            }
            const data = await response.json()
            setStatus(data)
         } catch (err) {
            setError(err instanceof Error ? err.message : "Unknown error")
         } finally {
            setLoading(false)
         }
      }

      fetchStatus()
   }, [])

   return { status, loading, error }
}

