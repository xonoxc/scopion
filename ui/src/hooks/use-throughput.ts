import { useQuery } from "@tanstack/react-query"

export interface ThroughputData {
  time: string
  events: number
}

const mockThroughput: ThroughputData[] = [
  { time: "00:00", events: 120 },
  { time: "04:00", events: 85 },
  { time: "08:00", events: 210 },
  { time: "12:00", events: 380 },
  { time: "16:00", events: 420 },
  { time: "20:00", events: 290 },
  { time: "Now", events: 340 },
]

export function useThroughput(hours: number = 24) {
  return useQuery({
    queryKey: ["throughput", hours],
    queryFn: async (): Promise<ThroughputData[]> => {
      try {
        const response = await fetch(`/api/throughput?hours=${hours}`)
        if (!response.ok) {
          throw new Error("Failed to fetch throughput")
        }
        return response.json()
      } catch {
        // Return mock data if fetch fails
        return mockThroughput
      }
    },
    refetchInterval: 30000, // Refetch every 30 seconds
    retry: false,
  })
}