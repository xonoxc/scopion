import { atom, useAtom } from "jotai"

export type SystemStatus = "ingesting" | "idle"

export const useSystemStatus = () => useAtom(atom<SystemStatus>("ingesting"))
