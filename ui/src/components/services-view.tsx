import { Server, ChevronRight, Activity, AlertCircle, Clock } from "lucide-react"
import { AreaChart, Area, ResponsiveContainer } from "recharts"
import { cn } from "~/lib/utils"
import { useServices } from "~/hooks/use-services"

interface ServicesViewProps {}



export function ServicesView({}: ServicesViewProps) {
    const { data: services, isLoading } = useServices()

    if (isLoading) {
        return (
            <div className="flex h-full items-center justify-center">
                <p className="text-sm text-muted-foreground">Loading services...</p>
            </div>
        )
    }

    const formatTime = (dateStr: string) => {
       const date = new Date(dateStr)
       const now = new Date()
       const diff = now.getTime() - date.getTime()
       if (diff < 60000) return "active"
       if (diff < 3600000) return `${Math.floor(diff / 60000)}m ago`
       return `${Math.floor(diff / 3600000)}h ago`
    }

    const servicesData = services?.map(s => ({
        ...s,
        status: s.error_count > 0 ? "errors" : "healthy" as "healthy" | "errors" | "idle",
        eventsPerMinute: Math.round(s.event_count / 60), // Rough estimate
        avgLatency: 50, // Placeholder
        sparkline: [
            { value: 10 },
            { value: 15 },
            { value: 12 },
            { value: 18 },
            { value: 14 },
            { value: 16 },
            { value: 13 },
        ], // Placeholder sparkline
    })) || []

    const healthyCount = servicesData.filter(s => s.status === "healthy").length
    const errorCount = servicesData.filter(s => s.status === "errors").length

   return (
      <div className="h-full flex flex-col">
         {/* Header with summary */}
         <div className="flex items-center justify-between border-b border-border bg-card px-6 py-4">
             <p className="text-xs text-muted-foreground">
                {servicesData.length} services • {healthyCount} healthy • {errorCount} with errors
             </p>
         </div>

         {/* Services grid */}
         <div className="flex-1 overflow-auto p-6">
             <div className="grid grid-cols-2 gap-4">
                {servicesData.map(service => (
                  <div
                     key={service.name}
                     onClick={() => {}} // TODO: navigate to /live with service filter
                     className="group rounded-xl border border-border bg-card p-5 transition-all hover:border-primary/30 hover:shadow-lg cursor-pointer"
                  >
                     <div className="flex items-start justify-between">
                        <div className="flex items-center gap-3">
                           <div
                              className={cn(
                                 "flex h-10 w-10 items-center justify-center rounded-lg",
                                 service.status === "healthy" && "bg-success/10",
                                 service.status === "errors" && "bg-destructive/10",
                                 service.status === "idle" && "bg-muted"
                              )}
                           >
                              <Server
                                 className={cn(
                                    "h-5 w-5",
                                    service.status === "healthy" && "text-success",
                                    service.status === "errors" && "text-destructive",
                                    service.status === "idle" && "text-muted-foreground"
                                 )}
                              />
                           </div>
                           <div>
                              <h3 className="font-mono text-sm font-semibold text-foreground">
                                 {service.name}
                              </h3>
                              <div className="mt-1 flex items-center gap-1.5">
                                 <div
                                    className={cn(
                                       "h-1.5 w-1.5 rounded-full",
                                       service.status === "healthy" && "bg-success",
                                       service.status === "errors" && "bg-destructive",
                                       service.status === "idle" && "bg-muted-foreground"
                                    )}
                                 />
                                 <span
                                    className={cn(
                                       "text-xs capitalize",
                                       service.status === "healthy" && "text-success",
                                       service.status === "errors" && "text-destructive",
                                       service.status === "idle" && "text-muted-foreground"
                                    )}
                                 >
                                    {service.status}
                                 </span>
                              </div>
                           </div>
                        </div>

                        <div className="h-10 w-20">
                           <ResponsiveContainer width="100%" height="100%">
                              <AreaChart data={service.sparkline}>
                                 <defs>
                                    <linearGradient
                                       id={`spark-${service.name}`}
                                       x1="0"
                                       y1="0"
                                       x2="0"
                                       y2="1"
                                    >
                                       <stop
                                          offset="0%"
                                          stopColor={
                                             service.status === "errors" ? "#e25c5c" : "#3b9e8c"
                                          }
                                          stopOpacity={0.3}
                                       />
                                       <stop
                                          offset="100%"
                                          stopColor={
                                             service.status === "errors" ? "#e25c5c" : "#3b9e8c"
                                          }
                                          stopOpacity={0}
                                       />
                                    </linearGradient>
                                 </defs>
                                 <Area
                                    type="monotone"
                                    dataKey="value"
                                    stroke={service.status === "errors" ? "#e25c5c" : "#3b9e8c"}
                                    strokeWidth={1.5}
                                    fill={`url(#spark-${service.name})`}
                                 />
                              </AreaChart>
                           </ResponsiveContainer>
                        </div>
                     </div>

                     <div className="mt-5 grid grid-cols-3 gap-4">
                        <div>
                           <div className="flex items-center gap-1.5 text-muted-foreground">
                              <Activity className="h-3.5 w-3.5" />
                              <span className="text-[10px] uppercase tracking-wide">
                                 Events/min
                              </span>
                           </div>
                           <p className="mt-1 font-mono text-lg font-semibold text-foreground">
                              {service.eventsPerMinute}
                           </p>
                        </div>
                        <div>
                           <div className="flex items-center gap-1.5 text-muted-foreground">
                              <Clock className="h-3.5 w-3.5" />
                              <span className="text-[10px] uppercase tracking-wide">Latency</span>
                           </div>
                           <p
                              className={cn(
                                 "mt-1 font-mono text-lg font-semibold",
                                 service.avgLatency > 500 ? "text-warning" : "text-foreground"
                              )}
                           >
                              {service.avgLatency}ms
                           </p>
                        </div>
                        <div>
                           <div className="flex items-center gap-1.5 text-muted-foreground">
                              <AlertCircle className="h-3.5 w-3.5" />
                              <span className="text-[10px] uppercase tracking-wide">Errors</span>
                           </div>
                            <p
                               className={cn(
                                  "mt-1 font-mono text-lg font-semibold",
                                  service.error_count > 0 ? "text-destructive" : "text-foreground"
                               )}
                            >
                               {service.error_count}
                            </p>
                        </div>
                     </div>

                     <div className="mt-4 flex items-center justify-between border-t border-border pt-4">
                         <span className="text-xs text-muted-foreground">
                            {formatTime(service.last_activity)}
                         </span>
                        <ChevronRight className="h-4 w-4 text-muted-foreground transition-transform group-hover:translate-x-1" />
                     </div>
                  </div>
               ))}
            </div>
         </div>
      </div>
   )
}
