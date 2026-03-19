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

interface TraceEvent {
   id: string
   timestamp: string
   level: string
   service: string
   name: string
   trace_id?: string
   data?: Record<string, unknown>
}

export function useTraceEvents(traceId: number) {
   return useQuery({
      queryKey: ["trace-events", traceId],
      queryFn: async (): Promise<TraceEvent[]> => {
         const response = await fetch(`/api/trace-events?trace_id=${traceId}`)
         if (!response.ok) {
            throw new Error("Failed to fetch trace events")
         }
         return response.json()
      },
   })
}
