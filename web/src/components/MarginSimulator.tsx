import React, { useState, useEffect, useCallback } from "react";
import { Calculator, ArrowRight, TrendingUp, CheckCircle, Sliders } from "lucide-react";

interface SimulationResponse {
  original_revenue_usd: number;
  original_cost_usd: number;
  original_margin_usd: number;
  original_margin_pct: number;
  original_unprofitable: number;

  simulated_revenue_usd: number;
  simulated_cost_usd: number;
  simulated_margin_usd: number;
  simulated_margin_pct: number;
  simulated_unprofitable: number;

  monthly_net_gain_usd: number;
}

export const MarginSimulator: React.FC = () => {
  const [subIncrease, setSubIncrease] = useState<number>(0);
  const [migrateSonnet, setMigrateSonnet] = useState<boolean>(true);
  const [migrateGpt4o, setMigrateGpt4o] = useState<boolean>(true);
  const [migrateGpt4Turbo, setMigrateGpt4Turbo] = useState<boolean>(true);
  const [migrateGemini, setMigrateGemini] = useState<boolean>(true);

  const [data, setData] = useState<SimulationResponse | null>(null);
  const [loading, setLoading] = useState<boolean>(false);

  const runSimulation = useCallback(async () => {
    setLoading(true);
    const replacements: Record<string, string> = {};
    if (migrateSonnet) {
      replacements["claude-3-5-sonnet"] = "claude-3-5-haiku";
      replacements["claude-3-5-sonnet-20241022"] = "claude-3-5-haiku-20241022";
    }
    if (migrateGpt4o) {
      replacements["gpt-4o"] = "gpt-4o-mini";
    }
    if (migrateGpt4Turbo) {
      replacements["gpt-4-turbo"] = "gpt-4o-mini";
      replacements["gpt-4"] = "gpt-4o-mini";
    }
    if (migrateGemini) {
      replacements["gemini-1.5-pro"] = "gemini-1.5-flash";
    }

    try {
      const res = await fetch("/api/simulator/simulate", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          subscription_increase_usd: subIncrease,
          model_replacements: replacements,
        }),
      });
      const json = await res.json();
      setData(json);
    } catch (err) {
      console.error("Simulation error:", err);
    } finally {
      setLoading(false);
    }
  }, [subIncrease, migrateSonnet, migrateGpt4o, migrateGpt4Turbo, migrateGemini]);

  useEffect(() => {
    runSimulation();
  }, [runSimulation]);

  return (
    <div className="bg-surface-1 border border-surface-border rounded-lg p-5 space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 border-b border-surface-border pb-4">
        <div>
          <h2 className="text-base font-semibold text-white flex items-center gap-2">
            <Calculator className="w-4 h-4 text-emerald-400" />
            <span>Interactive &quot;What-If&quot; Margin Simulator</span>
          </h2>
          <p className="text-xs text-zinc-400 mt-0.5">
            Model the exact financial return of subscription price changes and model migration policies
          </p>
        </div>

        {data && (
          <div className="bg-emerald-950/40 border border-emerald-500/30 rounded-lg px-3.5 py-2 flex items-center gap-2">
            <TrendingUp className="w-4 h-4 text-emerald-400" />
            <div>
              <span className="text-[10px] text-zinc-400 uppercase font-semibold block">Net Monthly Gain</span>
              <span className="text-sm font-bold font-mono text-emerald-300">
                +${data.monthly_net_gain_usd.toFixed(2)}/mo
              </span>
            </div>
          </div>
        )}
      </div>

      {/* Simulator Controls & Sliders */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-5">
        {/* Left: Pricing & Controls */}
        <div className="space-y-4 bg-surface-2 p-4 rounded-lg border border-surface-border">
          <h3 className="text-xs font-semibold uppercase tracking-wider text-zinc-300 font-mono flex items-center gap-1.5">
            <Sliders className="w-3.5 h-3.5 text-blue-400" />
            <span>Pricing &amp; Migration Variables</span>
          </h3>

          <div>
            <div className="flex items-center justify-between text-xs mb-1.5">
              <span className="text-zinc-300">SaaS Plan Price Adjustment:</span>
              <span className="font-mono font-bold text-white text-sm">
                +${subIncrease.toFixed(0)}/mo
              </span>
            </div>
            <input
              type="range"
              min="0"
              max="50"
              step="5"
              value={subIncrease}
              onChange={(e) => setSubIncrease(parseInt(e.target.value, 10))}
              className="w-full accent-emerald-500 cursor-pointer"
            />
            <div className="flex justify-between text-[10px] text-zinc-500 font-mono mt-1">
              <span>+$0</span>
              <span>+$25</span>
              <span>+$50/mo</span>
            </div>
          </div>

          <div className="pt-2 border-t border-surface-border/60 space-y-2">
            <span className="text-[11px] font-semibold text-zinc-400 uppercase font-mono block">
              Simulate Route Rightsizing Policies:
            </span>

            <label className="flex items-center justify-between text-xs text-zinc-300 p-2 rounded bg-surface-3 cursor-pointer hover:bg-surface-3/80">
              <span className="font-mono">Claude 3.5 Sonnet &rarr; Haiku 3.5</span>
              <input
                type="checkbox"
                checked={migrateSonnet}
                onChange={(e) => setMigrateSonnet(e.target.checked)}
                className="accent-emerald-500 w-4 h-4 rounded"
              />
            </label>

            <label className="flex items-center justify-between text-xs text-zinc-300 p-2 rounded bg-surface-3 cursor-pointer hover:bg-surface-3/80">
              <span className="font-mono">GPT-4o &rarr; GPT-4o-mini</span>
              <input
                type="checkbox"
                checked={migrateGpt4o}
                onChange={(e) => setMigrateGpt4o(e.target.checked)}
                className="accent-emerald-500 w-4 h-4 rounded"
              />
            </label>

            <label className="flex items-center justify-between text-xs text-zinc-300 p-2 rounded bg-surface-3 cursor-pointer hover:bg-surface-3/80">
              <span className="font-mono">GPT-4 Turbo &rarr; GPT-4o-mini</span>
              <input
                type="checkbox"
                checked={migrateGpt4Turbo}
                onChange={(e) => setMigrateGpt4Turbo(e.target.checked)}
                className="accent-emerald-500 w-4 h-4 rounded"
              />
            </label>

            <label className="flex items-center justify-between text-xs text-zinc-300 p-2 rounded bg-surface-3 cursor-pointer hover:bg-surface-3/80">
              <span className="font-mono">Gemini 1.5 Pro &rarr; 1.5 Flash</span>
              <input
                type="checkbox"
                checked={migrateGemini}
                onChange={(e) => setMigrateGemini(e.target.checked)}
                className="accent-emerald-500 w-4 h-4 rounded"
              />
            </label>
          </div>
        </div>

        {/* Right: Before vs After Impact Comparison */}
        <div className="space-y-3">
          {data ? (
            <>
              {/* Revenue Comparison */}
              <div className="bg-surface-2 p-3.5 rounded-lg border border-surface-border">
                <span className="text-[11px] text-zinc-400 font-mono uppercase block">SaaS Monthly Revenue</span>
                <div className="flex items-center justify-between mt-1.5 font-mono">
                  <div>
                    <span className="text-xs text-zinc-500 block">Baseline</span>
                    <span className="text-sm font-semibold text-zinc-300">${data.original_revenue_usd.toFixed(2)}</span>
                  </div>
                  <ArrowRight className="w-4 h-4 text-zinc-600" />
                  <div className="text-right">
                    <span className="text-xs text-emerald-400 block">Simulated</span>
                    <span className="text-base font-bold text-white">${data.simulated_revenue_usd.toFixed(2)}</span>
                  </div>
                </div>
              </div>

              {/* LLM Cloud Invoice Comparison */}
              <div className="bg-surface-2 p-3.5 rounded-lg border border-surface-border">
                <span className="text-[11px] text-zinc-400 font-mono uppercase block">Total LLM Cloud Cost</span>
                <div className="flex items-center justify-between mt-1.5 font-mono">
                  <div>
                    <span className="text-xs text-zinc-500 block">Baseline</span>
                    <span className="text-sm font-semibold text-rose-400">${data.original_cost_usd.toFixed(2)}</span>
                  </div>
                  <ArrowRight className="w-4 h-4 text-zinc-600" />
                  <div className="text-right">
                    <span className="text-xs text-emerald-400 block">Simulated</span>
                    <span className="text-base font-bold text-emerald-400">${data.simulated_cost_usd.toFixed(2)}</span>
                  </div>
                </div>
              </div>

              {/* Gross Margin % Jump */}
              <div className="bg-surface-2 p-3.5 rounded-lg border border-surface-border">
                <span className="text-[11px] text-zinc-400 font-mono uppercase block">Portfolio Gross Margin %</span>
                <div className="flex items-center justify-between mt-1.5 font-mono">
                  <div>
                    <span className="text-xs text-zinc-500 block">Current Margin</span>
                    <span className="text-sm font-semibold text-zinc-300">{data.original_margin_pct.toFixed(1)}%</span>
                  </div>
                  <ArrowRight className="w-4 h-4 text-zinc-600" />
                  <div className="text-right">
                    <span className="text-xs text-emerald-400 block">Projected Margin</span>
                    <span className="text-lg font-bold text-emerald-400">
                      {data.simulated_margin_pct.toFixed(1)}%
                    </span>
                  </div>
                </div>
              </div>

              {/* Unprofitable Accounts Saved */}
              <div className="bg-surface-2 p-3.5 rounded-lg border border-surface-border flex items-center justify-between">
                <div>
                  <span className="text-[11px] text-zinc-400 font-mono uppercase block">Unprofitable Accounts</span>
                  <div className="flex items-center gap-2 mt-1">
                    <span className="text-rose-400 font-mono font-bold text-sm">
                      {data.original_unprofitable} Baseline
                    </span>
                    <span className="text-zinc-600">&rarr;</span>
                    <span className="text-emerald-400 font-mono font-bold text-base flex items-center gap-1">
                      <CheckCircle className="w-4 h-4 text-emerald-400" />
                      {data.simulated_unprofitable} Unprofitable
                    </span>
                  </div>
                </div>
                {data.original_unprofitable > data.simulated_unprofitable && (
                  <span className="px-2 py-1 rounded bg-emerald-500/10 text-emerald-400 border border-emerald-500/30 text-xs font-semibold">
                    +{data.original_unprofitable - data.simulated_unprofitable} Accounts Saved!
                  </span>
                )}
              </div>
            </>
          ) : (
            <div className="py-12 text-center text-zinc-500 text-xs">
              {loading ? "Calculating financial projection..." : "Run a simulation to see results."}
            </div>
          )}
        </div>
      </div>
    </div>
  );
};
