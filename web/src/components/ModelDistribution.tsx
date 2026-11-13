import React from "react";
import { Cpu, DollarSign } from "lucide-react";
import { ModelMetric } from "../types";

interface ModelDistributionProps {
  models: ModelMetric[];
}

export const ModelDistribution: React.FC<ModelDistributionProps> = ({ models }) => {
  const totalCost = models.reduce((acc, m) => acc + m.total_cost_usd, 0);

  const getProviderBadge = (provider: string) => {
    switch (provider.toLowerCase()) {
      case "openai":
        return "bg-emerald-500/10 text-emerald-400 border-emerald-500/30";
      case "anthropic":
        return "bg-amber-500/10 text-amber-400 border-amber-500/30";
      case "gemini":
        return "bg-blue-500/10 text-blue-400 border-blue-500/30";
      case "groq":
        return "bg-purple-500/10 text-purple-400 border-purple-500/30";
      default:
        return "bg-zinc-500/10 text-zinc-400 border-zinc-500/30";
    }
  };

  return (
    <div className="bg-surface-1 border border-surface-border rounded-lg p-4">
      <div className="flex items-center justify-between mb-4">
        <div>
          <h2 className="text-base font-semibold text-white flex items-center gap-2">
            <Cpu className="w-4 h-4 text-blue-400" />
            <span>Model Usage &amp; Spend Distribution</span>
          </h2>
          <p className="text-xs text-zinc-400 mt-0.5">
            Breakdown of foundation model consumption across providers
          </p>
        </div>
      </div>

      {models.length === 0 ? (
        <div className="py-8 text-center text-zinc-500 text-xs">
          No model usage recorded yet.
        </div>
      ) : (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
          {models.map((m) => {
            const costShare = totalCost > 0 ? (m.total_cost_usd / totalCost) * 100 : 0;
            return (
              <div
                key={m.model}
                className="bg-surface-2 border border-surface-border rounded-md p-3 flex flex-col justify-between"
              >
                <div>
                  <div className="flex items-center justify-between mb-2">
                    <span
                      className={`text-[10px] font-mono uppercase font-semibold px-1.5 py-0.5 rounded border ${getProviderBadge(
                        m.provider
                      )}`}
                    >
                      {m.provider}
                    </span>
                    <span className="text-[11px] font-mono text-zinc-400">{costShare.toFixed(1)}% spend</span>
                  </div>
                  <div className="font-mono text-sm font-semibold text-white truncate" title={m.model}>
                    {m.model}
                  </div>
                </div>

                <div className="mt-3 pt-2 border-t border-surface-border/50 flex items-center justify-between font-mono text-xs">
                  <div>
                    <span className="text-[10px] text-zinc-500 block">Total Cost</span>
                    <span className="font-bold text-white flex items-center gap-0.5">
                      <DollarSign className="w-3 h-3 text-zinc-400" />
                      {m.total_cost_usd.toFixed(2)}
                    </span>
                  </div>
                  <div className="text-right">
                    <span className="text-[10px] text-zinc-500 block">Volume</span>
                    <span className="text-zinc-300">
                      {(m.total_tokens / 1_000).toFixed(0)}k tok
                    </span>
                  </div>
                </div>
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
};
