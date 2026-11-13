import React, { useEffect, useState } from "react";
import { X, Calendar, Trash2, FileText } from "lucide-react";
import { TenantDetailResponse } from "../types";

interface TenantModalProps {
  tenantId: string | null;
  onClose: () => void;
  onDeleteTenant: (id: string) => void;
}

export const TenantModal: React.FC<TenantModalProps> = ({ tenantId, onClose, onDeleteTenant }) => {
  const [data, setData] = useState<TenantDetailResponse | null>(null);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (!tenantId) return;
    setLoading(true);
    fetch(`/api/tenants/${tenantId}`)
      .then((r) => r.json())
      .then((res: TenantDetailResponse) => setData(res))
      .catch((err) => console.error(err))
      .finally(() => setLoading(false));
  }, [tenantId]);

  if (!tenantId) return null;

  return (
    <div className="fixed inset-0 z-50 bg-black/75 backdrop-blur-sm flex items-center justify-center p-4">
      <div className="bg-surface-1 border border-surface-border rounded-xl max-w-4xl w-full max-h-[90vh] overflow-y-auto shadow-2xl flex flex-col">
        {/* Modal Header */}
        <div className="p-4 sm:p-5 border-b border-surface-border flex items-center justify-between sticky top-0 bg-surface-1 z-10">
          <div>
            <div className="flex items-center gap-2">
              <h2 className="text-lg font-bold text-white">
                {data?.tenant.name || tenantId}
              </h2>
              <span className="text-xs font-mono px-2 py-0.5 rounded bg-surface-3 text-zinc-400">
                {tenantId}
              </span>
            </div>
            <p className="text-xs text-zinc-400 mt-0.5">
              Tenant Unit Economics, Feature Spending &amp; Audit Log
            </p>
          </div>

          <div className="flex items-center gap-2">
            <a
              href={`/api/tenants/${tenantId}/statement?format=html`}
              target="_blank"
              rel="noopener noreferrer"
              className="flex items-center gap-1.5 text-xs px-2.5 py-1.5 rounded-md bg-surface-2 hover:bg-surface-3 text-zinc-300 border border-surface-border transition-colors"
            >
              <FileText className="w-3.5 h-3.5 text-emerald-400" />
              <span>Official Statement</span>
            </a>

            <button
              onClick={onClose}
              className="p-1.5 rounded-lg text-zinc-400 hover:text-white hover:bg-surface-2 transition-colors"
            >
              <X className="w-5 h-5" />
            </button>
          </div>
        </div>

        {/* Modal Body */}
        <div className="p-4 sm:p-6 space-y-6">
          {loading ? (
            <div className="py-12 text-center text-zinc-500 text-sm">Loading audit data...</div>
          ) : data ? (
            <>
              {/* Financial Summary */}
              <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
                <div className="bg-surface-2 border border-surface-border rounded-lg p-3">
                  <span className="text-[11px] text-zinc-400">Monthly Subscription Plan</span>
                  <div className="text-lg font-bold font-mono text-white mt-1">
                    ${data.tenant.monthly_plan_usd.toFixed(2)}
                  </div>
                </div>

                <div className="bg-surface-2 border border-surface-border rounded-lg p-3">
                  <span className="text-[11px] text-zinc-400">Monthly Budget Cap</span>
                  <div className="text-lg font-bold font-mono text-white mt-1">
                    {data.tenant.monthly_budget_usd ? `$${data.tenant.monthly_budget_usd.toFixed(2)}` : "Unlimited"}
                  </div>
                </div>

                <div className="bg-surface-2 border border-surface-border rounded-lg p-3">
                  <span className="text-[11px] text-zinc-400">Webhook Alert URL</span>
                  <div className="text-xs font-mono text-zinc-300 mt-1 truncate" title={data.tenant.webhook_url}>
                    {data.tenant.webhook_url || "Not configured"}
                  </div>
                </div>
              </div>

              {/* Daily Spend History */}
              <div>
                <h3 className="text-xs font-semibold uppercase tracking-wider text-zinc-400 font-mono mb-2 flex items-center gap-1.5">
                  <Calendar className="w-3.5 h-3.5" />
                  <span>Daily Spend History (Last 30 Days)</span>
                </h3>
                {data.daily.length === 0 ? (
                  <div className="p-4 text-center text-zinc-500 text-xs bg-surface-2 rounded-lg">
                    No daily requests recorded for this account.
                  </div>
                ) : (
                  <div className="grid grid-cols-2 sm:grid-cols-4 lg:grid-cols-6 gap-2">
                    {data.daily.map((d) => (
                      <div key={d.date} className="bg-surface-2 border border-surface-border rounded p-2 text-center font-mono">
                        <span className="text-[10px] text-zinc-500 block">{d.date}</span>
                        <span className="text-xs font-bold text-white block mt-0.5">${d.total_cost_usd.toFixed(2)}</span>
                        <span className="text-[10px] text-zinc-400 block">{d.request_count} reqs</span>
                      </div>
                    ))}
                  </div>
                )}
              </div>

              {/* Feature Breakdown for this tenant */}
              <div>
                <h3 className="text-xs font-semibold uppercase tracking-wider text-zinc-400 font-mono mb-2">
                  Feature Attribution
                </h3>
                <div className="space-y-2">
                  {data.features.map((f) => (
                    <div
                      key={f.feature_tag}
                      className="bg-surface-2 border border-surface-border rounded p-2.5 flex items-center justify-between text-xs font-mono"
                    >
                      <div className="flex items-center gap-2">
                        <span className="font-semibold text-white">{f.feature_tag}</span>
                        {f.top_model && (
                          <span className="text-[10px] px-1.5 py-0.2 rounded bg-surface-3 text-zinc-400">
                            {f.top_model}
                          </span>
                        )}
                      </div>
                      <div className="flex items-center gap-3">
                        <span className="text-zinc-400">{f.request_count} reqs</span>
                        <span className="font-bold text-white">${f.total_cost_usd.toFixed(2)}</span>
                      </div>
                    </div>
                  ))}
                </div>
              </div>

              {/* Recent Requests */}
              <div>
                <h3 className="text-xs font-semibold uppercase tracking-wider text-zinc-400 font-mono mb-2">
                  Recent Request Log
                </h3>
                <div className="bg-surface-2 border border-surface-border rounded-lg overflow-x-auto max-h-48">
                  <table className="w-full text-left text-xs font-mono">
                    <thead className="bg-surface-3/50 text-[10px] text-zinc-400 border-b border-surface-border">
                      <tr>
                        <th className="py-2 px-3">Time</th>
                        <th className="py-2 px-3">Feature</th>
                        <th className="py-2 px-3">Model</th>
                        <th className="py-2 px-3">Tokens</th>
                        <th className="py-2 px-3">Latency</th>
                        <th className="py-2 px-3 text-right">Cost</th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-surface-border/50 text-[11px]">
                      {data.logs.map((l) => (
                        <tr key={l.id} className="hover:bg-surface-3/30">
                          <td className="py-1.5 px-3 text-zinc-400">{new Date(l.timestamp).toLocaleTimeString()}</td>
                          <td className="py-1.5 px-3 text-zinc-300">{l.feature_tag}</td>
                          <td className="py-1.5 px-3 text-blue-300">{l.model}</td>
                          <td className="py-1.5 px-3 text-zinc-400">{l.prompt_tokens + l.completion_tokens}</td>
                          <td className="py-1.5 px-3 text-zinc-400">{l.latency_ms}ms</td>
                          <td className="py-1.5 px-3 text-right font-bold text-emerald-400">${l.cost_usd.toFixed(5)}</td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              </div>

              {/* Delete Tenant Danger Zone */}
              <div className="pt-4 border-t border-surface-border flex items-center justify-between">
                <span className="text-xs text-zinc-500">
                  Delete tenant record and reset attribution metrics
                </span>
                <button
                  onClick={() => {
                    if (confirm(`Are you sure you want to delete tenant ${tenantId}?`)) {
                      onDeleteTenant(tenantId);
                      onClose();
                    }
                  }}
                  className="flex items-center gap-1.5 text-xs text-rose-400 hover:text-white px-3 py-1.5 rounded bg-rose-500/10 hover:bg-rose-600 border border-rose-500/30 transition-colors"
                >
                  <Trash2 className="w-3.5 h-3.5" />
                  <span>Delete Tenant</span>
                </button>
              </div>
            </>
          ) : (
            <div className="py-8 text-center text-zinc-500">Tenant not found.</div>
          )}
        </div>
      </div>
    </div>
  );
};
