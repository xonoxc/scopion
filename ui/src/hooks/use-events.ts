import { useQuery } from "@tanstack/react-query"
import type { Event } from "./use-livefeed"

export function useEvents(limit: number = 100) {
   return useQuery({
      queryKey: ["events", limit],
      queryFn: async (): Promise<Event[]> => {
         const response = await fetch(`/api/events?limit=${limit}`)
         if (!response.ok) {
            throw new Error("Failed to fetch events")
         }
         return response.json()
      },
      refetchInterval: 5000,
   })
}
