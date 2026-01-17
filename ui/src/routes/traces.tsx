import { createFileRoute } from "@tanstack/react-router"
import { TracesView } from "~/components/traces-view"

export const Route = createFileRoute("/traces")({
   component: Traces,
})

function Traces() {
   return <TracesView serviceFilter={null} />
}

