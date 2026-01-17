export type View = "live" | "errors" | "services" | "traces" | "overview"

export interface SelectedTrace {
   id: string
   name: string
   service: string
}
