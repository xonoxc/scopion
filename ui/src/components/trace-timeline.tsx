import { useState } from "react"
import {
   ArrowLeft,
   Copy,
   Check,
   ChevronDown,
   ChevronRight,
   AlertCircle,
   Clock,
   Hash,
   Layers,
} from "lucide-react"
import type { SelectedTrace } from "~/types/view"
import { cn } from "~/lib/utils"

interface Span {
   id: string
   name: string
   duration: number
   startOffset: number
   hasError: boolean
   errorMessage?: string
   children: Span[]
   metadata?: Record<string, string>
}

interface TraceEvent {
   id: string
   timestamp: string
   level: string
   service: string
   name: string
   trace_id?: string
   data?: Record<string, any>
}

interface TraceTimelineProps {
   trace: SelectedTrace
   events: TraceEvent[]
   onClose: () => void
}

// Convert events to spans
function eventsToSpans(events: TraceEvent[]): Span[] {
   if (events.length === 0) return []

   const startTime = new Date(events[0].timestamp).getTime()
   const endTime = new Date(events[events.length - 1].timestamp).getTime()

   return events.map((event, index) => {
      const eventTime = new Date(event.timestamp).getTime()
      const startOffset = eventTime - startTime

      // For simplicity, assume each event is a span with duration based on time to next event
      const nextEventTime =
         index < events.length - 1 ? new Date(events[index + 1].timestamp).getTime() : endTime
      const duration = Math.max(nextEventTime - eventTime, 1) // Minimum 1ms

      return {
         id: event.id,
         name: event.name,
         duration: duration,
         startOffset: startOffset,
         hasError: event.level === "error",
         errorMessage: event.level === "error" ? event.name : undefined,
         metadata: event.data || {},
         children: [],
      }
   })
}

