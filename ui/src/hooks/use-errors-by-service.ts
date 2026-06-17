import { useQuery } from "@tanstack/react-query"
import { attempt } from "~/lib/attempt"

export interface ErrorByService {
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
         const result = await attempt(
            fetch(`/api/errors-by-service?hours=${hours}`).then(async res => {
               if (!res.ok) {
                  throw new Error("Failed to fetch errors by service")
               }
               return res.json()
            })
         )

         return result.match(
            data => data,
            e => {
               console.warn("useErrorsByService: API fetch failed, using mock data:", e)
               return mockErrorsByService
            }
         )
      },
      refetchInterval: 10000,
      retry: false,
   })
}
