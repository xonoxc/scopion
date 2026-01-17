import { createFileRoute } from "@tanstack/react-router"
import { ServicesView } from "~/components/services-view"

export const Route = createFileRoute("/services")({
   component: Services,
})

function Services() {
   return <ServicesView />
}

