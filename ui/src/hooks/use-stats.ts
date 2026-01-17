import { useQuery } from "@tanstack/react-query"

interface Stats {
   total_events: number
   error_rate: number
   active_services: number
}

export function useStats() {
   return useQuery({
      queryKey: ["stats"],
      queryFn: async (): Promise<Stats> => {
         const response = await fetch("/api/stats")
         if (!response.ok) {
            throw new Error("Failed to fetch stats")
         }
         return response.json()
      },
      refetchInterval: 5000,
   })
}

