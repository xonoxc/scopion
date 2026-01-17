import { useEffect } from "react"
import { useLiveFeed } from "./use-livefeed"
import { useNotificationActions } from "~/stores/notifications"

export function useErrorNotifications() {
     const { filteredEvents } = useLiveFeed({ serviceFilter: null })
     const { addNotification } = useNotificationActions()

    useEffect(() => {
        // Check for new error events and create notifications
        const latestEvents = filteredEvents.slice(0, 10) // Check recent events

        latestEvents.forEach(event => {
            if (event.level === "error") {
                // Check if we already have a notification for this event (by trace_id)
                const existingNotification = document.querySelector(`[data-trace-id="${event.trace_id}"]`)

                if (!existingNotification) {
                    addNotification({
                        type: "error",
                        title: `Error in ${event.service}`,
                        message: event.name,
                        traceId: event.trace_id,
                    })

                    // Add a data attribute to prevent duplicates
                    const marker = document.createElement("div")
                    marker.setAttribute("data-trace-id", event.trace_id)
                    marker.style.display = "none"
                    document.body.appendChild(marker)

                    // Clean up after some time
                    setTimeout(() => {
                        document.body.removeChild(marker)
                    }, 30000)
                }
            }
        })
    }, [filteredEvents])
}