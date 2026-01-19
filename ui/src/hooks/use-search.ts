import { useQuery } from "@tanstack/react-query"

interface Event {
   id: string
   timestamp: string
   level: "info" | "warn" | "error"
   service: string
   name: string
   trace_id: string
}

export function useSearch(query: string, enabled: boolean = false) {
   return useQuery({
      queryKey: ["search", query],
      queryFn: async (): Promise<Event[]> => {
         if (!query.trim()) return []
         const response = await fetch(`/api/search?q=${encodeURIComponent(query)}`)
         if (!response.ok) {
            throw new Error("Failed to search")
         }
         const data = await response.json()
         return Array.isArray(data) ? data : []
      },
      enabled: enabled && query.trim().length > 0,
   })
}

