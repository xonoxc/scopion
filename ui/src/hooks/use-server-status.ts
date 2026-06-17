import { useState, useEffect } from "react"
import { attempt } from "~/lib/attempt"

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
         const result = await attempt(
            fetch("/api/status").then(async res => {
               if (!res.ok) {
                  throw new Error("Failed to fetch server status")
               }
               return res.json()
            })
         )
         result.match(
            data => setStatus(data),
            err => setError(err instanceof Error ? err.message : "Unknown error")
         )
         setLoading(false)
      }

      fetchStatus()
   }, [])

   return { status, loading, error }
}
