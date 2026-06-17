import { useQuery } from "@tanstack/react-query"
import { attempt } from "~/lib/attempt"

interface TraceInfo {
   id: string
   name: string
   service: string
   duration: number
   spans: number
   timestamp: string
   has_error: boolean
}

const mockTraces: TraceInfo[] = [
   {
      id: "trace-1",
      name: "HTTP Request",
      service: "web-service",
      duration: 150,
      spans: 3,
      timestamp: new Date(Date.now() - 10000).toISOString(),
      has_error: false,
   },
   {
      id: "trace-2",
      name: "Database Query",
      service: "db-service",
      duration: 200,
      spans: 2,
      timestamp: new Date(Date.now() - 20000).toISOString(),
      has_error: true,
   },
   {
      id: "trace-3",
      name: "API Call",
      service: "api-service",
      duration: 75,
      spans: 1,
      timestamp: new Date(Date.now() - 30000).toISOString(),
      has_error: false,
   },
]

export function useTraces(limit: number = 50) {
   return useQuery({
      queryKey: ["traces", limit],
      queryFn: async (): Promise<TraceInfo[]> => {
         const result = await attempt(
            fetch(`/api/traces?limit=${limit}`).then(async res => {
               if (!res.ok) {
                  throw new Error("Failed to fetch traces")
               }
               return res.json()
            })
         )
         return result.match(
            data => data,
            e => {
               console.warn("useTraces: API fetch failed, using mock data:", e)
               return mockTraces.slice(0, limit)
            }
         )
      },
      refetchInterval: 10000,
      retry: false,
   })
}
