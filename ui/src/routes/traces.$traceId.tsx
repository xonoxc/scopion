import { createFileRoute, useNavigate } from "@tanstack/react-router"
import { TraceTimeline } from "~/components/trace-timeline"
import { useQuery } from "@tanstack/react-query"

interface TraceEvent {
   id: string
   timestamp: string
   level: string
   service: string
   name: string
   trace_id?: string
   data?: Record<string, unknown>
}

export const Route = createFileRoute("/traces/$traceId")({
   component: Trace,
})

function Trace() {
   const { traceId } = Route.useParams()
   const navigate = useNavigate()

   const { data: events, isLoading } = useQuery({
      queryKey: ["trace-events", traceId],
      queryFn: async (): Promise<TraceEvent[]> => {
         const response = await fetch(`/api/trace-events?trace_id=${traceId}`)
         if (!response.ok) {
            throw new Error("Failed to fetch trace events")
         }
         return response.json()
      },
   })

   if (isLoading) {
      return (
         <div className="flex h-full items-center justify-center">
            <p className="text-sm text-muted-foreground">Loading trace...</p>
         </div>
      )
   }

   if (!events || events.length === 0) {
      return (
         <div className="flex h-full items-center justify-center">
            <p className="text-sm text-muted-foreground">Trace not found</p>
         </div>
      )
   }

   const trace = {
      id: traceId,
      name: events[0]?.name || `Trace ${traceId}`,
      service: events[0]?.service || "unknown",
   }

   return <TraceTimeline trace={trace} events={events} onClose={() => navigate({ to: "/traces" })} />
}
