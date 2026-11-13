import React, { useEffect, useState } from "react";
import { Zap, DollarSign, Database, CheckCircle2 } from "lucide-react";

interface CacheFeatureMetric {
  feature_tag: string;
  prompt_tokens: number;
  cached_tokens: number;
  hit_rate_pct: number;
  saved_usd: number;
}

interface CacheAnalytics {
  total_prompt_tokens: number;
  total_cached_tokens: number;
  overall_hit_rate_pct: number;
  total_saved_usd: number;
  top_features: CacheFeatureMetric[];
}

export const CacheRoiAnalytics: React.FC = () => {
  const [data, setData] = useState<CacheAnalytics | null>(null);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    setLoading(true);
    fetch("/api/cache-analytics")
      .then((r) => r.json())
      .then((res: CacheAnalytics) => setData(res))
      .catch((err) => console.error("Cache analytics error:", err))
      .finally(() => setLoading(false));
  }, []);

  return (
    <div className="bg-surface-1 border border-surface-border rounded-lg p-5 space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 border-b border-surface-border pb-4">
        <div>
          <h2 className="text-base font-semibold text-white flex items-center gap-2">
            <Zap className="w-4 h-4 text-emerald-400" />
            <span>Prompt Cache Efficiency &amp; ROI Analytics</span>
          </h2>
          <p className="text-xs text-zinc-400 mt-0.5">
            Real-time attribution of token reuse and dollar savings from Anthropic &amp; OpenAI prompt caching
          </p>
        </div>

        {data && (
          <div className="bg-emerald-950/40 border border-emerald-500/30 rounded-lg px-3.5 py-2 flex items-center gap-2">
            <DollarSign className="w-4 h-4 text-emerald-400" />
            <div>
              <span className="text-[10px] text-zinc-400 uppercase font-semibold block">Dollars Saved via Caching</span>
              <span className="text-base font-bold font-mono text-emerald-300">
                ${data.total_saved_usd.toFixed(2)} USD
              </span>
            </div>
          </div>
        )}
      </div>

      {loading ? (
        <div className="py-12 text-center text-zinc-500 text-xs">Loading prompt cache analytics...</div>
      ) : data ? (
        <>
          {/* Top KPI row */}
          <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
            <div className="bg-surface-2 p-3.5 rounded-lg border border-surface-border">
              <span className="text-[11px] text-zinc-400 font-mono uppercase block">Overall Cache Hit Rate</span>
              <div className="flex items-baseline gap-2 mt-1">
                <span className="text-2xl font-bold font-mono text-emerald-400">
                  {data.overall_hit_rate_pct.toFixed(1)}%
                </span>
                <span className="text-xs text-zinc-500 font-mono">of input tokens</span>
              </div>
            </div>

            <div className="bg-surface-2 p-3.5 rounded-lg border border-surface-border">
              <span className="text-[11px] text-zinc-400 font-mono uppercase block">Total Cached Tokens</span>
              <div className="flex items-baseline gap-2 mt-1">
                <span className="text-2xl font-bold font-mono text-white">
                  {(data.total_cached_tokens / 1_000).toLocaleString("en-US", { maximumFractionDigits: 0 })}k
                </span>
                <span className="text-xs text-zinc-500 font-mono">reused tokens</span>
              </div>
            </div>

            <div className="bg-surface-2 p-3.5 rounded-lg border border-surface-border">
              <span className="text-[11px] text-zinc-400 font-mono uppercase block">Total Prompt Volume</span>
              <div className="flex items-baseline gap-2 mt-1">
                <span className="text-2xl font-bold font-mono text-zinc-300">
                  {(data.total_prompt_tokens / 1_000).toLocaleString("en-US", { maximumFractionDigits: 0 })}k
                </span>
                <span className="text-xs text-zinc-500 font-mono">context tokens</span>
              </div>
            </div>
          </div>

          {/* Per-feature breakdown */}
          <div>
            <h3 className="text-xs font-semibold uppercase tracking-wider text-zinc-300 font-mono mb-3 flex items-center gap-1.5">
              <Database className="w-3.5 h-3.5 text-blue-400" />
              <span>Cache Hit Efficiency per Feature Route</span>
            </h3>

            {data.top_features.length === 0 ? (
              <div className="py-8 text-center text-zinc-500 text-xs bg-surface-2 rounded-lg">
                No cached token usage recorded yet. Anthropic prompt caching activates automatically for system prompts &gt;1,024 tokens.
              </div>
            ) : (
              <div className="space-y-2.5">
                {data.top_features.map((f) => (
                  <div
                    key={f.feature_tag}
                    className="bg-surface-2 border border-surface-border rounded-md p-3 flex flex-col sm:flex-row sm:items-center justify-between gap-3 text-xs font-mono"
                  >
                    <div>
                      <span className="font-semibold text-white text-sm">{f.feature_tag}</span>
                      <div className="flex items-center gap-3 text-zinc-400 text-[11px] mt-0.5">
                        <span>{(f.prompt_tokens / 1_000).toFixed(0)}k prompt tok</span>
                        <span>&bull;</span>
                        <span className="text-emerald-400 font-medium">
                          {(f.cached_tokens / 1_000).toFixed(0)}k cached ({f.hit_rate_pct.toFixed(1)}% hit rate)
                        </span>
                      </div>
                    </div>

                    <div className="text-right">
                      <span className="text-[10px] text-zinc-500 block uppercase font-sans font-semibold">Saved</span>
                      <span className="text-sm font-bold text-emerald-400">
                        +${f.saved_usd.toFixed(2)}
                      </span>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>

          {/* Best Practices Banner */}
          <div className="p-3.5 rounded-lg bg-surface-2/60 border border-surface-border text-xs text-zinc-400 flex items-start gap-2.5 font-sans">
            <CheckCircle2 className="w-4 h-4 text-emerald-400 shrink-0 mt-0.5" />
            <div>
              <strong className="text-zinc-200">Engineering Recommendation:</strong> Keep your system prompt, schemas, and few-shot examples at the exact beginning of messages to ensure 80%+ cache hits on Anthropic and OpenAI.
            </div>
          </div>
        </>
      ) : null}
    </div>
  );
};
