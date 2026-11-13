import React, { useState } from "react";
import { Sparkles, ArrowRight, DollarSign, Check, Zap, RefreshCw } from "lucide-react";
import { Recommendation } from "../types";

interface RecommendationOptimizerProps {
  recommendations: Recommendation[];
  onApply: (id: string) => void;
  onGenerate: () => void;
}

export const RecommendationOptimizer: React.FC<RecommendationOptimizerProps> = ({
  recommendations,
  onApply,
  onGenerate,
}) => {
  const [loading, setLoading] = useState(false);

  const activeRecs = recommendations.filter((r) => !r.applied);
  const totalPotentialSavings = activeRecs.reduce((sum, r) => sum + r.monthly_savings_usd, 0);

  const handleGenerate = async () => {
    setLoading(true);
    try {
      await onGenerate();
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="bg-surface-1 border border-surface-border rounded-lg p-4">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 mb-4">
        <div>
          <h2 className="text-base font-semibold text-white flex items-center gap-2">
            <Sparkles className="w-4 h-4 text-emerald-400" />
            <span>1-Click Model Downgrade &amp; Rightsizing Advisor</span>
          </h2>
          <p className="text-xs text-zinc-400 mt-0.5">
            Automated model cascade recommendations for routes with negative gross margins
          </p>
        </div>

        <div className="flex items-center gap-3">
          <div className="text-right">
            <span className="text-[11px] text-zinc-400 block">Total Potential Cloud Savings</span>
            <span className="text-base font-bold font-mono text-emerald-400 flex items-center justify-end gap-0.5">
              <DollarSign className="w-4 h-4" />
              {totalPotentialSavings.toFixed(2)}/mo
            </span>
          </div>

          <button
            onClick={handleGenerate}
            disabled={loading}
            className="flex items-center gap-1.5 text-xs font-medium px-3 py-1.5 rounded-md bg-emerald-600 hover:bg-emerald-500 text-white transition-colors disabled:opacity-50"
          >
            <RefreshCw className={`w-3.5 h-3.5 ${loading ? "animate-spin" : ""}`} />
            <span>{loading ? "Analyzing..." : "Re-Analyze Routes"}</span>
          </button>
        </div>
      </div>

      {activeRecs.length === 0 ? (
        <div className="py-10 text-center border border-dashed border-surface-border rounded-lg bg-surface-2/40">
          <Zap className="w-8 h-8 text-emerald-400 mx-auto mb-2 opacity-80" />
          <h3 className="text-sm font-semibold text-white">All Routes Fully Optimized</h3>
          <p className="text-xs text-zinc-400 max-w-md mx-auto mt-1">
            No unprofitable routes or excessive model tier usages detected. Click &quot;Re-Analyze Routes&quot; to re-scan token consumption.
          </p>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
          {activeRecs.map((r) => (
            <div
              key={r.id}
              className="bg-surface-2 border border-surface-border rounded-lg p-3.5 flex flex-col justify-between hover:border-emerald-500/40 transition-colors"
            >
              <div>
                {/* Meta tags */}
                <div className="flex items-center justify-between text-xs mb-2">
                  <div className="flex items-center gap-1.5 font-mono">
                    <span className="px-1.5 py-0.5 rounded bg-surface-3 text-zinc-300 text-[11px]">
                      {r.tenant_id}
                    </span>
                    <span className="text-zinc-500">/</span>
                    <span className="font-semibold text-white">{r.feature_tag}</span>
                  </div>
                  <div className="flex items-center gap-1 text-emerald-400 font-mono font-bold text-xs bg-emerald-950/40 px-2 py-0.5 rounded border border-emerald-800/40">
                    <span>+${r.monthly_savings_usd.toFixed(2)}/mo</span>
                  </div>
                </div>

                {/* Model Transition */}
                <div className="flex items-center gap-2 p-2 rounded bg-surface-1 border border-surface-border/80 my-2 text-xs font-mono">
                  <div className="flex-1 text-rose-300 bg-rose-950/20 px-2 py-1 rounded border border-rose-900/30 truncate">
                    {r.current_model}
                  </div>
                  <ArrowRight className="w-4 h-4 text-zinc-500 shrink-0" />
                  <div className="flex-1 text-emerald-300 bg-emerald-950/20 px-2 py-1 rounded border border-emerald-900/30 truncate">
                    {r.suggested_model}
                  </div>
                </div>

                {/* Reason */}
                <p className="text-xs text-zinc-400 font-sans mt-2">{r.reason}</p>
              </div>

              {/* Action */}
              <div className="mt-3 pt-2.5 border-t border-surface-border/60 flex items-center justify-between">
                <span className="text-[11px] text-zinc-500 font-mono">
                  Calculated from current monthly run rate
                </span>
                <button
                  onClick={() => onApply(r.id)}
                  className="flex items-center gap-1.5 text-xs px-2.5 py-1 rounded bg-emerald-600 hover:bg-emerald-500 text-white font-medium transition-colors"
                >
                  <Check className="w-3.5 h-3.5" />
                  <span>Apply Rightsizing</span>
                </button>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
};
