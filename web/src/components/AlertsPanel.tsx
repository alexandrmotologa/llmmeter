import React from "react";
import { BellRing, CheckCircle, AlertTriangle, AlertOctagon, Info } from "lucide-react";
import { Alert } from "../types";

interface AlertsPanelProps {
  alerts: Alert[];
  onResolve: (id: string) => void;
}

export const AlertsPanel: React.FC<AlertsPanelProps> = ({ alerts, onResolve }) => {
  const getSeverityIcon = (sev: string) => {
    switch (sev.toLowerCase()) {
      case "critical":
        return <AlertOctagon className="w-4 h-4 text-rose-400" />;
      case "warning":
        return <AlertTriangle className="w-4 h-4 text-amber-400" />;
      default:
        return <Info className="w-4 h-4 text-blue-400" />;
    }
  };

  const getSeverityBadge = (sev: string) => {
    switch (sev.toLowerCase()) {
      case "critical":
        return "bg-rose-500/10 text-rose-400 border-rose-500/30";
      case "warning":
        return "bg-amber-500/10 text-amber-400 border-amber-500/30";
      default:
        return "bg-blue-500/10 text-blue-400 border-blue-500/30";
    }
  };

  return (
    <div className="bg-surface-1 border border-surface-border rounded-lg p-4">
      <div className="flex items-center justify-between mb-4">
        <div>
          <h2 className="text-base font-semibold text-white flex items-center gap-2">
            <BellRing className="w-4 h-4 text-amber-400" />
            <span>FinOps Anomaly &amp; Budget Alerts</span>
          </h2>
          <p className="text-xs text-zinc-400 mt-0.5">
            Automated alerts for token spikes, budget limit caps, and negative margin accounts
          </p>
        </div>
      </div>

      {alerts.length === 0 ? (
        <div className="py-8 text-center text-zinc-500 text-xs">
          No unresolved alerts. All customer usage is operating within normal boundaries.
        </div>
      ) : (
        <div className="space-y-2.5">
          {alerts.map((a) => (
            <div
              key={a.id}
              className="bg-surface-2 border border-surface-border rounded-md p-3 flex flex-col sm:flex-row sm:items-center justify-between gap-3"
            >
              <div className="flex items-start gap-2.5">
                <div className="mt-0.5 shrink-0">{getSeverityIcon(a.severity)}</div>
                <div>
                  <div className="flex items-center gap-2 mb-1">
                    <span
                      className={`text-[10px] font-mono uppercase font-semibold px-1.5 py-0.2 rounded border ${getSeverityBadge(
                        a.severity
                      )}`}
                    >
                      {a.type}
                    </span>
                    <span className="text-xs font-mono font-medium text-white">{a.tenant_id}</span>
                    <span className="text-[11px] text-zinc-500 font-mono">
                      {new Date(a.timestamp).toLocaleString("en-US", {
                        month: "short",
                        day: "numeric",
                        hour: "2-digit",
                        minute: "2-digit",
                      })}
                    </span>
                  </div>
                  <p className="text-xs text-zinc-300 font-sans">{a.message}</p>
                </div>
              </div>

              <div className="shrink-0 flex items-center justify-end">
                <button
                  onClick={() => onResolve(a.id)}
                  className="flex items-center gap-1.5 text-xs px-2.5 py-1 rounded bg-surface-3 hover:bg-emerald-600/30 text-zinc-300 hover:text-emerald-300 border border-zinc-700 hover:border-emerald-500/50 transition-colors"
                >
                  <CheckCircle className="w-3.5 h-3.5" />
                  <span>Resolve</span>
                </button>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
};
