import { createFileRoute } from "@tanstack/react-router"
import { LiveFeed } from "~/components/live-feed"

export const Route = createFileRoute("/live")({
   component: Live,
})

function Live() {
   return <LiveFeed serviceFilter={null} />
}

