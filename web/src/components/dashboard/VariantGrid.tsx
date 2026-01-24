import React from 'react';

const VariantCard = ({ name, model, description, accuracy }: { name: string; model: string; description: string; accuracy: number }) => (
  <div className="group bg-card border border-white/5 rounded-xl p-5 hover:border-primary/50 hover:shadow-lg hover:shadow-primary/5 transition-all duration-300">
    <div className="flex justify-between items-start mb-4">
        <span className="text-[10px] font-bold text-gray-500 uppercase tracking-widest">{name}</span>
        <span className="text-xs font-mono text-gray-400">Accuracy: <span className="text-white">{accuracy}%</span></span>
    </div>

    <div className="flex items-start gap-3">
        {/* Icon based on model */}
        <div className="w-8 h-8 rounded bg-white/5 flex items-center justify-center text-gray-400 group-hover:text-primary transition-colors border border-white/10">
            {model.includes("GPT") ? (
                <svg className="w-4 h-4" viewBox="0 0 24 24" fill="currentColor"><path d="M12 2a10 10 0 1 0 10 10A10 10 0 0 0 12 2zm0 18a8 8 0 1 1 8-8 8 8 0 0 1-8 8z"/></svg> 
            ) : (
                <svg className="w-4 h-4" viewBox="0 0 24 24" fill="currentColor"><path d="M12 2L2 7l10 5 10-5-10-5zM2 17l10 5 10-5M2 12l10 5 10-5"/></svg>
            )}
        </div>
        
        <div>
            <h4 className="text-sm font-bold text-white mb-1">{model}</h4>
            <p className="text-xs text-gray-500 leading-relaxed max-w-[200px]">{description}</p>
        </div>
    </div>
  </div>
);

export const VariantGrid = () => {
    const variants = [
        { name: "New Variant", model: "GPT-4o", desc: "Economic Risks and Forecasts\nMoody's Analytics", acc: 80 },
        { name: "Variant 8", model: "GPT-4o", desc: "Economic Risks and Forecasts\nMoody's Analytics", acc: 86 },
        { name: "Variant 7", model: "Claude 4.1 Opus", desc: "Economic Risks and Forecasts\nMoody's Analytics", acc: 90 },
        { name: "Variant 6", model: "Claude 4 Opus", desc: "Economic Risks and Forecasts\nMoody's Analytics", acc: 85 },
        { name: "Variant 5", model: "GPT-4o", desc: "Economic Risks and Forecasts\nMoody's Analytics", acc: 84 },
        { name: "Variant 4", model: "Gemini 3", desc: "Economic Risks and Forecasts\nMoody's Analytics", acc: 88 },
    ];
    
  return (
    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6 animate-fade-in-up" style={{ animationDelay: '0.2s' }}>
        {variants.map((v, i) => (
            <VariantCard 
                key={i}
                name={v.name}
                model={v.model}
                description={v.desc}
                accuracy={v.acc}
            />
        ))}
    </div>
  );
};
