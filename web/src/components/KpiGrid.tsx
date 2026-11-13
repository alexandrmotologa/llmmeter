import React from "react";
import { DollarSign, Cpu, TrendingUp, TrendingDown, AlertOctagon, BellRing } from "lucide-react";
import { OverviewSummary } from "../types";

interface KpiGridProps {
  summary: OverviewSummary | null;
  onFilterUnprofitable: () => void;
  onViewAlerts: () => void;
}

export const KpiGrid: React.FC<KpiGridProps> = ({ summary, onFilterUnprofitable, onViewAlerts }) => {
  if (!summary) return null;

  const isMarginHealthy = summary.net_margin_pct >= 20.0;
  const isMarginNegative = summary.net_margin_pct < 0;

  return (
    <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-6 gap-3">
      {/* Revenue */}
      <div className="bg-surface-1 border border-surface-border rounded-lg p-3.5 flex flex-col justify-between">
        <div className="flex items-center justify-between text-zinc-400 text-xs">
          <span>Monthly Subscription</span>
          <DollarSign className="w-4 h-4 text-zinc-400" />
        </div>
        <div className="mt-2">
          <div className="text-xl font-bold font-mono text-white">
            ${summary.total_revenue_usd.toLocaleString("en-US", { minimumFractionDigits: 2, maximumFractionDigits: 2 })}
          </div>
          <p className="text-[11px] text-zinc-500 mt-0.5">SaaS Tiers Invoiced</p>
        </div>
      </div>

      {/* LLM Spend */}
      <div className="bg-surface-1 border border-surface-border rounded-lg p-3.5 flex flex-col justify-between">
        <div className="flex items-center justify-between text-zinc-400 text-xs">
          <span>Total LLM Cost</span>
          <Cpu className="w-4 h-4 text-zinc-400" />
        </div>
        <div className="mt-2">
          <div className="text-xl font-bold font-mono text-rose-400">
            ${summary.total_cost_usd.toLocaleString("en-US", { minimumFractionDigits: 2, maximumFractionDigits: 2 })}
          </div>
          <p className="text-[11px] text-zinc-500 mt-0.5">OpenAI + Anthropic + Cloud</p>
        </div>
      </div>

      {/* Net Gross Margin */}
      <div className="bg-surface-1 border border-surface-border rounded-lg p-3.5 flex flex-col justify-between">
        <div className="flex items-center justify-between text-zinc-400 text-xs">
          <span>Net AI Gross Margin</span>
          {isMarginNegative ? (
            <TrendingDown className="w-4 h-4 text-rose-400" />
          ) : (
            <TrendingUp className="w-4 h-4 text-emerald-400" />
          )}
        </div>
        <div className="mt-2">
          <div className="flex items-baseline gap-2">
            <span
              className={`text-xl font-bold font-mono ${
                isMarginNegative ? "text-rose-400" : isMarginHealthy ? "text-emerald-400" : "text-amber-400"
              }`}
            >
              ${summary.net_margin_usd.toLocaleString("en-US", { minimumFractionDigits: 2, maximumFractionDigits: 2 })}
            </span>
          </div>
          <div className="flex items-center gap-1.5 mt-0.5">
            <span
              className={`text-[10px] font-mono font-semibold px-1.5 py-0.2 rounded ${
                isMarginNegative
                  ? "bg-rose-500/20 text-rose-400 border border-rose-500/30"
                  : isMarginHealthy
                  ? "bg-emerald-500/20 text-emerald-400 border border-emerald-500/30"
                  : "bg-amber-500/20 text-amber-400 border border-amber-500/30"
              }`}
            >
              {summary.net_margin_pct.toFixed(1)}% Margin
            </span>
            <span className="text-[11px] text-zinc-500">Gross Return</span>
          </div>
        </div>
      </div>

      {/* Negative Margin Accounts */}
      <div
        onClick={onFilterUnprofitable}
        className={`bg-surface-1 border rounded-lg p-3.5 flex flex-col justify-between cursor-pointer transition-all hover:border-rose-500/50 ${
          summary.unprofitable_count > 0 ? "border-rose-500/40 bg-rose-950/10" : "border-surface-border"
        }`}
      >
        <div className="flex items-center justify-between text-zinc-400 text-xs">
          <span>Negative Margin Clients</span>
          <AlertOctagon className={`w-4 h-4 ${summary.unprofitable_count > 0 ? "text-rose-400 animate-pulse" : "text-zinc-400"}`} />
        </div>
        <div className="mt-2">
          <div className="text-xl font-bold font-mono text-rose-400">
            {summary.unprofitable_count} <span className="text-xs text-zinc-400 font-sans">/ {summary.active_tenants_count}</span>
          </div>
          <p className="text-[11px] text-zinc-400 mt-0.5 flex items-center gap-1">
            {summary.unprofitable_count > 0 ? "Silent Margin Bleed!" : "All accounts profitable"}
          </p>
        </div>
      </div>

      {/* Total Token Volume */}
      <div className="bg-surface-1 border border-surface-border rounded-lg p-3.5 flex flex-col justify-between">
        <div className="flex items-center justify-between text-zinc-400 text-xs">
          <span>Token Volume</span>
          <Cpu className="w-4 h-4 text-zinc-400" />
        </div>
        <div className="mt-2">
          <div className="text-xl font-bold font-mono text-white">
            {(summary.total_tokens_month / 1_000_000).toFixed(2)}M
          </div>
          <p className="text-[11px] text-zinc-500 mt-0.5">Input + Output + Cached</p>
        </div>
      </div>

      {/* Active Alerts */}
      <div
        onClick={onViewAlerts}
        className={`bg-surface-1 border rounded-lg p-3.5 flex flex-col justify-between cursor-pointer transition-all hover:border-amber-500/50 ${
          summary.active_alerts_count > 0 ? "border-amber-500/40 bg-amber-950/10" : "border-surface-border"
        }`}
      >
        <div className="flex items-center justify-between text-zinc-400 text-xs">
          <span>Active Alerts</span>
          <BellRing className={`w-4 h-4 ${summary.active_alerts_count > 0 ? "text-amber-400 animate-bounce" : "text-zinc-400"}`} />
        </div>
        <div className="mt-2">
          <div className="text-xl font-bold font-mono text-amber-400">
            {summary.active_alerts_count}
          </div>
          <p className="text-[11px] text-zinc-400 mt-0.5">Spikes &amp; Budget Caps</p>
        </div>
      </div>
    </div>
  );
};
