import { createFileRoute, Outlet, useLocation } from "@tanstack/react-router"
import { TracesView } from "~/components/traces-view"

export const Route = createFileRoute("/traces")({
   component: Traces,
})

function Traces() {
   const { pathname } = useLocation()
   if (pathname === "/traces") {
      return <TracesView serviceFilter={null} />
   }
   return <Outlet />
}

