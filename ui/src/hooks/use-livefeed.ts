import { useEffect, useState } from "react"
import { useEvents } from "./use-events"
import type { LiveFeedProps } from "~/components/live-feed"

export interface Event {
   id: string
   timestamp: string
   level: "info" | "warn" | "error"
   service: string
   name: string
   trace_id: string
}

export function useLiveFeed({ serviceFilter }: Omit<LiveFeedProps, "onSelectTrace">) {
   const { data: apiEvents = [], isLoading } = useEvents(100)
   const [liveEvents, setLiveEvents] = useState<Event[]>([])
   const [copiedId, setCopiedId] = useState<string | null>(null)
   const [isPaused, setIsPaused] = useState(false)

   const allEvents = [...liveEvents, ...apiEvents].slice(0, 100)

   useEffect(() => {
      if (isPaused) return

      const eventSource = new EventSource("/api/live")

      eventSource.onmessage = event => {
         const newEvent: Event = JSON.parse(event.data)
         setLiveEvents(prev => [newEvent, ...prev].slice(0, 50))
      }

      eventSource.onerror = () => {
         console.error("EventSource failed.", eventSource.readyState)
      }

      return () => {
         eventSource.close()
      }
   }, [isPaused])

   const filteredEvents = serviceFilter
      ? allEvents.filter(e => e.service === serviceFilter)
      : allEvents

   const copyTraceId = (traceId: string) => {
      navigator.clipboard.writeText(traceId)
      setCopiedId(traceId)
      setTimeout(() => setCopiedId(null), 2000)
   }

   return {
      filteredEvents,
      copiedId,
      isPaused,
      setIsPaused,
      copyTraceId,
      isLoading,
   }
}
