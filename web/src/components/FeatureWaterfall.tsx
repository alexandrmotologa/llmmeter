import React from "react";
import { Layers, ArrowUpRight } from "lucide-react";
import { FeatureMetric } from "../types";

interface FeatureWaterfallProps {
  features: FeatureMetric[];
  onSelectFeature?: (featureTag: string) => void;
}

export const FeatureWaterfall: React.FC<FeatureWaterfallProps> = ({ features }) => {
  const maxCost = features.length > 0 ? features[0].total_cost_usd : 1;
  const totalSpend = features.reduce((sum, f) => sum + f.total_cost_usd, 0);

  return (
    <div className="bg-surface-1 border border-surface-border rounded-lg p-4">
      <div className="flex items-center justify-between mb-4">
        <div>
          <h2 className="text-base font-semibold text-white flex items-center gap-2">
            <Layers className="w-4 h-4 text-emerald-400" />
            <span>Feature Cost Attribution Waterfall</span>
          </h2>
          <p className="text-xs text-zinc-400 mt-0.5">
            Identify which product features drive the largest share of the cloud AI invoice
          </p>
        </div>
        <div className="text-right">
          <span className="text-xs text-zinc-400">Total Feature Spend</span>
          <div className="text-base font-bold font-mono text-white">
            ${totalSpend.toFixed(2)}
          </div>
        </div>
      </div>

      {features.length === 0 ? (
        <div className="py-8 text-center text-zinc-500 text-xs">
          No feature tag usage recorded yet. Tag requests with <code>X-Feature-Tag</code>.
        </div>
      ) : (
        <div className="space-y-3">
          {features.map((f, idx) => {
            const pct = maxCost > 0 ? (f.total_cost_usd / maxCost) * 100 : 0;
            const shareOfTotal = totalSpend > 0 ? (f.total_cost_usd / totalSpend) * 100 : 0;

            return (
              <div
                key={f.feature_tag}
                className="bg-surface-2/60 border border-surface-border/60 rounded-md p-3 hover:border-zinc-600 transition-colors"
              >
                <div className="flex items-center justify-between text-xs mb-1.5">
                  <div className="flex items-center gap-2">
                    <span className="text-[10px] font-mono text-zinc-500 w-4">#{idx + 1}</span>
                    <span className="font-mono font-semibold text-white text-sm">{f.feature_tag}</span>
                    {f.top_model && (
                      <span className="text-[10px] px-1.5 py-0.5 rounded bg-surface-3 text-zinc-300 font-mono">
                        {f.top_model}
                      </span>
                    )}
                  </div>
                  <div className="flex items-center gap-3 font-mono">
                    <span className="text-zinc-400 text-[11px]">{shareOfTotal.toFixed(1)}% of total</span>
                    <span className="font-bold text-white text-sm">${f.total_cost_usd.toFixed(2)}</span>
                  </div>
                </div>

                {/* Waterfall Visual Bar */}
                <div className="w-full bg-surface-3 h-2 rounded-full overflow-hidden mb-2">
                  <div
                    className="bg-gradient-to-r from-emerald-500 to-blue-500 h-full rounded-full transition-all duration-500"
                    style={{ width: `${Math.max(pct, 2)}%` }}
                  />
                </div>

                {/* Sub-metrics */}
                <div className="flex items-center justify-between text-[11px] text-zinc-400 font-mono">
                  <span>{(f.total_tokens / 1_000).toFixed(0)}k tokens</span>
                  <span>{f.request_count.toLocaleString("en-US")} requests</span>
                  <span className="flex items-center gap-1 text-zinc-300">
                    avg ${f.avg_cost_usd.toFixed(4)}/req
                    <ArrowUpRight className="w-3 h-3 text-zinc-500" />
                  </span>
                </div>
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
};
