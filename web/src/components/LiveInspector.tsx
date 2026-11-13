import React from "react";
import { Activity, Clock, Zap } from "lucide-react";
import { RequestLog } from "../types";

interface LiveInspectorProps {
  logs: RequestLog[];
  onRefresh: () => void;
}

export const LiveInspector: React.FC<LiveInspectorProps> = ({ logs }) => {
  return (
    <div className="bg-surface-1 border border-surface-border rounded-lg overflow-hidden">
      <div className="p-4 border-b border-surface-border flex items-center justify-between">
        <div>
          <h2 className="text-base font-semibold text-white flex items-center gap-2">
            <Activity className="w-4 h-4 text-emerald-400" />
            <span>Live Proxy Request Stream</span>
          </h2>
          <p className="text-xs text-zinc-400 mt-0.5">
            Transparent sub-3ms reverse proxy execution logs &amp; exact token cost attribution
          </p>
        </div>
      </div>

      <div className="overflow-x-auto">
        <table className="w-full text-left border-collapse text-xs">
          <thead>
            <tr className="border-b border-surface-border bg-surface-2/60 text-zinc-400 uppercase font-mono text-[10px] tracking-wider">
              <th className="py-2.5 px-4">Time</th>
              <th className="py-2.5 px-3">Tenant</th>
              <th className="py-2.5 px-3">Feature</th>
              <th className="py-2.5 px-3">Model</th>
              <th className="py-2.5 px-3">Prompt Tok</th>
              <th className="py-2.5 px-3">Completion</th>
              <th className="py-2.5 px-3">Cached</th>
              <th className="py-2.5 px-3">Latency</th>
              <th className="py-2.5 px-4 text-right">Cost (USD)</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-surface-border/50 font-mono">
            {logs.length === 0 ? (
              <tr>
                <td colSpan={9} className="py-8 text-center text-zinc-500 font-sans">
                  No requests intercepted yet. Send requests to <code>/v1/chat/completions</code>.
                </td>
              </tr>
            ) : (
              logs.map((l) => (
                <tr key={l.id} className="hover:bg-surface-2/70 transition-colors">
                  <td className="py-2.5 px-4 text-zinc-400 text-[11px] whitespace-nowrap">
                    <span className="flex items-center gap-1">
                      <Clock className="w-3 h-3 text-zinc-500" />
                      {new Date(l.timestamp).toLocaleTimeString()}
                    </span>
                  </td>
                  <td className="py-2.5 px-3 font-semibold text-white whitespace-nowrap">
                    {l.tenant_id}
                  </td>
                  <td className="py-2.5 px-3 text-zinc-300">
                    <span className="bg-surface-3 px-1.5 py-0.5 rounded text-[11px]">
                      {l.feature_tag}
                    </span>
                  </td>
                  <td className="py-2.5 px-3 text-blue-300 font-medium whitespace-nowrap">
                    {l.model}
                  </td>
                  <td className="py-2.5 px-3 text-zinc-400">{l.prompt_tokens.toLocaleString()}</td>
                  <td className="py-2.5 px-3 text-zinc-400">{l.completion_tokens.toLocaleString()}</td>
                  <td className="py-2.5 px-3 text-zinc-500">
                    {l.cached_tokens > 0 ? (
                      <span className="text-emerald-400 font-semibold">{l.cached_tokens.toLocaleString()}</span>
                    ) : (
                      "-"
                    )}
                  </td>
                  <td className="py-2.5 px-3 text-zinc-300">
                    <span className="flex items-center gap-0.5 text-[11px]">
                      <Zap className="w-2.5 h-2.5 text-amber-400" />
                      {l.latency_ms}ms
                    </span>
                  </td>
                  <td className="py-2.5 px-4 text-right font-bold text-emerald-400">
                    ${l.cost_usd.toFixed(5)}
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
};
