import React from "react";
import { Gauge, Download, PlusCircle, RefreshCw } from "lucide-react";

interface NavbarProps {
  currentTab: string;
  setTab: (tab: string) => void;
  unprofitableCount: number;
  activeAlertsCount: number;
  onRefresh: () => void;
  onOpenNewTenant: () => void;
}

export const Navbar: React.FC<NavbarProps> = ({
  currentTab,
  setTab,
  unprofitableCount,
  activeAlertsCount,
  onRefresh,
  onOpenNewTenant,
}) => {
  const tabs = [
    { id: "overview", label: "Executive Overview" },
    {
      id: "tenants",
      label: "Tenant Margins",
      badge: unprofitableCount > 0 ? `${unprofitableCount} Negative` : undefined,
      badgeColor: "bg-rose-500/20 text-rose-400 border border-rose-500/30",
    },
    { id: "features", label: "Feature Breakdown" },
    { id: "simulator", label: "What-If Simulator" },
    { id: "cache", label: "Prompt Cache ROI" },
    {
      id: "recommendations",
      label: "Downgrade Advisor",
    },
    {
      id: "alerts",
      label: "Anomalies & Budget",
      badge: activeAlertsCount > 0 ? `${activeAlertsCount} Alert` : undefined,
      badgeColor: "bg-amber-500/20 text-amber-400 border border-amber-500/30",
    },
    { id: "inspector", label: "Live Inspector" },
  ];

  return (
    <header className="border-b border-surface-border bg-surface-1/90 backdrop-blur sticky top-0 z-40">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex items-center justify-between h-16">
          {/* Logo & Brand */}
          <div className="flex items-center gap-3">
            <div className="w-9 h-9 rounded-lg bg-emerald-500/10 border border-emerald-500/30 flex items-center justify-center text-emerald-400">
              <Gauge className="w-5 h-5" />
            </div>
            <div>
              <div className="flex items-center gap-2">
                <span className="font-bold text-lg tracking-tight text-white font-mono">llmmeter</span>
                <span className="text-[10px] font-semibold uppercase tracking-wider px-1.5 py-0.5 rounded bg-surface-3 text-zinc-400 border border-zinc-700">
                  FinOps Go
                </span>
                <div className="flex items-center gap-1 text-[11px] text-emerald-400 bg-emerald-950/40 px-2 py-0.5 rounded-full border border-emerald-800/40">
                  <span className="w-1.5 h-1.5 rounded-full bg-emerald-400 animate-pulse"></span>
                  Proxy Active &lt;3ms
                </div>
              </div>
              <p className="text-xs text-zinc-400 hidden sm:block">
                Multi-Provider Cost Attribution &amp; Margin Analyzer
              </p>
            </div>
          </div>

          {/* Quick Actions */}
          <div className="flex items-center gap-2">
            <button
              onClick={onOpenNewTenant}
              className="flex items-center gap-1.5 text-xs font-medium px-3 py-1.5 rounded-md bg-emerald-600 hover:bg-emerald-500 text-white transition-colors"
            >
              <PlusCircle className="w-3.5 h-3.5" />
              <span>Add Tenant</span>
            </button>

            <a
              href="/api/export/csv"
              className="flex items-center gap-1.5 text-xs font-medium px-3 py-1.5 rounded-md bg-surface-2 hover:bg-surface-3 text-zinc-300 border border-surface-border transition-colors"
              title="Export Tenant Gross Margins to CSV"
            >
              <Download className="w-3.5 h-3.5" />
              <span className="hidden md:inline">Export CSV</span>
            </a>

            <button
              onClick={onRefresh}
              className="p-1.5 rounded-md bg-surface-2 hover:bg-surface-3 text-zinc-400 hover:text-white border border-surface-border transition-colors"
              title="Refresh Data"
            >
              <RefreshCw className="w-3.5 h-3.5" />
            </button>
          </div>
        </div>

        {/* Tab Navigation */}
        <nav className="flex space-x-1 overflow-x-auto pb-px border-t border-surface-border/50 pt-2 text-xs font-medium">
          {tabs.map((tab) => {
            const isActive = currentTab === tab.id;
            return (
              <button
                key={tab.id}
                onClick={() => setTab(tab.id)}
                className={`flex items-center gap-2 px-3 py-2 rounded-md whitespace-nowrap transition-colors ${
                  isActive
                    ? "bg-surface-3 text-white font-semibold"
                    : "text-zinc-400 hover:text-zinc-200 hover:bg-surface-2"
                }`}
              >
                <span>{tab.label}</span>
                {tab.badge && (
                  <span className={`text-[10px] px-1.5 py-0.2 rounded-full font-mono ${tab.badgeColor}`}>
                    {tab.badge}
                  </span>
                )}
              </button>
            );
          })}
        </nav>
      </div>
    </header>
  );
};
