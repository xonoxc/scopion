import { create } from "zustand"
import type { SetterActions } from "~/types/state"
import type { SelectedTrace, View } from "~/types/view"

interface DashboardState {
   currentView: View
   selectedService: string | null
   selectedTrace: SelectedTrace | null
}

type DashboardActions = SetterActions<DashboardState>

export const useDashboardStore = create<DashboardState & DashboardActions>(set => ({
   selectedTrace: null,
   selectedService: null,
   currentView: "overview",

   setCurrentView: currentView =>
      set({
         currentView,
      }),
   setSelectedTrace: selectedTrace => set({ selectedTrace }),
   setSelectedService: selectedService => set({ selectedService }),
}))
