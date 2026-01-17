import { createFileRoute } from "@tanstack/react-router"
import { OverviewView } from "~/components/overview-view"

export const Route = createFileRoute("/overview")({
   component: Overview,
})

function Overview() {
   return <OverviewView />
}

