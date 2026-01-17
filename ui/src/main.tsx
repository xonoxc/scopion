import ReactDOM from "react-dom/client"
import { RouterProvider, createRouter } from "@tanstack/react-router"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { routeTree } from "./routeTree.gen"
import "./styles.css"

const queryClient = new QueryClient()

const router = createRouter({
   routeTree,
   defaultPreload: "intent",
   scrollRestoration: true,
})

const rootElement = document.getElementById("app")!

if (!rootElement.innerHTML) {
   const root = ReactDOM.createRoot(rootElement)
   root.render(
      <QueryClientProvider client={queryClient}>
         <RouterProvider router={router} />
      </QueryClientProvider>
   )
}
