import { createFileRoute } from "@tanstack/react-router"
import { TraceTimeline } from "~/components/trace-timeline"

export const Route = createFileRoute("/traces/$traceId")({
   component: Trace,
})

function Trace() {
   const { traceId } = Route.useParams()
   const trace = {
      id: traceId,
      name: "Trace " + traceId,
      service: "unknown",
   }

   return <TraceTimeline trace={trace} onClose={() => {}} />
}
