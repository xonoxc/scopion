import { createFileRoute, useNavigate } from "@tanstack/react-router"
import { TraceTimeline } from "~/components/trace-timeline"
import { useTraceEvents } from "~/hooks/use-events"

export const Route = createFileRoute("/traces/$traceId")({
   component: Trace,
})

function getTraceMeta(
   traceId: string,
   firstEvent: { name?: string | null; service?: string | null }
) {
   return {
      id: traceId,
      name: firstEvent.name ?? `Trace ${traceId}`,
      service: firstEvent.service ?? "unknown",
   }
}

function Trace() {
   const { traceId } = Route.useParams()
   const navigate = useNavigate()
   const { data: events, isLoading } = useTraceEvents(traceId)

   if (isLoading) return <LoadingState />
   if (!events?.length) return <EmptyState />

   const trace = getTraceMeta(traceId, events[0])

   return (
      <TraceTimeline
         trace={trace}
         events={events}
         onClose={() =>
            navigate({
               to: "/traces",
            })
         }
      />
   )
}

function EmptyState() {
   return (
      <div className="flex h-full items-center justify-center">
         <p className="text-sm text-muted-foreground">Trace not found</p>
      </div>
   )
}

function LoadingState() {
   return (
      <div className="flex h-full items-center justify-center">
         <p className="text-sm text-muted-foreground">Loading trace...</p>
      </div>
   )
}
