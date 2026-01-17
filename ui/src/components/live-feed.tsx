import { Copy, Check, Pause, Play } from "lucide-react"
import { useLiveFeed, type Event } from "~/hooks/use-livefeed"
import { formatTime } from "~/utils/time"
import { cn } from "~/lib/utils"
import { useNavigate } from "@tanstack/react-router"

import type { StateTuple } from "~/types/state"

export interface LiveFeedProps {
    serviceFilter: string | null
}

export function LiveFeed({ serviceFilter }: LiveFeedProps) {
    const navigate = useNavigate()
   const { filteredEvents, copiedId, isPaused, setIsPaused, copyTraceId, isLoading } = useLiveFeed({
      serviceFilter,
   })

   if (isLoading && filteredEvents.length === 0) {
      return <LiveFeedLoadingState />
   }

   switch (true) {
      case filteredEvents.length === 0:
         return <LiveFeedEmptyState />
      default:
         return (
            <div className="h-full flex flex-col">
               <div className="flex items-center justify-between border-b border-border bg-card px-6 py-4">
                  <div>
                     <p className="text-xs text-muted-foreground">
                        Real-time event stream {serviceFilter && `• filtered by ${serviceFilter}`}
                     </p>
                  </div>
                  <PauseButton paused={[isPaused, setIsPaused]} />
               </div>

               {/* Table header */}
               <LiveFeedTableHeader />

               {/* Events list */}
          <LiveFeedEventRow
             filteredEvents={filteredEvents}
             navigate={navigate}
             copiedId={copiedId}
             copyTraceId={copyTraceId}
          />
            </div>
         )
   }
}

function LiveFeedLoadingState() {
   return (
      <div className="flex h-full items-center justify-center">
         <div className="text-center">
            <p className="text-sm text-muted-foreground">Loading events...</p>
         </div>
      </div>
   )
}

function LiveFeedEmptyState() {
   return (
      <div className="flex h-full items-center justify-center">
         <div className="text-center">
            <p className="text-sm text-muted-foreground">No events yet.</p>
            <p className="mt-1 text-xs text-muted-foreground/70">
               SCOPION will display activity here as soon as events arrive.
            </p>
         </div>
      </div>
   )
}

function PauseButton({ paused }: { paused: StateTuple<boolean> }) {
   const [isPaused, setIsPaused] = paused

   return (
      <button
         onClick={() => setIsPaused(!isPaused)}
         className={cn(
            "flex items-center gap-2 rounded-lg px-3 py-1.5 text-xs font-medium transition-colors",
            isPaused
               ? "bg-primary/10 text-primary hover:bg-primary/20"
               : "bg-muted text-muted-foreground hover:bg-muted/80"
         )}
      >
         {isPaused ? <Play className="h-3.5 w-3.5" /> : <Pause className="h-3.5 w-3.5" />}
         {isPaused ? "Resume" : "Pause"}
      </button>
   )
}

function LiveFeedEventRow({
    filteredEvents,
    navigate,
    copiedId,
    copyTraceId,
}: {
    filteredEvents: Event[]
    navigate: any
    copiedId: string | null
    copyTraceId: (traceId: string) => void
}) {
   return (
      <div className="flex-1 overflow-auto">
         {filteredEvents.map(event => (
            <div
               key={event.id}
               onClick={() =>
                   navigate({ to: `/traces/${event.trace_id}` })
               }
                className={cn(
                   "group flex cursor-pointer items-center gap-4 border-b border-border px-6 py-3.5 transition-colors hover:bg-accent/50 animate-slide-in",
                   event.level === "error" && "bg-destructive/5 hover:bg-destructive/10"
                )}
            >
                <span className="w-20 shrink-0 font-mono text-xs text-muted-foreground">
                   {formatTime(new Date(event.timestamp))}
                </span>

                <span
                   className={cn(
                      "w-16 shrink-0 rounded-md px-2 py-1 text-center text-[10px] font-semibold uppercase tracking-wide",
                      event.level === "info" && "bg-muted text-muted-foreground",
                      event.level === "warn" && "bg-warning/15 text-warning",
                      event.level === "error" && "bg-destructive/15 text-destructive"
                   )}
                >
                   {event.level}
                </span>

               <span className="w-24 shrink-0 truncate text-xs text-muted-foreground">
                  {event.service}
               </span>

                <span className="flex-1 truncate text-sm font-medium text-foreground">
                   {event.name}
                </span>

                <div className="flex w-28 shrink-0 items-center justify-end gap-2">
                   <span className="font-mono text-xs text-muted-foreground/70">
                      {event.trace_id.substring(0, 8)}
                   </span>
                  <button
                     onClick={e => {
                        e.stopPropagation()
                        copyTraceId(event.trace_id)
                     }}
                     className="rounded-md p-1.5 opacity-0 transition-all hover:bg-muted group-hover:opacity-100"
                  >
                     {copiedId === event.trace_id ? (
                        <Check className="h-3.5 w-3.5 text-success" />
                     ) : (
                        <Copy className="h-3.5 w-3.5 text-muted-foreground" />
                     )}
                  </button>
               </div>
            </div>
         ))}
      </div>
   )
}

function LiveFeedTableHeader() {
   return (
      <div className="flex items-center gap-4 border-b border-border bg-muted/30 px-6 py-2.5">
         <span className="w-20 text-[11px] font-medium uppercase tracking-wider text-muted-foreground">
            Time
         </span>
         <span className="w-16 text-[11px] font-medium uppercase tracking-wider text-muted-foreground">
            Level
         </span>
         <span className="w-24 text-[11px] font-medium uppercase tracking-wider text-muted-foreground">
            Service
         </span>
          <span className="flex-1 text-[11px] font-medium uppercase tracking-wider text-muted-foreground">
             Event
          </span>
          <span className="w-28 text-right text-[11px] font-medium uppercase tracking-wider text-muted-foreground">
             Trace ID
          </span>
      </div>
   )
}
