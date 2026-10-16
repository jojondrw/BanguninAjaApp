import { render, screen, within } from '@testing-library/react'
import { describe, expect, it } from 'vitest'

import {
  budgetItemFormValues,
  EMPTY_BUDGET_ITEM_FORM,
  progressBasis,
  scheduleStatus,
  toBudgetItemRequest,
  type BudgetItem,
  type ProgressCurve,
} from '../models/project'
import { ProgressCurveChart } from '../views/project/ProgressCurveChart'

function curve(overrides: Partial<ProgressCurve> = {}): ProgressCurve {
  return {
    weighting: 'budget',
    undatedPhases: 0,
    phases: [],
    points: [{ date: '2026-01-01', planned: 0, actual: 0 }],
    today: { date: '2026-01-01', planned: 0, actual: 0, deviation: 0 },
    ...overrides,
  }
}

describe('status jadwal dari deviasi kurva S', () => {
  it('menganggap selisih sampai dua poin masih sesuai rencana', () => {
    expect(scheduleStatus(0)).toEqual({ label: 'sesuai rencana', tone: 'good' })
    expect(scheduleStatus(-2)).toEqual({ label: 'sesuai rencana', tone: 'good' })
    expect(scheduleStatus(1.9)).toEqual({ label: 'sesuai rencana', tone: 'good' })
  })

  it('memberi peringatan kalau realisasi tertinggal dari rencana', () => {
    expect(scheduleStatus(-26.4)).toEqual({ label: '26 poin di belakang rencana', tone: 'warn' })
  })

  it('menyebut proyek yang lebih cepat dari rencana', () => {
    expect(scheduleStatus(7.6)).toEqual({ label: '8 poin di depan rencana', tone: 'good' })
  })

  it('tidak menilai jadwal kalau rencana belum bisa dihitung', () => {
    expect(scheduleStatus(null)).toEqual({ label: 'rencana belum lengkap', tone: 'neutral' })
  })
})

describe('dasar perhitungan progres proyek', () => {
  it('menyebut bobot RAB hanya kalau RAB sudah dikaitkan ke tahap', () => {
    expect(progressBasis(curve(), 5)).toBe('berbobot nilai RAB')
    expect(progressBasis(curve({ weighting: 'equal' }), 5)).toBe('rata-rata 5 tahap')
    expect(progressBasis(undefined, 3)).toBe('rata-rata 3 tahap')
    expect(progressBasis(curve(), 0)).toBe('belum ada tahap')
  })
})

describe('tahap pada form item RAB', () => {
  it('mengirim null kalau item belum dikaitkan ke tahap', () => {
    expect(toBudgetItemRequest({ ...EMPTY_BUDGET_ITEM_FORM, code: '1.1' }).phaseId).toBeNull()
  })

  it('mengirim id tahap yang dipilih dan memuatnya kembali saat mengubah item', () => {
    const request = toBudgetItemRequest({ ...EMPTY_BUDGET_ITEM_FORM, phaseId: 'tahap-struktur', code: '2.1' })
    expect(request.phaseId).toBe('tahap-struktur')

    const item: BudgetItem = {
      id: 'item',
      projectId: 'proyek',
      phaseId: 'tahap-struktur',
      code: '2.1',
      description: 'Beton struktur',
      volume: 10,
      unitOfMeasureId: 'm3',
      unitPrice: 1_000,
      total: 10_000,
      createdAt: '2026-01-01T00:00:00Z',
      updatedAt: '2026-01-01T00:00:00Z',
    }
    expect(budgetItemFormValues(item).phaseId).toBe('tahap-struktur')
    expect(budgetItemFormValues({ ...item, phaseId: null }).phaseId).toBe('')
  })
})

describe('grafik kurva S', () => {
  it('menyediakan tabel rencana dan realisasi untuk pembaca layar', () => {
    render(
      <ProgressCurveChart
        today="2026-01-15"
        points={[
          { date: '2026-01-01', planned: 0, actual: 0 },
          { date: '2026-01-15', planned: 12.5, actual: 8 },
          { date: '2026-02-01', planned: 40, actual: null },
        ]}
      />,
    )

    expect(screen.getByRole('img', { name: 'Kurva S progres rencana dan realisasi' })).toBeInTheDocument()
    const rows = within(screen.getByRole('table')).getAllByRole('row')
    expect(rows).toHaveLength(4)
    expect(within(rows[2]).getByText('12,5%')).toBeInTheDocument()
    expect(within(rows[2]).getByText('8%')).toBeInTheDocument()
    expect(within(rows[3]).getByText('belum tercatat')).toBeInTheDocument()
  })
})
