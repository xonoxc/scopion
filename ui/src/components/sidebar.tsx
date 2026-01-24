import type React from "react"
import { LayoutDashboard, Activity, AlertCircle, Server, GitBranch, X } from "lucide-react"
import { Link, useLocation } from "@tanstack/react-router"

import { cn } from "~/lib/utils"
import type { View } from "~/types/view"

interface SidebarProps {
   selectedService: string | null
   onClearService: () => void
}

const navItems: { id: View; label: string; icon: React.ElementType }[] = [
   { id: "overview", label: "Overview", icon: LayoutDashboard },
   { id: "live", label: "Live Feed", icon: Activity },
   { id: "errors", label: "Errors", icon: AlertCircle },
   { id: "services", label: "Services", icon: Server },
   { id: "traces", label: "Traces", icon: GitBranch },
]

export function Sidebar({ selectedService, onClearService }: SidebarProps) {
   const location = useLocation()
   return (
      <aside className="hidden md:flex w-65 flex-col border-r border-border bg-background transition-all duration-300">
         {/* Logo - Minimalist & Pro */}
          <div className="flex h-16 items-center gap-3 px-5 border-b border-border">
            <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary/20">
               <span className="text-xl font-bold text-primary leading-none">S</span>
            </div>
             <div className="flex flex-col">
                <span className="text-sm font-bold tracking-normal text-foreground">SCOPION</span>
                <span className="text-[10px] uppercase font-bold text-muted-foreground tracking-wider">
                   Telemetry Pro
                </span>
             </div>
         </div>

         {/* Service filter badge */}
         {selectedService && (
            <div className="mx-3 mt-4 mb-1">
                <div className="group flex items-center justify-between rounded-md border border-primary/30 bg-primary/10 px-3 py-2 transition-colors">
                   <div className="flex items-center gap-2.5 overflow-hidden">
                      <Server className="h-3.5 w-3.5 text-primary shrink-0" />
                      <span className="text-xs font-semibold text-foreground truncate">
                         {selectedService}
                      </span>
                   </div>
                  <button
                     onClick={onClearService}
                     className="rounded hover:bg-primary/20 p-0.5 text-primary/70 hover:text-primary transition-colors"
                  >
                     <X className="h-3 w-3" />
                  </button>
               </div>
            </div>
         )}

         {/* Navigation - Clean List */}
         <nav className="flex-1 px-3 py-4 overflow-y-auto custom-scrollbar">
            <ul className="space-y-0.5">
               {navItems.map(item => {
                  const Icon = item.icon
                  const isActive = location.pathname === `/${item.id}`
                  return (
                     <li key={item.id}>
                        <Link
                           to={`/${item.id}`}
                            className={cn(
                               "group flex w-full items-center gap-3.5 rounded-lg px-3 py-2.5 text-[13px] font-semibold transition-all duration-150",
                               isActive
                                  ? "bg-sidebar-accent text-foreground"
                                  : "text-muted-foreground hover:bg-sidebar-accent hover:text-foreground"
                            )}
                        >
                            <Icon
                               className={cn(
                                  "h-4 w-4 transition-colors",
                                  isActive
                                     ? "text-primary"
                                     : "text-muted-foreground group-hover:text-primary"
                               )}
                            />
                           <span>{item.label}</span>
                            {/* Error count badge for errors view */}
                            {item.id === "errors" && (
                               <span className="ml-auto flex h-5 min-w-5 items-center justify-center rounded bg-destructive/15 px-1.5 text-[10px] font-bold text-destructive">
                                  4
                               </span>
                            )}
                        </Link>
                     </li>
                  )
               })}
            </ul>
         </nav>

         {/* Footer - Minimal */}
          <div className="mt-auto border-t border-border px-5 py-5">
             <div className="flex items-center justify-between text-[11px] text-muted-foreground font-medium">
               <span>v1.0.0</span>
               <div className="flex items-center gap-2">
                  <div className="h-1.5 w-1.5 rounded-full bg-emerald-500" />
                  <span>Stable</span>
               </div>
            </div>
         </div>
      </aside>
   )
}
