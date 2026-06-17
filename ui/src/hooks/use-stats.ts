import { useQuery } from "@tanstack/react-query"
import { attempt } from "~/lib/attempt"

export interface Stats {
   total_events: number
   error_rate: number
   active_services: number
   p50_latency: number
}

const mockStats: Stats = {
   total_events: 2600,
   error_rate: 0.5,
   active_services: 3,
   p50_latency: 55,
}

export function useStats() {
   return useQuery({
      queryKey: ["stats"],
      queryFn: async (): Promise<Stats> => {
         const result = await attempt(
            fetch("/api/stats").then(async res => {
               if (!res.ok) {
                  throw new Error("Failed to fetch stats")
               }
               return res.json()
            })
         )
         return result.match(
            data => data,
            e => {
               console.warn("useStats: API fetch failed, using mock data:", e)
               return mockStats
            }
         )
      },
      refetchInterval: 5000,
      retry: false,
   })
}
