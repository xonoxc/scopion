import { useQuery } from "@tanstack/react-query"

interface ServiceInfo {
   name: string
   error_count: number
   last_activity: string
   event_count: number
}

export function useServices() {
   return useQuery({
      queryKey: ["services"],
      queryFn: async (): Promise<ServiceInfo[]> => {
         const response = await fetch("/api/services")
         if (!response.ok) {
            throw new Error("Failed to fetch services")
         }
         return response.json()
      },
      refetchInterval: 10000,
   })
}

