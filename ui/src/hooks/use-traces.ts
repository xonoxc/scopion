import { useQuery } from "@tanstack/react-query"

interface TraceInfo {
   id: string
   name: string
   service: string
   duration: number
   spans: number
   timestamp: string
   has_error: boolean
}

export function useTraces(limit: number = 50) {
   return useQuery({
      queryKey: ["traces", limit],
      queryFn: async (): Promise<TraceInfo[]> => {
         const response = await fetch(`/api/traces?limit=${limit}`)
         if (!response.ok) {
            throw new Error("Failed to fetch traces")
         }
         return response.json()
      },
      refetchInterval: 10000,
   })
}

