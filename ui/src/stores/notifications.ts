import { atom, useAtom } from "jotai"

export interface Notification {
   id: string
   type: "error" | "warning" | "info"
   title: string
   message: string
   timestamp: Date
   read: boolean
   traceId?: string
}

export const notificationsAtom = atom<Notification[]>([])

export function useNotifications() {
   return useAtom(notificationsAtom)
}

export function useNotificationActions() {
   const [, setNotifications] = useAtom(notificationsAtom)

   const addNotification = (notification: Omit<Notification, "id" | "timestamp" | "read">) => {
      const newNotification: Notification = {
         ...notification,
         id: crypto.randomUUID(),
         timestamp: new Date(),
         read: false,
      }
      setNotifications(prev => [newNotification, ...prev.slice(0, 49)])
   }

   const markAsRead = (id: string) => {
      setNotifications(prev =>
         prev.map((n: Notification) => (n.id === id ? { ...n, read: true } : n))
      )
   }

   const markAllAsRead = () => {
      setNotifications(prev => prev.map((n: Notification) => ({ ...n, read: true })))
   }

   const clearNotifications = () => {
      setNotifications([])
   }

   return {
      addNotification,
      markAsRead,
      markAllAsRead,
      clearNotifications,
   }
}