export function TraceTimeline({ trace, events, onClose }: TraceTimelineProps) {
   const [copiedId, setCopiedId] = useState(false)
   const [expandedSpans, setExpandedSpans] = useState<Set<string>>(new Set())

   const spans = eventsToSpans(events)
   const totalDuration =
      spans.length > 0 ? Math.max(...spans.map(s => s.startOffset + s.duration)) : 0
   const totalSpans = spans.length

   const copyTraceId = () => {
      navigator.clipboard.writeText(trace.id)
      setCopiedId(true)
      setTimeout(() => setCopiedId(false), 2000)
   }

   const toggleSpan = (spanId: string) => {
      setExpandedSpans(prev => {
         const next = new Set(prev)
         if (next.has(spanId)) {
            next.delete(spanId)
         } else {
            next.add(spanId)
         }
         return next
      })
   }

   const renderSpan = (span: Span, depth = 0) => {
      const hasChildren = span.children.length > 0
      const isExpanded = expandedSpans.has(span.id)
      const widthPercent = (span.duration / totalDuration) * 100
      const leftPercent = (span.startOffset / totalDuration) * 100

      return (
         <div key={span.id} className="animate-fade-in">
            <div
               className={cn(
                  "group flex items-center gap-3 border-b border-border py-3 pr-6 transition-colors hover:bg-accent/30",
                  span.hasError && "bg-destructive/5"
               )}
               style={{ paddingLeft: `${24 + depth * 24}px` }}
            >
               {/* Expand toggle */}
               <button
                  onClick={() => toggleSpan(span.id)}
                  className={cn(
                     "flex h-6 w-6 shrink-0 items-center justify-center rounded-md transition-colors",
                     hasChildren ? "hover:bg-muted" : "invisible"
                  )}
               >
                  {hasChildren &&
                     (isExpanded ? (
                        <ChevronDown className="h-4 w-4 text-muted-foreground" />
                     ) : (
                        <ChevronRight className="h-4 w-4 text-muted-foreground" />
                     ))}
               </button>

               {/* Span name */}
               <div className="flex min-w-45 items-center gap-2">
                  <span
                     className={cn(
                        "text-sm font-medium",
                        span.hasError ? "text-destructive" : "text-foreground"
                     )}
                  >
                     {span.name}
                  </span>
                  {span.hasError && <AlertCircle className="h-4 w-4 text-destructive" />}
               </div>

               {/* Duration bar with scale markers */}
               <div className="relative flex-1 h-8">
                  {/* Background grid lines */}
                  <div className="absolute inset-0 flex">
                     {[0, 25, 50, 75, 100].map(percent => (
                        <div
                           key={percent}
                           className="flex-1 border-l border-border/30 first:border-l-0"
                        />
                     ))}
                  </div>
                  {/* Duration bar */}
                  <div className="absolute inset-y-0 left-0 right-0 flex items-center">
                     <div
                        className={cn(
                           "h-5 rounded-md transition-all",
                           span.hasError ? "bg-destructive/60" : "bg-primary/60"
                        )}
                        style={{
                           width: `${Math.max(widthPercent, 3)}%`,
                           marginLeft: `${leftPercent}%`,
                        }}
                     />
                  </div>
               </div>

               {/* Duration */}
               <span
                  className={cn(
                     "shrink-0 font-mono text-sm font-medium w-20 text-right",
                     span.hasError
                        ? "text-destructive"
                        : span.duration > 50
                          ? "text-warning"
                          : "text-muted-foreground"
                  )}
               >
                  {span.duration}ms
               </span>
            </div>

            {/* Error message */}
            {span.hasError && span.errorMessage && isExpanded && (
               <div
                  className="border-b border-border bg-destructive/5 px-6 py-3"
                  style={{ paddingLeft: `${54 + depth * 24}px` }}
               >
                  <p className="font-mono text-xs text-destructive">{span.errorMessage}</p>
               </div>
            )}

            {/* Metadata */}
            {span.metadata && isExpanded && (
               <div
                  className="border-b border-border bg-muted/20 px-6 py-3"
                  style={{ paddingLeft: `${54 + depth * 24}px` }}
               >
                  <div className="flex flex-wrap gap-x-6 gap-y-2">
                     {Object.entries(span.metadata).map(([key, value]) => (
                        <div key={key} className="flex items-center gap-2">
                           <span className="text-[11px] text-muted-foreground">{key}</span>
                           <span className="font-mono text-[11px] text-foreground bg-muted px-1.5 py-0.5 rounded">
                              {value}
                           </span>
                        </div>
                     ))}
                  </div>
               </div>
            )}

            {/* Children */}
            {hasChildren && isExpanded && span.children.map(child => renderSpan(child, depth + 1))}
         </div>
      )
   }

   return (
      <div className="h-full flex flex-col">
         {/* Header */}
         <div className="border-b border-border bg-card px-6 py-4">
            <div className="flex items-center justify-between">
               <div className="flex items-center gap-4">
                  <button
                     onClick={onClose}
                     className="flex items-center gap-2 rounded-lg px-3 py-1.5 text-sm text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
                  >
                     <ArrowLeft className="h-4 w-4" />
                     Back
                  </button>
                  <div className="h-6 w-px bg-border" />
                  <div>
                     <h1 className="text-sm font-semibold text-foreground">{trace.name}</h1>
                     <p className="text-xs text-muted-foreground">{trace.service}</p>
                  </div>
               </div>
               <button
                  onClick={copyTraceId}
                  className="flex items-center gap-2 rounded-lg bg-secondary px-4 py-2 text-xs font-medium text-secondary-foreground transition-colors hover:bg-secondary/80"
               >
                  {copiedId ? (
                     <>
                        <Check className="h-4 w-4" />
                        Copied
                     </>
                  ) : (
                     <>
                        <Copy className="h-4 w-4" />
                        Copy trace ID
                     </>
                  )}
               </button>
            </div>
         </div>

         {/* Summary stats */}
         <div className="flex items-center gap-8 border-b border-border bg-card px-6 py-3">
            <div className="flex items-center gap-2">
               <Clock className="h-4 w-4 text-muted-foreground" />
               <span className="text-xs text-muted-foreground">Duration</span>
               <span className="font-mono text-sm font-semibold text-foreground">
                  {totalDuration}ms
               </span>
            </div>
            <div className="flex items-center gap-2">
               <Layers className="h-4 w-4 text-muted-foreground" />
               <span className="text-xs text-muted-foreground">Spans</span>
               <span className="font-mono text-sm font-semibold text-foreground">{totalSpans}</span>
            </div>
            <div className="flex items-center gap-2">
               <Hash className="h-4 w-4 text-muted-foreground" />
               <span className="text-xs text-muted-foreground">Trace ID</span>
               <span className="font-mono text-sm text-muted-foreground">{trace.id}</span>
            </div>
         </div>

         {/* Timeline header with scale */}
         <div className="flex items-center gap-3 border-b border-border bg-muted/30 px-6 py-2.5">
            <span className="w-6" />
            <span className="min-w-45 text-[11px] font-medium uppercase tracking-wider text-muted-foreground">
               Span
            </span>
            <div className="flex-1 flex justify-between text-[10px] text-muted-foreground font-mono">
               <span>0ms</span>
               <span>{Math.round(totalDuration / 4)}ms</span>
               <span>{Math.round(totalDuration / 2)}ms</span>
               <span>{Math.round((totalDuration * 3) / 4)}ms</span>
               <span>{totalDuration}ms</span>
            </div>
            <span className="w-20 text-right text-[11px] font-medium uppercase tracking-wider text-muted-foreground">
               Duration
            </span>
         </div>

         {/* Spans */}
         <div className="flex-1 overflow-auto">{spans.map(span => renderSpan(span))}</div>
      </div>
   )
}
