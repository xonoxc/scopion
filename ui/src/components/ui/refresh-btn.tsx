import { RefreshCw } from "lucide-react"

export function RefreshButton() {
   return (
      <button className="rounded-lg p-2 text-muted-foreground transition-colors hover:bg-accent hover:text-foreground">
         <RefreshCw className="h-4 w-4" />
      </button>
   )
}
