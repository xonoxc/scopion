import { useState, useEffect, useRef } from "react"
import { Calendar, ChevronDown, Search, Bell, X, Check, Trash2 } from "lucide-react"
import { useNavigate } from "@tanstack/react-router"
import { cn } from "~/lib/utils"
import { useSystemStatus } from "~/stores/system-status"
import { useSearch } from "~/hooks/use-search"
import { useNotifications, useNotificationActions } from "~/stores/notifications"
import { useErrorNotifications } from "~/hooks/use-error-notifications"
import { useServerStatus } from "~/hooks/use-server-status"
import { RefreshButton } from "~/components/ui/refresh-btn"
import { formatTime } from "~/utils/time"

import type { View } from "~/types/view"
import type { SystemStatus } from "~/stores/system-status"

interface TopBarProps {
     currentView: View
}

const viewTitles: Record<View, string> = {
    overview: "Overview",
    live: "Live Feed",
    errors: "Errors",
    services: "Services",
    traces: "Traces",
 }

export function TopBar({ currentView }: TopBarProps) {
    const [status, setStatus] = useSystemStatus()
    const [eventRate, setEventRate] = useState<number>(42)
    const { status: serverStatus } = useServerStatus()

    // Enable error notifications monitoring
    useErrorNotifications()

    useEffect(() => {
       const interval = setInterval(() => {
          setEventRate(Math.floor(Math.random() * 80) + 10)
          setStatus(Math.random() > 0.1 ? "ingesting" : "idle")
       }, 3000)
       return () => clearInterval(interval)
    }, [])

   // TopBar inspired by OpenSea Pro - Flat, Matte, Functional
   return (
      <header className="sticky top-0 z-30 flex items-center justify-between border-b border-[#262626] bg-[#0b0b0b] px-8 h-16">
         <div className="flex items-center gap-8">
             {/* Simple Title */}
            <h1 className="text-lg font-bold text-white flex items-center gap-3">
               {viewTitles[currentView]}
            </h1>
            
            {/* Status Pill - Minimal */}
             <div className="flex items-center gap-2.5 rounded-full bg-[#1a1a1a] border border-[#262626] px-3 py-1.5">
                <div
                   className={cn(
                      "h-2 w-2 rounded-full",
                      status === "ingesting" ? "bg-emerald-500" : "bg-[#525252]"
                   )}
                />
                <span className="text-[11px] font-medium text-[#a3a3a3] capitalize">{status}</span>
                <div className="h-3 w-px bg-[#333]" />
                <span className="text-[11px] font-mono text-[#e5e5e5]">{eventRate} evt/s</span>
             </div>

             {/* Demo Mode Badge */}
             {serverStatus?.demo_enabled && (
                <div className="flex items-center gap-2 rounded-full bg-orange-500/10 border border-orange-500/20 px-3 py-1.5">
                   <div className="h-2 w-2 rounded-full bg-orange-500" />
                   <span className="text-[11px] font-medium text-orange-400">Demo Mode</span>
                </div>
             )}
         </div>

         <div className="flex items-center gap-4">
             <SearchBar />
             <div className="h-6 w-px bg-[#262626]" />
             <DateRangePicker />
             <NotificationsButton />
            <RefreshButton />
         </div>
      </header>
   )
}

export function StatusIndicator({
   status,
   eventRate,
}: {
   status: SystemStatus
   eventRate: number
}) {
   return (
      <div className="flex items-center gap-2 rounded-full bg-muted/50 px-3 py-1.5">
         <div
            className={cn(
               "h-2 w-2 rounded-full",
               status === "ingesting" ? "bg-success animate-subtle-pulse" : "bg-muted-foreground"
            )}
         />
         <span className="text-xs text-muted-foreground capitalize">{status}</span>
         <span className="text-xs text-muted-foreground">•</span>
         <span className="font-mono text-xs text-muted-foreground">{eventRate} evt/s</span>
      </div>
   )
}

function DateRangePicker() {
   return (
      <button className="flex items-center gap-2 rounded-lg border border-border bg-card px-3 py-2 text-xs font-medium text-foreground transition-colors hover:bg-accent">
         <Calendar className="h-3.5 w-3.5 text-muted-foreground" />
         <span>Last 24 hours</span>
         <ChevronDown className="h-3 w-3 text-muted-foreground" />
      </button>
   )
}

