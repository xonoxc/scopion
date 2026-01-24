import React from 'react';

const MetricRow = ({ label, value, barColor = "bg-primary" }: { label: string; value: string; barColor?: string }) => {
    // Parse percentage for bar width
    const percentage = parseInt(value);
    
    return (
      <div className="flex items-center justify-between group py-3 border-b border-white/5 last:border-0 hover:bg-white/[0.02] px-2 rounded -mx-2 transition-colors">
        <span className="text-sm text-gray-400 font-medium w-32">{label}</span>
        
        {/* Dotted Line */}
        <div className="flex-1 mx-4 border-b-2 border-dotted border-gray-800 relative top-[1px]"></div>
        
        <div className="flex items-center gap-3">
             {/* Mini Bar */}
             <div className="w-24 h-1.5 bg-gray-800 rounded-full overflow-hidden">
                 <div className={`h-full ${barColor} rounded-full`} style={{ width: `${percentage}%` }}></div>
             </div>
             <span className="text-sm font-bold text-white font-mono w-10 text-right">{value}</span>
        </div>
      </div>
    );
};

const StatSmall = ({ label, value }: { label: string; value: string | number }) => (
    <div>
        <div className="text-[10px] uppercase font-bold text-gray-500 tracking-wider mb-1">{label}</div>
        <div className="text-2xl font-light text-white">{value}</div>
    </div>
);

export const VariantMonitor = () => {
  return (
    <div className="bg-card glass-panel rounded-2xl border border-white/10 p-1 mb-12 animate-fade-in-up" style={{ animationDelay: '0.1s' }}>
      
      {/* Header Tab */}
      <div className="flex items-center justify-between px-6 py-4">
        <div className="flex items-center gap-3">
            <h2 className="text-sm font-bold text-white uppercase tracking-wider">Variant 9</h2>
            <span className="px-2 py-0.5 rounded text-[10px] font-bold bg-primary/20 text-primary border border-primary/20">DEPLOYED</span>
        </div>
        
        <div className="flex gap-2">
            <button className="px-4 py-1.5 rounded text-xs font-bold text-gray-400 border border-white/10 hover:text-white hover:border-white/20 transition-colors">EDIT</button>
            <button className="px-4 py-1.5 rounded text-xs font-bold text-white bg-white/10 border border-white/10 hover:bg-white/20 transition-colors">LAUNCH</button>
        </div>
      </div>

      <div className="bg-page/50 rounded-xl border border-white/5 m-1 p-6 grid grid-cols-1 lg:grid-cols-2 gap-12">
        
        {/* Left: Report Card */}
        <div>
            <div className="flex items-center justify-between mb-8">
                <div>
                    <h3 className="text-base font-semibold text-white">Report Card</h3>
                    <p className="text-xs text-gray-500">Last evaluated Jan 14</p>
                </div>
                <button className="text-[10px] font-bold text-gray-400 border border-white/10 px-3 py-1 rounded hover:text-white hover:border-white/20 transition-all">VIEW</button>
            </div>
            
            <div className="space-y-1">
                <MetricRow label="Accuracy" value="90%" barColor="bg-emerald-500" />
                <MetricRow label="Quality" value="100%" barColor="bg-blue-500" />
                <MetricRow label="Retrieval" value="84%" barColor="bg-purple-500" />
                <MetricRow label="Trust and Safety" value="75%" barColor="bg-orange-500" />
            </div>
        </div>

        {/* Right: Chart */}
        <div className="relative">
            <div className="flex items-center justify-between mb-8">
                 <div>
                    <h3 className="text-base font-semibold text-white">Requests (Live)</h3>
                    <p className="text-xs text-gray-500">Jan 7 - Jan 14, 2025</p>
                </div>
                
                <div className="flex gap-4 text-[10px] font-bold text-gray-500">
                    <button className="hover:text-white">24H</button>
                    <button className="text-white border-b-2 border-white pb-0.5">7D</button>
                    <button className="hover:text-white">14D</button>
                    <button className="hover:text-white">1M</button>
                </div>
            </div>

            {/* Chart Area */}
            <div className="h-48 w-full relative">
                {/* Y-Axis Lines */}
                <div className="absolute inset-0 flex flex-col justify-between text-[10px] text-gray-700 font-mono pointer-events-none">
                    <div className="border-t border-dashed border-white/5 w-full h-0"></div>
                    <div className="border-t border-dashed border-white/5 w-full h-0"></div>
                    <div className="border-t border-dashed border-white/5 w-full h-0"></div>
                </div>

                {/* SVG Line Chart */}
                <svg className="w-full h-full overflow-visible" preserveAspectRatio="none">
                    <defs>
                        <linearGradient id="gradient" x1="0" x2="0" y1="0" y2="1">
                            <stop offset="0%" stopColor="#10b981" stopOpacity="0.2" />
                            <stop offset="100%" stopColor="#10b981" stopOpacity="0" />
                        </linearGradient>
                    </defs>
                    
                    {/* Area */}
                    <path 
                        d="M0,150 L20,150 L40,150 L60,80 L80,120 L100,100 L120,150 L140,150 L160,150 L160,200 L0,200 Z" 
                        fill="url(#gradient)" 
                        className="text-emerald-500"
                        transform="scale(3.5, 0.8)" /* Rough scaling adjustment */
                    />
                    
                    {/* Line */}
                    <polyline 
                        points="0,150 70,150 140,80 210,130 280,100 350,150 420,150 600,150" 
                        fill="none" 
                        stroke="#10b981" 
                        strokeWidth="2" 
                        strokeLinecap="round" 
                        strokeLinejoin="round"
                        vectorEffect="non-scaling-stroke"
                    />
                    
                    {/* Points */}
                    <circle cx="140" cy="80" r="4" className="fill-page stroke-emerald-500 stroke-2" />
                    <circle cx="280" cy="100" r="4" className="fill-page stroke-emerald-500 stroke-2" />
                </svg>
                
                {/* X-Axis Labels */}
                <div className="flex justify-between mt-2 text-[10px] text-gray-500 font-mono uppercase">
                    <span>01/7</span>
                    <span>01/8</span>
                    <span>01/9</span>
                    <span>01/10</span>
                    <span>01/11</span>
                    <span>01/12</span>
                    <span>01/13</span>
                     <span>01/14</span>
                </div>
            </div>
            
             {/* Key Metrics on right of chart grid (overlay basically) or seperate */}
             <div className="absolute right-0 top-12 bottom-8 w-px bg-white/5"></div>
             
             <div className="grid grid-cols-2 gap-8 mt-6 pt-6 border-t border-white/5">
                 <StatSmall label="Total Requests" value={<span className="text-emerald-400">3</span>} />
                 <StatSmall label="Users" value="2" />
                 <StatSmall label="Total Errors" value="0" />
                 <StatSmall label="Avg Latency" value="6,374 ms" />
             </div>
        </div>
      </div>
    </div>
  );
};
