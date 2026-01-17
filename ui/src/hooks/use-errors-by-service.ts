import { useQuery } from "@tanstack/react-query"

interface ErrorByService {
   service: string
   count: number
}

export function useErrorsByService(hours: number = 24) {
   return useQuery({
      queryKey: ["errors-by-service", hours],
      queryFn: async (): Promise<ErrorByService[]> => {
         const response = await fetch(`/api/errors-by-service?hours=${hours}`)
         if (!response.ok) {
            throw new Error("Failed to fetch errors by service")
         }
         return response.json()
      },
      refetchInterval: 10000,
   })
}