function NotificationsButton() {
     const navigate = useNavigate()
     const [notifications] = useNotifications()
     const { markAsRead, markAllAsRead, clearNotifications } = useNotificationActions()
     const [isOpen, setIsOpen] = useState(false)
     const containerRef = useRef<HTMLDivElement>(null)

    const unreadCount = notifications.filter(n => !n.read).length

    useEffect(() => {
       const handleClickOutside = (e: MouseEvent) => {
          if (containerRef.current && !containerRef.current.contains(e.target as Node)) {
             setIsOpen(false)
          }
       }

       if (isOpen) {
          document.addEventListener("mousedown", handleClickOutside)
       }
       return () => document.removeEventListener("mousedown", handleClickOutside)
    }, [isOpen])

     const handleNotificationClick = (notification: any) => {
        if (notification.traceId) {
           navigate({ to: `/traces/${notification.traceId}` })
        }
        markAsRead(notification.id)
        setIsOpen(false)
     }

    const handleMarkAllRead = () => {
       markAllAsRead()
    }

    const handleClearAll = () => {
       clearNotifications()
       setIsOpen(false)
    }

    return (
       <div ref={containerRef} className="relative">
          <button
             onClick={() => setIsOpen(!isOpen)}
             className="relative rounded-lg p-2 text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
          >
             <Bell className="h-4 w-4" />
             {unreadCount > 0 && (
                <span className="absolute -right-0.5 -top-0.5 flex h-4 w-4 items-center justify-center rounded-full bg-destructive text-[9px] font-bold text-destructive-foreground">
                   {unreadCount > 9 ? "9+" : unreadCount}
                </span>
             )}
          </button>

          {isOpen && (
             <div className="absolute right-0 top-full mt-1 w-80 rounded-lg border border-border bg-card shadow-lg z-50 max-h-96 overflow-auto">
                <div className="flex items-center justify-between px-4 py-3 border-b border-border">
                   <h3 className="text-sm font-semibold text-foreground">Notifications</h3>
                   <div className="flex items-center gap-1">
                      {unreadCount > 0 && (
                         <button
                            onClick={handleMarkAllRead}
                            className="text-xs text-muted-foreground hover:text-foreground flex items-center gap-1"
                         >
                            <Check className="h-3 w-3" />
                            Mark all read
                         </button>
                      )}
                      <button
                         onClick={handleClearAll}
                         className="text-xs text-muted-foreground hover:text-destructive flex items-center gap-1 ml-2"
                      >
                         <Trash2 className="h-3 w-3" />
                         Clear all
                      </button>
                   </div>
                </div>

                {notifications.length === 0 ? (
                   <div className="px-4 py-8 text-center">
                      <div className="mx-auto flex h-8 w-8 items-center justify-center rounded-full bg-muted">
                         <Bell className="h-4 w-4 text-muted-foreground" />
                      </div>
                      <p className="mt-2 text-sm text-muted-foreground">No notifications</p>
                   </div>
                ) : (
                   <div className="divide-y divide-border">
                      {notifications.slice(0, 10).map((notification) => (
                         <button
                            key={notification.id}
                            onClick={() => handleNotificationClick(notification)}
                            className={cn(
                               "w-full px-4 py-3 text-left hover:bg-accent/50 transition-colors",
                               !notification.read && "bg-accent/20"
                            )}
                         >
                            <div className="flex items-start gap-3">
                               <div
                                  className={cn(
                                     "flex h-6 w-6 shrink-0 items-center justify-center rounded-full",
                                     notification.type === "error" && "bg-destructive/15",
                                     notification.type === "warning" && "bg-warning/15",
                                     notification.type === "info" && "bg-muted"
                                  )}
                               >
                                  <div
                                     className={cn(
                                        "h-2 w-2 rounded-full",
                                        notification.type === "error" && "bg-destructive",
                                        notification.type === "warning" && "bg-warning",
                                        notification.type === "info" && "bg-muted-foreground"
                                     )}
                                  />
                               </div>
                               <div className="flex-1 min-w-0">
                                  <p className="text-sm font-medium text-foreground">
                                     {notification.title}
                                  </p>
                                  <p className="text-xs text-muted-foreground mt-0.5">
                                     {notification.message}
                                  </p>
                                  <p className="text-[10px] text-muted-foreground/70 mt-1">
                                     {formatTime(notification.timestamp)}
                                  </p>
                               </div>
                            </div>
                         </button>
                      ))}
                   </div>
                )}
             </div>
          )}
       </div>
    )
}

