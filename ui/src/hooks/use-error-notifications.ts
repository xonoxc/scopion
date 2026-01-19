import { useEffect } from "react"
import { useLiveFeed } from "./use-livefeed"
import { useNotificationActions } from "~/stores/notifications"

export function useErrorNotifications() {
   const { filteredEvents } = useLiveFeed({ serviceFilter: null })
   const { addNotification } = useNotificationActions()

   useEffect(() => {
      const latestEvents = filteredEvents.slice(0, 10)

      latestEvents.forEach(event => {
         if (event.level === "error") {
            const existingNotification = document.querySelector(
               `[data-trace-id="${event.trace_id}"]`
            )

            if (!existingNotification) {
               addNotification({
                  type: "error",
                  title: `Error in ${event.service}`,
                  message: event.name,
                  traceId: event.trace_id,
               })

               const marker = document.createElement("div")
               marker.setAttribute("data-trace-id", event.trace_id)
               marker.style.display = "none"
               document.body.appendChild(marker)

               setTimeout(() => {
                  document.body.removeChild(marker)
               }, 30000)
            }
         }
      })
   }, [filteredEvents])
}

