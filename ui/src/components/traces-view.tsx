import { GitBranch, Filter } from "lucide-react"
import { cn } from "~/lib/utils"
import { useTraces } from "~/hooks/use-traces"
import { useNavigate } from "@tanstack/react-router"

interface TracesViewProps {
   serviceFilter: string | null
}

export function TracesView({ serviceFilter }: TracesViewProps) {
   const navigate = useNavigate()
   const { data: traces, isLoading } = useTraces()

   if (isLoading) {
      return (
         <div className="flex h-full items-center justify-center">
            <p className="text-sm text-muted-foreground">Loading traces...</p>
         </div>
      )
   }

   const formatTime = (dateStr: string) => {
      const date = new Date(dateStr)
      return date.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit", second: "2-digit" })
   }

   const tracesData =
      traces?.map(t => ({
         id: t.id,
         name: t.name,
         service: t.service,
         duration: t.duration,
         timestamp: new Date(t.timestamp),
         hasError: t.has_error,
         spanCount: t.spans,
      })) || []

   const filteredTraces = serviceFilter
      ? tracesData.filter(t => t.service === serviceFilter)
      : tracesData
   const maxDuration =
      filteredTraces.length > 0 ? Math.max(...filteredTraces.map(t => t.duration)) : 1000

   return (
      <div className="h-full flex flex-col">
         {/* Header */}
         <div className="flex items-center justify-between border-b border-border bg-card px-6 py-4">
            <p className="text-xs text-muted-foreground">
               {filteredTraces.length} traces {serviceFilter && `• filtered by ${serviceFilter}`}
            </p>
            <button className="flex items-center gap-2 rounded-lg border border-border bg-card px-3 py-1.5 text-xs text-muted-foreground transition-colors hover:bg-accent hover:text-foreground">
               <Filter className="h-3.5 w-3.5" />
               Filters
            </button>
         </div>

         {/* Table header */}
         <div className="flex items-center gap-4 border-b border-border bg-muted/30 px-6 py-2.5">
            <span className="w-8" />
            <span className="flex-1 text-[11px] font-medium uppercase tracking-wider text-muted-foreground">
               Trace
            </span>
            <span className="w-24 text-[11px] font-medium uppercase tracking-wider text-muted-foreground">
               Service
            </span>
            <span className="w-20 text-[11px] font-medium uppercase tracking-wider text-muted-foreground">
               Spans
            </span>
            <span className="w-48 text-[11px] font-medium uppercase tracking-wider text-muted-foreground">
               Duration
            </span>
            <span className="w-24 text-right text-[11px] font-medium uppercase tracking-wider text-muted-foreground">
               Time
            </span>
         </div>

         {/* Traces list */}
         <div className="flex-1 overflow-auto">
            {filteredTraces.map(trace => {
               const durationPercent = (trace.duration / maxDuration) * 100
               return (
                  <div
                     key={trace.id}
                     onClick={() => navigate({ to: `/traces/${trace.id}` })}
                     className={cn(
                        "group flex items-center gap-4 border-b border-border px-6 py-4 transition-colors hover:bg-accent/50 cursor-pointer",
                        trace.hasError && "bg-destructive/5 hover:bg-destructive/10"
                     )}
                  >
                     <GitBranch
                        className={cn(
                           "h-4 w-4 shrink-0",
                           trace.hasError ? "text-destructive" : "text-muted-foreground"
                        )}
                     />

                     <div className="flex-1 min-w-0">
                        <div className="flex items-center gap-2">
                           <span className="text-sm font-medium text-foreground truncate">
                              {trace.name}
                           </span>
                           {trace.hasError && (
                              <span className="shrink-0 rounded-md bg-destructive/15 px-2 py-0.5 text-[10px] font-semibold text-destructive">
                                 ERROR
                              </span>
                           )}
                        </div>
                     </div>

                     <span className="w-24 truncate text-xs text-muted-foreground">
                        {trace.service}
                     </span>

                     <span className="w-20 text-xs text-muted-foreground">
                        {trace.spanCount} spans
                     </span>

                     {/* Duration bar */}
                     <div className="w-48 flex items-center gap-3">
                        <div className="flex-1 h-2 rounded-full bg-muted overflow-hidden">
                           <div
                              className={cn(
                                 "h-full rounded-full transition-all",
                                 trace.hasError
                                    ? "bg-destructive"
                                    : trace.duration > 1000
                                      ? "bg-warning"
                                      : "bg-primary"
                              )}
                              style={{ width: `${durationPercent}%` }}
                           />
                        </div>
                        <span
                           className={cn(
                              "font-mono text-xs w-16 text-right",
                              trace.duration > 1000 ? "text-warning" : "text-muted-foreground"
                           )}
                        >
                           {trace.duration}ms
                        </span>
                     </div>

                     <span className="w-24 text-right text-xs text-muted-foreground">
                        {formatTime(trace.timestamp.toISOString())}
                     </span>
                  </div>
               )
            })}
         </div>
      </div>
   )
}
