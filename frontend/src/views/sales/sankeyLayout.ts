import { LEAD_STAGES, type LeadFlow, type LeadStage } from '../../models/sales'

export interface SankeyNode {
  key: string
  total: number
  top: number
  height: number
}

export interface SankeyLink {
  source: string
  stage: LeadStage
  total: number
  sourceTop: number
  stageTop: number
  height: number
}

export interface SankeyLayout {
  sources: SankeyNode[]
  stages: SankeyNode[]
  links: SankeyLink[]
  total: number
}

// Tata letak sankey sumber ke tahap: kedua kolom memakai skala yang sama, jadi
// tebal pita di kiri dan kanan selalu sama untuk jumlah prospek yang sama.
// Pita keluar dari tiap sumber diurutkan menurut tahap, dan pita masuk ke tiap
// tahap diurutkan menurut sumber, supaya pita tidak saling memotong di ujung.
export function layoutSankey(flows: LeadFlow[], height: number, gap: number): SankeyLayout {
  const live = flows.filter((flow) => flow.total > 0)
  const total = live.reduce((sum, flow) => sum + flow.total, 0)
  const sourceTotals = new Map<string, number>()
  const stageTotals = new Map<LeadStage, number>()
  for (const flow of live) {
    sourceTotals.set(flow.source, (sourceTotals.get(flow.source) ?? 0) + flow.total)
    stageTotals.set(flow.stage, (stageTotals.get(flow.stage) ?? 0) + flow.total)
  }
  const sourceKeys = [...sourceTotals.keys()].sort(
    (a, b) => (sourceTotals.get(b) ?? 0) - (sourceTotals.get(a) ?? 0) || a.localeCompare(b),
  )
  const stageKeys = LEAD_STAGES.filter((stage) => (stageTotals.get(stage) ?? 0) > 0)
  const busiest = Math.max(sourceKeys.length, stageKeys.length, 1)
  const scale = total > 0 ? Math.max(0, height - gap * (busiest - 1)) / total : 0

  const stack = (keys: string[], totals: Map<string, number>): SankeyNode[] => {
    const used = keys.reduce((sum, key) => sum + (totals.get(key) ?? 0) * scale, 0) + gap * (keys.length - 1)
    let top = (height - used) / 2
    return keys.map((key) => {
      const node = { key, total: totals.get(key) ?? 0, top, height: (totals.get(key) ?? 0) * scale }
      top += node.height + gap
      return node
    })
  }
  const sources = stack(sourceKeys, sourceTotals)
  const stages = stack(stageKeys, stageTotals as Map<string, number>)

  const sourceCursor = new Map(sources.map((node) => [node.key, node.top]))
  const stageCursor = new Map(stages.map((node) => [node.key, node.top]))
  const stageOrder = (stage: LeadStage) => LEAD_STAGES.indexOf(stage)
  const sourceOrder = (source: string) => sourceKeys.indexOf(source)

  const bySource = [...live].sort(
    (a, b) => sourceOrder(a.source) - sourceOrder(b.source) || stageOrder(a.stage) - stageOrder(b.stage),
  )
  const sourceTops = new Map<string, number>()
  for (const flow of bySource) {
    const top = sourceCursor.get(flow.source) ?? 0
    sourceTops.set(`${flow.source}|${flow.stage}`, top)
    sourceCursor.set(flow.source, top + flow.total * scale)
  }

  const byStage = [...live].sort(
    (a, b) => stageOrder(a.stage) - stageOrder(b.stage) || sourceOrder(a.source) - sourceOrder(b.source),
  )
  const links = byStage.map((flow) => {
    const stageTop = stageCursor.get(flow.stage) ?? 0
    stageCursor.set(flow.stage, stageTop + flow.total * scale)
    return {
      source: flow.source,
      stage: flow.stage,
      total: flow.total,
      sourceTop: sourceTops.get(`${flow.source}|${flow.stage}`) ?? 0,
      stageTop,
      height: flow.total * scale,
    }
  })

  return { sources, stages, links, total }
}

export function ribbonPath(link: SankeyLink, left: number, right: number): string {
  const middle = (left + right) / 2
  const top0 = link.sourceTop
  const top1 = link.stageTop
  const bottom0 = top0 + link.height
  const bottom1 = top1 + link.height
  return [
    `M${left},${top0}`,
    `C${middle},${top0} ${middle},${top1} ${right},${top1}`,
    `L${right},${bottom1}`,
    `C${middle},${bottom1} ${middle},${bottom0} ${left},${bottom0}`,
    'Z',
  ].join(' ')
}
