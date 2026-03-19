import { useQuery } from "@tanstack/react-query"

export interface Stats {
   total_events: number
   error_rate: number
   active_services: number
}

const mockStats: Stats = {
   total_events: 2600,
   error_rate: 0.5,
   active_services: 3,
}

export function useStats() {
   return useQuery({
      queryKey: ["stats"],
      queryFn: async (): Promise<Stats> => {
         try {
            const response = await fetch("/api/stats")
            if (!response.ok) {
               throw new Error("Failed to fetch stats")
            }
            return response.json()
         } catch {
            return mockStats
         }
      },
      refetchInterval: 5000,
      retry: false,
   })
}
