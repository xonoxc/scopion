import React from 'react';

const ActionCard = ({ icon, label }: { icon: React.ReactNode; label: string }) => (
  <button className="group flex flex-col items-start justify-between p-5 bg-card border border-white/5 rounded-xl hover:border-primary/50 transition-all duration-300 h-32 w-full text-left relative overflow-hidden">
    <div className="text-gray-400 group-hover:text-primary transition-colors">
      {icon}
    </div>
    <div className="flex items-center justify-between w-full mt-auto">
        <span className="text-[10px] font-bold text-gray-400 uppercase tracking-widest group-hover:text-white transition-colors">{label}</span>
        <span className="opacity-0 group-hover:opacity-100 transform translate-x-2 group-hover:translate-x-0 transition-all duration-300 text-primary">
            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><line x1="7" y1="17" x2="17" y2="7"></line><polyline points="7 7 17 7 17 17"></polyline></svg>
        </span>
    </div>
    
    {/* Hover Effect */}
    <div className="absolute inset-0 bg-primary/5 opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none" />
  </button>
);

export const AgentHeader = () => {
  return (
    <div className="flex flex-col xl:flex-row gap-8 mb-12 animate-fade-in-up">
      {/* Left: Agent Details */}
      <div className="flex-1">
        <h1 className="text-3xl font-medium text-white mb-6 tracking-tight">Equity Research Agent</h1>
        
        <div className="space-y-3">
          <div className="flex items-center gap-4 text-sm">
            <span className="text-gray-500 w-20">Created:</span>
            <span className="text-gray-300 font-mono">Sep 6, 2025</span>
          </div>
          <div className="flex items-center gap-4 text-sm">
            <span className="text-gray-500 w-20">Variants:</span>
            <span className="text-gray-300 font-mono">10</span>
          </div>
          <div className="flex items-center gap-4 text-sm">
            <span className="text-gray-500 w-20">Accuracy:</span>
            <span className="text-emerald-400 font-mono font-bold">81%</span>
          </div>
        </div>
      </div>

      {/* Right: Actions Grid */}
      <div className="flex-[1.5] grid grid-cols-2 md:grid-cols-4 gap-4">
        <ActionCard 
            label="Configure Agent" 
            icon={<svg className="w-6 h-6" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5"><path d="M12.22 2h-.44a2 2 0 0 0-2 2v.18a2 2 0 0 1-1 1.73l-.43.25a2 2 0 0 1-2 0l-.15-.08a2 2 0 0 0-2.73.73l-.22.38a2 2 0 0 0 .73 2.73l.15.1a2 2 0 0 1 1 1.72v.51a2 2 0 0 1-1 1.74l-.15.09a2 2 0 0 0-.73 2.73l.22.38a2 2 0 0 0 2.73.73l.15-.08a2 2 0 0 1 2 0l.43.25a2 2 0 0 1 1 1.73V20a2 2 0 0 0 2 2h.44a2 2 0 0 0 2-2v-.18a2 2 0 0 1 1-1.73l.43-.25a2 2 0 0 1 2 0l.15.08a2 2 0 0 0 2.73-.73l.22-.38a2 2 0 0 0-.73-2.73l-.15-.1a2 2 0 0 1-1-1.72v-.51a2 2 0 0 1 1-1.74l.15-.09a2 2 0 0 0 .73-2.73l-.22-.38a2 2 0 0 0-2.73-.73l-.15.08a2 2 0 0 1-2 0l-.43-.25a2 2 0 0 1-1-1.73V4a2 2 0 0 0-2-2z"></path><circle cx="12" cy="12" r="3"></circle></svg>} 
        />
        <ActionCard 
            label="Create Variant" 
            icon={<svg className="w-6 h-6" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5"><rect x="3" y="3" width="18" height="18" rx="2" ry="2"></rect><line x1="12" y1="8" x2="12" y2="16"></line><line x1="8" y1="12" x2="16" y2="12"></line></svg>} 
        />
        <ActionCard 
            label="Deploy Variants" 
            icon={<svg className="w-6 h-6" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5"><path d="M13 2L3 14h9l-1 8 10-12h-9l1-8z"></path></svg>} 
        />
        <ActionCard 
            label="View Performance" 
            icon={<svg className="w-6 h-6" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5"><circle cx="12" cy="12" r="10"></circle><polyline points="12 6 12 12 16 14"></polyline></svg>} 
        />
      </div>
    </div>
  );
};
