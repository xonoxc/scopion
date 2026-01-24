import React from "react";

export const Header = () => {
  return (
    <header className="h-16 border-b border-white/5 bg-page/50 backdrop-blur-md sticky top-0 z-40 px-8 flex items-center justify-between">
      {/* Breadcrumbs */}
      <div className="flex items-center gap-2 text-[11px] font-bold tracking-widest text-gray-500 uppercase">
        <span className="hover:text-gray-300 cursor-pointer transition-colors">
          Agents
        </span>
        <span className="text-gray-700">•</span>
        <span className="text-gray-300">Wealth Management</span>
      </div>

      {/* Right Side: User Profile */}
      <div className="flex items-center gap-4">
        <div className="text-right hidden sm:block">
          <div className="text-[10px] text-gray-500 uppercase font-bold tracking-wider">
            Logged in as
          </div>
          <div className="text-sm font-medium text-white">Hi, Alex</div>
        </div>

        <div className="w-8 h-8 rounded-full bg-linear-to-br from-gray-700 to-gray-600 ring-2 ring-white/10 overflow-hidden">
          {/* Placeholder Avatar */}
          <img
            src="https://i.pravatar.cc/100?img=33"
            alt="User"
            className="w-full h-full object-cover opacity-80 hover:opacity-100 transition-opacity"
          />
        </div>
      </div>
    </header>
  );
};
