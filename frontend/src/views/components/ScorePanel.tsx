import { DIMENSION_LABEL, RISK_SEVERITY_TONE, type Predictive } from '../../models/site'

// scoreTone colors the overall score band: red < 55, amber < 70, green otherwise.
function scoreTone(score: number): string {
  if (score < 55) {
    return 'text-red-600'
  }
  if (score < 70) {
    return 'text-amber-600'
  }
  return 'text-green-700'
}

export function ScorePanel({ predictive }: { predictive: Predictive }) {
  return (
    <div className="space-y-5">
      <div className="rounded-xl border border-slate-200 bg-slate-50 p-5 text-center">
        <p className="text-xs text-slate-500">Skor kelayakan keseluruhan</p>
        <p className={`mt-1 text-4xl font-semibold tabular-nums ${scoreTone(predictive.overall_score)}`}>
          {predictive.overall_score}
          <span className="text-lg text-slate-400">/100</span>
        </p>
      </div>

      <div className="space-y-3">
        {predictive.dimension_scores.map((dimension) => (
          <div key={dimension.dimension_code}>
            <div className="flex items-baseline justify-between gap-3">
              <span className="text-sm font-medium text-slate-700">
                {DIMENSION_LABEL[dimension.dimension_code] ?? dimension.dimension_code}
              </span>
              <span className="text-sm font-semibold tabular-nums text-slate-900">{dimension.value}</span>
            </div>
            <div className="mt-1 h-1.5 w-full overflow-hidden rounded-full bg-slate-200">
              <div
                className="h-full rounded-full bg-navy-600"
                style={{ width: `${Math.max(0, Math.min(100, dimension.value))}%` }}
              />
            </div>
            {dimension.explanation ? (
              <p className="mt-1 text-xs text-slate-500">{dimension.explanation}</p>
            ) : null}
          </div>
        ))}
      </div>

      {predictive.risk_flags.length > 0 ? (
        <div>
          <p className="mb-2 text-xs font-medium tracking-wide text-slate-500 uppercase">Catatan risiko</p>
          <ul className="space-y-2">
            {predictive.risk_flags.map((flag) => (
              <li key={flag.code} className="flex items-start gap-2">
                <span className={`mt-0.5 rounded px-1.5 py-0.5 text-[10px] font-medium ${RISK_SEVERITY_TONE[flag.severity]}`}>
                  {flag.severity}
                </span>
                <span className="text-xs text-slate-600">{flag.message}</span>
              </li>
            ))}
          </ul>
        </div>
      ) : null}
    </div>
  )
}
