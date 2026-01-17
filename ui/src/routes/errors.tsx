import { createFileRoute } from "@tanstack/react-router"
import { ErrorsView } from "~/components/errors-view"

export const Route = createFileRoute("/errors")({
   component: Errors,
})

function Errors() {
   return <ErrorsView serviceFilter={null} />
}

