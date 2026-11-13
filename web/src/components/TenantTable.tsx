import React, { useState } from "react";
import { Search, AlertOctagon, TrendingDown, TrendingUp, ChevronRight, Filter } from "lucide-react";
import { TenantMetric } from "../types";

interface TenantTableProps {
  tenants: TenantMetric[];
  onSelectTenant: (tenantId: string) => void;
  unprofitableOnly?: boolean;
}

export const TenantTable: React.FC<TenantTableProps> = ({
  tenants,
  onSelectTenant,
  unprofitableOnly = false,
}) => {
  const [search, setSearch] = useState("");
  const [filterUnprofitable, setFilterUnprofitable] = useState(unprofitableOnly);

  const filtered = tenants.filter((t) => {
    const matchesSearch =
      t.tenant_name.toLowerCase().includes(search.toLowerCase()) ||
      t.tenant_id.toLowerCase().includes(search.toLowerCase());
    if (!matchesSearch) return false;
    if (filterUnprofitable) return t.is_unprofitable;
    return true;
  });

  return (
    <div className="bg-surface-1 border border-surface-border rounded-lg overflow-hidden">
      {/* Header & Controls */}
      <div className="p-4 border-b border-surface-border flex flex-col sm:flex-row sm:items-center justify-between gap-3">
        <div>
          <h2 className="text-base font-semibold text-white">Tenant Chargeback &amp; Unit Economics</h2>
          <p className="text-xs text-zinc-400 mt-0.5">
            Real-time Gross Margin Heatmap per SaaS customer
          </p>
        </div>

        <div className="flex items-center gap-2">
          {/* Search */}
          <div className="relative">
            <Search className="w-3.5 h-3.5 text-zinc-400 absolute left-2.5 top-2.5" />
            <input
              type="text"
              placeholder="Search tenant..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              className="pl-8 pr-3 py-1.5 text-xs bg-surface-2 border border-surface-border rounded-md text-white placeholder-zinc-500 focus:outline-none focus:border-zinc-500 w-48 sm:w-56"
            />
          </div>

          {/* Negative Margin Toggle */}
          <button
            onClick={() => setFilterUnprofitable(!filterUnprofitable)}
            className={`flex items-center gap-1.5 text-xs px-2.5 py-1.5 rounded-md border transition-colors ${
              filterUnprofitable
                ? "bg-rose-500/20 text-rose-300 border-rose-500/40"
                : "bg-surface-2 text-zinc-400 border-surface-border hover:text-white"
            }`}
          >
            <Filter className="w-3.5 h-3.5" />
            <span>Unprofitable Only</span>
          </button>
        </div>
      </div>

      {/* Table */}
      <div className="overflow-x-auto">
        <table className="w-full text-left border-collapse text-xs">
          <thead>
            <tr className="border-b border-surface-border bg-surface-2/60 text-zinc-400 uppercase font-mono text-[10px] tracking-wider">
              <th className="py-2.5 px-4">Tenant / Customer</th>
              <th className="py-2.5 px-3">Subscription</th>
              <th className="py-2.5 px-3">LLM Cost</th>
              <th className="py-2.5 px-3">Gross Margin ($)</th>
              <th className="py-2.5 px-3">Gross Margin (%)</th>
              <th className="py-2.5 px-3">Budget Cap</th>
              <th className="py-2.5 px-3">Tokens</th>
              <th className="py-2.5 px-3">Requests</th>
              <th className="py-2.5 px-4 text-right">Action</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-surface-border/50 font-mono">
            {filtered.length === 0 ? (
              <tr>
                <td colSpan={9} className="py-8 text-center text-zinc-500 font-sans">
                  No tenants found matching your filter.
                </td>
              </tr>
            ) : (
              filtered.map((t) => {
                const isNegative = t.is_unprofitable;
                const isHealthy = t.gross_margin_pct >= 25.0;

                // Budget calculation
                let budgetText = "No Cap";
                let budgetColor = "text-zinc-500";
                if (t.monthly_budget_usd && t.monthly_budget_usd > 0) {
                  const pctUsed = (t.total_cost_usd / t.monthly_budget_usd) * 100;
                  budgetText = `$${t.monthly_budget_usd.toFixed(0)} (${pctUsed.toFixed(0)}%)`;
                  if (pctUsed >= 100) {
                    budgetColor = "text-rose-400 font-semibold";
                  } else if (pctUsed >= 80) {
                    budgetColor = "text-amber-400";
                  } else {
                    budgetColor = "text-zinc-300";
                  }
                }

                return (
                  <tr
                    key={t.tenant_id}
                    className={`hover:bg-surface-2/80 transition-colors ${
                      isNegative ? "bg-rose-950/10" : ""
                    }`}
                  >
                    {/* Tenant Info */}
                    <td className="py-3 px-4">
                      <div className="flex items-center gap-2">
                        {isNegative && (
                          <span title="Negative Gross Margin!">
                            <AlertOctagon className="w-4 h-4 text-rose-400 shrink-0" />
                          </span>
                        )}
                        <div>
                          <div className="font-sans font-medium text-white flex items-center gap-1.5">
                            <span>{t.tenant_name}</span>
                          </div>
                          <span className="text-[11px] text-zinc-500">{t.tenant_id}</span>
                        </div>
                      </div>
                    </td>

                    {/* Subscription */}
                    <td className="py-3 px-3 text-zinc-300">
                      ${t.monthly_plan_usd.toFixed(2)}
                      <span className="text-[10px] text-zinc-500 font-sans block">/month</span>
                    </td>

                    {/* LLM Cost */}
                    <td className="py-3 px-3 text-rose-400 font-semibold">
                      ${t.total_cost_usd.toFixed(2)}
                    </td>

                    {/* Gross Margin ($) */}
                    <td className="py-3 px-3">
                      <span
                        className={`font-semibold flex items-center gap-1 ${
                          isNegative ? "text-rose-400" : isHealthy ? "text-emerald-400" : "text-amber-400"
                        }`}
                      >
                        {isNegative ? (
                          <TrendingDown className="w-3 h-3 shrink-0" />
                        ) : (
                          <TrendingUp className="w-3 h-3 shrink-0" />
                        )}
                        ${t.gross_margin_usd.toFixed(2)}
                      </span>
                    </td>

                    {/* Gross Margin (%) */}
                    <td className="py-3 px-3">
                      <span
                        className={`text-[11px] px-2 py-0.5 rounded font-semibold inline-block ${
                          isNegative
                            ? "bg-rose-500/20 text-rose-300 border border-rose-500/40"
                            : isHealthy
                            ? "bg-emerald-500/20 text-emerald-300 border border-emerald-500/40"
                            : "bg-amber-500/20 text-amber-300 border border-amber-500/40"
                        }`}
                      >
                        {t.gross_margin_pct.toFixed(1)}%
                      </span>
                    </td>

                    {/* Budget Cap */}
                    <td className={`py-3 px-3 text-[11px] ${budgetColor}`}>
                      {budgetText}
                    </td>

                    {/* Tokens */}
                    <td className="py-3 px-3 text-zinc-400 text-[11px]">
                      {(t.total_tokens / 1_000).toLocaleString("en-US", { maximumFractionDigits: 0 })}k
                    </td>

                    {/* Requests */}
                    <td className="py-3 px-3 text-zinc-400 text-[11px]">
                      {t.request_count.toLocaleString("en-US")}
                    </td>

                    {/* Action */}
                    <td className="py-3 px-4 text-right">
                      <button
                        onClick={() => onSelectTenant(t.tenant_id)}
                        className="inline-flex items-center gap-1 text-xs text-zinc-400 hover:text-white px-2 py-1 rounded bg-surface-2 hover:bg-surface-3 transition-colors border border-surface-border font-sans"
                      >
                        <span>Audit</span>
                        <ChevronRight className="w-3.5 h-3.5" />
                      </button>
                    </td>
                  </tr>
                );
              })
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
};
