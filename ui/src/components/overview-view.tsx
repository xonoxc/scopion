import { Activity, AlertCircle, Server, Clock } from "lucide-react"
import { AreaChart, Area, XAxis, YAxis, ResponsiveContainer, Tooltip } from "recharts"
import { useNavigate } from "@tanstack/react-router"
import { cn } from "~/lib/utils"
import { useStats } from "~/hooks/use-stats"
import { useErrorsByService } from "~/hooks/use-errors-by-service"
import { useThroughput } from "~/hooks/use-throughput"

interface OverviewViewProps {}

export function OverviewView({}: OverviewViewProps) {
     const navigate = useNavigate()
    const { data: stats, isLoading: statsLoading } = useStats()
    const { data: errorsByService, isLoading: errorsLoading } = useErrorsByService()
    const { data: throughputData } = useThroughput(24)

   if (statsLoading || errorsLoading) {
      return (
         <div className="flex h-full items-center justify-center">
            <p className="text-sm text-muted-foreground">Loading overview...</p>
         </div>
      )
   }

   const statsData = stats
      ? [
           {
              label: "Total Events",
              value: stats.total_events.toLocaleString(),
              change: "+12.3%", // TODO: calculate real change
              trend: "up" as const,
              icon: Activity,
              color: "text-primary",
           },
           {
              label: "Error Rate",
              value: `${stats.error_rate.toFixed(2)}%`,
              change: "-0.08%", // TODO: calculate real change
              trend: "down" as const,
              icon: AlertCircle,
              color: "text-destructive",
           },
           {
              label: "Active Services",
              value: stats.active_services.toString(),
              change: "All healthy", // TODO: calculate health status
              trend: "neutral" as const,
              icon: Server,
              color: "text-success",
           },
           {
              label: "Avg Latency",
              value: "55ms", // TODO: add latency to stats API
              change: "-8ms",
              trend: "down" as const,
              icon: Clock,
              color: "text-warning",
           },
        ]
      : []

    // Process throughput data to show events per second
    const processedThroughputData = throughputData?.map(item => ({
        time: item.time,
        events: Math.round(item.events / 3600 * 100) / 100, // Convert to events per second with 2 decimal places
    })) || []

    const errorsData =
       errorsByService?.map(e => ({
          service: e.service,
          errors: e.count,
          color: "#e25c5c", // Use consistent color for errors
       })) || []

   return (
      <div className="h-full overflow-auto p-6 custom-scrollbar bg-[#0b0b0b]">
         {/* Stats Grid - OpenSea Pro Style Cards */}
         <div className="grid grid-cols-4 gap-4 mb-6">
            {statsData.map(stat => {
               const Icon = stat.icon
               return (
                  <div
                     key={stat.label}
                     className="group relative overflow-hidden rounded-xl border border-[#262626] bg-[#121212] p-5 transition-all hover:border-[#404040]"
                  >
                     <div className="flex items-center justify-between mb-3">
                        <span className="text-[13px] font-semibold text-[#8a8a8a]">
                           {stat.label}
                        </span>
                        <Icon
                           className={cn(
                              "h-4 w-4 opacity-50 transition-opacity group-hover:opacity-100",
                              stat.color
                           )}
                        />
                     </div>
                     <div className="flex items-end justify-between">
                        <div className="text-2xl font-bold text-white tracking-tight">
                           {stat.value}
                        </div>
                        {stat.trend !== "neutral" && (
                           <div
                              className={cn(
                                 "text-[11px] font-bold flex items-center gap-1",
                                 stat.trend === "up" ? "text-[#34d69b]" : "text-[#f76dc0]"
                              )}
                           >
                              {stat.trend === "up" ? "+" : ""}
                              {stat.change}
                           </div>
                        )}
                     </div>
                  </div>
               )
            })}

            {/* Add a "Best Offer" type card for engagement if needed, or stick to 4 */}
         </div>

         <div className="grid grid-cols-3 gap-6">
            {/* Main Chart Area - "Price History" vibe */}
            <div className="col-span-2 rounded-xl border border-[#262626] bg-[#121212] p-5">
               <div className="flex items-center justify-between mb-6">
                  <div className="flex items-center gap-4">
                     <h3 className="text-sm font-bold text-white flex items-center gap-2">
                        <Activity className="h-4 w-4 text-[#2081e2]" />
                        Event Volume
                     </h3>
                     <div className="flex items-center bg-[#1a1a1a] rounded-lg p-0.5 border border-[#262626]">
                        {["1h", "6h", "24h", "7d"].map((range, i) => (
                           <button
                              key={range}
                              className={cn(
                                 "px-3 py-1 text-[11px] font-semibold rounded-md transition-all",
                                 i === 2
                                    ? "bg-[#262626] text-white shadow-sm"
                                    : "text-[#8a8a8a] hover:text-[#e5e5e5]"
                              )}
                           >
                              {range}
                           </button>
                        ))}
                     </div>
                  </div>
               </div>
                <div className="h-60 w-full">
                   <ResponsiveContainer width="100%" height="100%">
                      <AreaChart data={processedThroughputData}>
                         <defs>
                            <linearGradient id="proGradient" x1="0" y1="0" x2="0" y2="1">
                               <stop offset="0%" stopColor="#2081e2" stopOpacity={0.2} />
                               <stop offset="100%" stopColor="#2081e2" stopOpacity={0} />
                            </linearGradient>
                         </defs>
                         <XAxis
                            dataKey="time"
                            axisLine={false}
                            tickLine={false}
                            tick={{ fill: "#525252", fontSize: 10, fontWeight: 600 }}
                            dy={10}
                         />
                         <YAxis
                            axisLine={false}
                            tickLine={false}
                            tick={{ fill: "#525252", fontSize: 10, fontWeight: 600 }}
                            dx={-10}
                            label={{ value: 'evt/sec', angle: -90, position: 'insideLeft', style: { textAnchor: 'middle', fill: '#525252', fontSize: 10, fontWeight: 600 } }}
                         />
                         <Tooltip
                            contentStyle={{
                               backgroundColor: "#1a1a1a",
                               borderColor: "#262626",
                               borderRadius: "8px",
                               color: "#fff",
                            }}
                            itemStyle={{ color: "#2081e2" }}
                            formatter={(value: number | undefined) => value !== undefined ? [`${value} evt/sec`, 'Events'] : ['', 'Events']}
                         />
                         <Area
                            type="monotone"
                            dataKey="events"
                            stroke="#2081e2"
                            strokeWidth={2}
                            fill="url(#proGradient)"
                         />
                      </AreaChart>
                   </ResponsiveContainer>
                </div>
            </div>

            {/* "Top Collections" Style List for Errors */}
            <div className="rounded-xl border border-[#262626] bg-[#121212] flex flex-col overflow-hidden">
               <div className="p-4 border-b border-[#262626] flex items-center justify-between bg-[#141414]">
                  <h3 className="text-sm font-bold text-white">Top Errors</h3>
                  <button
                     onClick={() => navigate({ to: "/errors" })}
                     className="text-[11px] font-semibold text-[#2081e2] hover:text-[#4aa8f2]"
                  >
                     View All
                  </button>
               </div>

               <div className="flex-1 overflow-auto">
                  <div className="grid grid-cols-[auto_1fr_auto] gap-x-4 px-4 py-2 border-b border-[#262626] text-[10px] font-bold text-[#8a8a8a] uppercase tracking-wider">
                     <span>#</span>
                     <span>Service</span>
                     <span>Count</span>
                  </div>
                  <div className="divide-y divide-[#1a1a1a]">
                     {errorsData.map((error, idx) => (
                        <div
                           key={error.service}
                           onClick={() => navigate({ to: "/errors" })}
                           className="group grid grid-cols-[auto_1fr_auto] items-center gap-x-4 px-4 py-3 hover:bg-[#1a1a1a] cursor-pointer transition-colors"
                        >
                           <span className="text-xs font-medium text-[#525252] w-4">{idx + 1}</span>
                           <div className="flex items-center gap-3 overflow-hidden">
                              <div className="h-8 w-8 rounded-lg bg-[#262626] flex items-center justify-center shrink-0">
                                 <Server className="h-4 w-4 text-[#e5e5e5]" />
                              </div>
                              <div className="flex flex-col min-w-0">
                                 <span className="text-sm font-bold text-white truncate group-hover:text-[#2081e2] transition-colors">
                                    {error.service}
                                 </span>
                                 <div className="flex items-center gap-1.5">
                                    <span className="text-[10px] text-[#8a8a8a] truncate">
                                       Backend
                                    </span>
                                    {idx < 2 && (
                                       <span className="text-[10px] text-[#34d69b] font-bold">
                                          +12%
                                       </span>
                                    )}
                                 </div>
                              </div>
                           </div>
                           <div className="flex flex-col items-end">
                              <span className="text-sm font-bold text-white">{error.errors}</span>
                              <span className="text-[10px] font-medium text-[#f76dc0]">
                                 High vol
                              </span>
                           </div>
                        </div>
                     ))}
                  </div>
               </div>
            </div>
         </div>
      </div>
   )
}
