import React from "react";
import { Sidebar } from "./Sidebar";
import { Header } from "./Header";

interface DashboardLayoutProps {
  children: React.ReactNode;
}

export const DashboardLayout: React.FC<DashboardLayoutProps> = ({
  children,
}) => {
  return (
    <div className="min-h-screen bg-page text-white font-sans flex antialiased selection:bg-primary/30">
      {/* Sidebar */}
      <Sidebar />

      {/* Main Content Area */}
      <div className="flex-1 flex flex-col ml-64 transition-all duration-300">
        <Header />

        <main className="flex-1 p-8 pt-6">
          <div className="max-w-400 mx-auto animate-fade-in-up">{children}</div>
        </main>
      </div>
    </div>
  );
};
