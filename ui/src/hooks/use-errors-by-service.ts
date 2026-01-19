import { useQuery } from "@tanstack/react-query"

interface ErrorByService {
   service: string
   count: number
}

const mockErrorsByService: ErrorByService[] = [
   {
      service: "web-service",
      count: 2,
   },
   {
      service: "api-service",
      count: 1,
   },
   {
      service: "db-service",
      count: 0,
   },
]

export function useErrorsByService(hours: number = 24) {
   return useQuery({
      queryKey: ["errors-by-service", hours],
      queryFn: async (): Promise<ErrorByService[]> => {
         try {
            const response = await fetch(`/api/errors-by-service?hours=${hours}`)
            if (!response.ok) {
               throw new Error("Failed to fetch errors by service")
            }
            return response.json()
         } catch {
            // Return mock data if fetch fails
            return mockErrorsByService
         }
      },
      refetchInterval: 10000,
      retry: false,
   })
}
