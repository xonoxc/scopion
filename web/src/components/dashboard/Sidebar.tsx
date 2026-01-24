import React from 'react';

const NavSection = ({ title, children }: { title: string; children: React.ReactNode }) => (
  <div className="mb-8">
    <h3 className="text-[11px] font-bold text-gray-500 uppercase tracking-wider mb-3 pl-3 flex items-center gap-2">
      {/* Icon placeholder if needed, or just text */}
      {title}
    </h3>
    <div className="flex flex-col space-y-0.5">
      {children}
    </div>
  </div>
);

const NavItem = ({ label, active = false, hasSubmenu = false }: { label: string; active?: boolean; hasSubmenu?: boolean }) => (
  <a
    href="#"
    className={`
      group flex items-center justify-between px-3 py-2 rounded-md text-[13px] font-medium transition-all duration-200
      ${active 
        ? 'bg-white/5 text-white shadow-sm ring-1 ring-white/5' 
        : 'text-gray-400 hover:text-white hover:bg-white/5'
      }
    `}
  >
    <span className="flex items-center gap-2">
      {label}
    </span>
    {hasSubmenu && (
      <span className="w-1.5 h-1.5 bg-white rounded-full opacity-0 group-hover:opacity-100 transition-opacity" />
    )}
  </a>
);

export const Sidebar = () => {
  return (
    <aside className="fixed left-0 top-0 bottom-0 w-64 bg-page border-r border-white/5 flex flex-col z-50">
      {/* Logo Area */}
      <div className="h-16 flex items-center px-6 border-b border-white/5 mb-6">
        <div className="flex items-center gap-3">
          <div className="w-8 h-8 bg-linear-to-br from-gray-700 to-gray-900 rounded-lg flex items-center justify-center border border-white/10 shadow-lg">
             <div className="w-3 h-3 bg-white rotate-45" />
          </div>
          <span className="font-bold text-white tracking-tight">Agentic UI</span>
          <span className="ml-auto text-gray-600">
             <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><rect x="3" y="3" width="18" height="18" rx="2" ry="2"></rect><line x1="9" y1="3" x2="9" y2="21"></line></svg>
          </span>
        </div>
      </div>

      {/* Scrollable Nav */}
      <div className="flex-1 overflow-y-auto px-4 py-2 custom-scrollbar">
        
        <NavSection title="Monitor">
            <div className="flex items-center gap-2 text-gray-500 mb-2 pl-3 text-xs uppercase font-bold tracking-widest opacity-60">
                 <svg className="w-3 h-3" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/></svg>
                 MONITOR
            </div>
          <NavItem label="Activity Dashboard" />
          <NavItem label="Call History" />
          <NavItem label="Live Calls" />
        </NavSection>

        <NavSection title="Orchestrate">
             <div className="flex items-center gap-2 text-gray-500 mb-2 pl-3 text-xs uppercase font-bold tracking-widest opacity-60">
                 <svg className="w-3 h-3" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M6 9l6 6 6-6"/></svg>
                 ORCHESTRATE
            </div>
          <NavItem label="../AGENTS ▮" active />
          <NavItem label="Campaigns" />
          <NavItem label="Playbooks" />
        </NavSection>

        <NavSection title="Delegate">
             <div className="flex items-center gap-2 text-gray-500 mb-2 pl-3 text-xs uppercase font-bold tracking-widest opacity-60">
                 <svg className="w-3 h-3" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M15 3h6v6M14 10L21 3M9 21H3v-6M10 14L3 21"/></svg>
                 DELEGATE
            </div>
          <NavItem label="Phone Numbers" />
          <NavItem label="Voice Library" />
          <NavItem label="Integrations" />
          <NavItem label="Events" />
        </NavSection>
      </div>

      {/* Footer Info */}
      <div className="p-6 border-t border-white/5 text-[10px] text-gray-600 leading-relaxed">
        <p className="font-medium text-gray-500">Agentic UI Design System</p>
        <p>Enterprise v. 1.0</p>
        <p className="mt-2">©2026 Agentic UI</p>
      </div>
    </aside>
  );
};