function SearchBar() {
     const navigate = useNavigate()
     const [query, setQuery] = useState("")
     const [isOpen, setIsOpen] = useState(false)
     const inputRef = useRef<HTMLInputElement>(null)
     const containerRef = useRef<HTMLDivElement>(null)

     const { data: results = [], isLoading } = useSearch(query, isOpen)

    useEffect(() => {
       const handleKeyDown = (e: KeyboardEvent) => {
          if (e.key === "k" && (e.metaKey || e.ctrlKey)) {
             e.preventDefault()
             setIsOpen(true)
             setTimeout(() => inputRef.current?.focus(), 0)
          }
          if (e.key === "Escape" && isOpen) {
             setIsOpen(false)
             setQuery("")
          }
       }

       document.addEventListener("keydown", handleKeyDown)
       return () => document.removeEventListener("keydown", handleKeyDown)
    }, [isOpen])

    useEffect(() => {
       const handleClickOutside = (e: MouseEvent) => {
          if (containerRef.current && !containerRef.current.contains(e.target as Node)) {
             setIsOpen(false)
             setQuery("")
          }
       }

       if (isOpen) {
          document.addEventListener("mousedown", handleClickOutside)
       }
       return () => document.removeEventListener("mousedown", handleClickOutside)
    }, [isOpen])

     const handleSelectTrace = (event: any) => {
        navigate({ to: `/traces/${event.trace_id}` })
        setIsOpen(false)
        setQuery("")
     }

    return (
       <div ref={containerRef} className="relative">
          {!isOpen ? (
             <button
                onClick={() => setIsOpen(true)}
                className="flex items-center gap-2 rounded-lg bg-muted/50 px-3 py-2 text-xs text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
             >
                <Search className="h-3.5 w-3.5" />
                <span>Search traces...</span>
                <kbd className="ml-2 rounded bg-background px-1.5 py-0.5 font-mono text-[10px]">⌘K</kbd>
             </button>
          ) : (
             <div className="relative">
                <div className="flex items-center gap-2 rounded-lg bg-background border border-border px-3 py-2 shadow-lg">
                   <Search className="h-3.5 w-3.5 text-muted-foreground" />
                   <input
                      ref={inputRef}
                      value={query}
                      onChange={(e) => setQuery(e.target.value)}
                      placeholder="Search traces, services..."
                      className="flex-1 bg-transparent text-xs outline-none placeholder:text-muted-foreground"
                   />
                   <button
                      onClick={() => {
                         setIsOpen(false)
                         setQuery("")
                      }}
                      className="text-muted-foreground hover:text-foreground"
                   >
                      <X className="h-3.5 w-3.5" />
                   </button>
                </div>

                {(results.length > 0 || isLoading) && (
                   <div className="absolute top-full mt-1 w-96 rounded-lg border border-border bg-card shadow-lg z-50 max-h-80 overflow-auto">
                      {isLoading && (
                         <div className="px-4 py-3 text-xs text-muted-foreground">
                            Searching...
                         </div>
                      )}
                      {!isLoading && results.length === 0 && query && (
                         <div className="px-4 py-3 text-xs text-muted-foreground">
                            No results found for "{query}"
                         </div>
                      )}
                      {!isLoading && results.map((event) => (
                         <button
                            key={event.id}
                            onClick={() => handleSelectTrace(event)}
                            className="w-full px-4 py-3 text-left hover:bg-accent/50 transition-colors border-b border-border/50 last:border-b-0"
                         >
                            <div className="flex items-center justify-between">
                               <div className="flex-1 min-w-0">
                                  <div className="flex items-center gap-2">
                                     <span className="text-sm font-medium text-foreground truncate">
                                        {event.name}
                                     </span>
                                     <span
                                        className={cn(
                                           "px-1.5 py-0.5 rounded text-[10px] font-semibold uppercase",
                                           event.level === "error" && "bg-destructive/15 text-destructive",
                                           event.level === "warn" && "bg-warning/15 text-warning",
                                           event.level === "info" && "bg-muted text-muted-foreground"
                                        )}
                                     >
                                        {event.level}
                                     </span>
                                  </div>
                                  <div className="flex items-center gap-2 mt-1">
                                     <span className="text-xs text-muted-foreground">{event.service}</span>
                                     <span className="text-xs text-muted-foreground">•</span>
                                     <span className="font-mono text-xs text-muted-foreground">
                                        {event.trace_id.substring(0, 8)}
                                     </span>
                                  </div>
                               </div>
                               <span className="text-xs text-muted-foreground ml-2">
                                  {formatTime(new Date(event.timestamp))}
                               </span>
                            </div>
                         </button>
                      ))}
                   </div>
                )}
             </div>
          )}
       </div>
    )
}
