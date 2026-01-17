import { Outlet, createRootRoute, useLocation } from "@tanstack/react-router"
import { Sidebar } from "~/components/sidebar"
import { TopBar } from "~/components/top-bar"

export const Route = createRootRoute({
    component: Root,
})

function Root() {
    const location = useLocation()
    const currentView = location.pathname.split('/')[1] as any || "overview"

    return (
        <div className="flex h-screen bg-black">
            <SidebarContainer />
            <div className="flex flex-1 flex-col overflow-hidden">
                <TopBar currentView={currentView} />
                <main className="flex-1 overflow-auto">
                    <Outlet />
                </main>
            </div>
        </div>
    )
}

function SidebarContainer() {
    return <Sidebar selectedService={null} onClearService={() => {}} />
}
