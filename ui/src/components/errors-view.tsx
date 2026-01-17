import { AlertCircle, ChevronRight } from "lucide-react"
import { useErrorsByService } from "~/hooks/use-errors-by-service"
import { useNavigate } from "@tanstack/react-router"

interface ErrorsViewProps {
     serviceFilter: string | null
}

export function ErrorsView({ serviceFilter }: ErrorsViewProps) {
    const navigate = useNavigate()
    const { data: errorsByService, isLoading } = useErrorsByService()

    if (isLoading) {
        return (
            <div className="flex h-full items-center justify-center">
                <p className="text-sm text-muted-foreground">Loading errors...</p>
            </div>
        )
    }

    const filteredGroups = serviceFilter
       ? errorsByService?.filter(g => g.service === serviceFilter) || []
       : errorsByService || []
    const totalErrors = filteredGroups.reduce((sum, g) => sum + g.count, 0)

   if (filteredGroups.length === 0) {
      return (
         <div className="flex h-full items-center justify-center">
            <div className="text-center">
               <div className="mx-auto flex h-12 w-12 items-center justify-center rounded-full bg-muted">
                  <AlertCircle className="h-6 w-6 text-muted-foreground" />
               </div>
               <p className="mt-4 text-sm font-medium text-foreground">No errors found</p>
               <p className="mt-1 text-xs text-muted-foreground">
                  {serviceFilter
                     ? `No errors for ${serviceFilter}`
                     : "Your services are running smoothly."}
               </p>
            </div>
         </div>
      )
   }

   return (
      <div className="h-full flex flex-col">
         {/* Header with summary */}
         <div className="flex items-center justify-between border-b border-border bg-card px-6 py-4">
            <p className="text-xs text-muted-foreground">
               {filteredGroups.length} error groups • {totalErrors} total errors{" "}
               {serviceFilter && `• filtered by ${serviceFilter}`}
            </p>
         </div>

          {/* Error cards grid */}
          <div className="flex-1 overflow-auto p-6">
             <div className="grid gap-4">
                {filteredGroups.map(group => (
                   <div
                      key={group.service}
                      onClick={() => navigate({ to: "/traces" })}
                      className="group rounded-xl border border-border bg-card p-5 transition-all hover:border-destructive/30 hover:shadow-lg hover:shadow-destructive/5 cursor-pointer"
                   >
                      <div className="flex items-start justify-between gap-6">
                         <div className="flex items-start gap-4">
                            <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-destructive/10">
                               <AlertCircle className="h-5 w-5 text-destructive" />
                            </div>
                            <div>
                               <h3 className="text-sm font-semibold text-foreground">
                                  Errors in {group.service}
                               </h3>
                               <div className="mt-1.5 flex items-center gap-3">
                                  <span className="rounded-md bg-muted px-2 py-0.5 text-xs text-muted-foreground">
                                     {group.service}
                                  </span>
                                  <span className="text-xs text-muted-foreground">
                                     Last 24 hours
                                  </span>
                               </div>
                            </div>
                         </div>

                         <div className="flex items-center gap-6">
                            <div className="text-right">
                               <p className="font-mono text-2xl font-semibold text-destructive">
                                  {group.count}
                               </p>
                               <div className="mt-1 flex items-center justify-end gap-1 text-xs text-muted-foreground">
                                  <span>Total errors</span>
                               </div>
                            </div>

                            <button className="flex h-9 w-9 items-center justify-center rounded-lg bg-secondary text-secondary-foreground transition-colors hover:bg-secondary/80">
                               <ChevronRight className="h-4 w-4" />
                            </button>
                         </div>
                      </div>
                   </div>
                ))}
             </div>
          </div>
      </div>
   )
}
