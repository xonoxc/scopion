import { useQuery } from "@tanstack/react-query"

interface ServiceInfo {
   name: string
   error_count: number
   last_activity: string
   event_count: number
}

const mockServices: ServiceInfo[] = [
   {
      name: "web-service",
      error_count: 2,
      last_activity: new Date(Date.now() - 5000).toISOString(),
      event_count: 1200,
   },
   {
      name: "db-service",
      error_count: 0,
      last_activity: new Date(Date.now() - 10000).toISOString(),
      event_count: 800,
   },
   {
      name: "api-service",
      error_count: 1,
      last_activity: new Date(Date.now() - 15000).toISOString(),
      event_count: 600,
   },
]

export function useServices() {
   return useQuery({
      queryKey: ["services"],
      queryFn: async (): Promise<ServiceInfo[]> => {
         try {
            const response = await fetch("/api/services")
            if (!response.ok) {
               throw new Error("Failed to fetch services")
            }
            return response.json()
         } catch {
            // Return mock data if fetch fails
            return mockServices
         }
      },
      refetchInterval: 10000,
      retry: false,
   })
}
