import { describe, it, expect, beforeEach } from 'vitest'
import { buildLines, seedCompletedLines, checkBingo } from '../bingo'
import type { Cell } from '@/services/api.type'

/** Build a size×size grid of cells with id = "r_c" */
function makeGrid(size: number): Cell[] {
    return Array.from({ length: size * size }, (_, i) => ({
        cell_id: `${Math.floor(i / size)}_${i % size}`,
        content: `cell${i}`,
        value: 0,
    }))
}

function id(r: number, c: number) { return `${r}_${c}` }

describe('buildLines', () => {
    it('returns correct count for 4x4 (4 rows + 4 cols + 2 diags = 10)', () => {
        expect(buildLines(4)).toHaveLength(10)
    })

    it('returns correct count for 5x5 (5+5+2=12)', () => {
        expect(buildLines(5)).toHaveLength(12)
    })

    it('diagonal keys are diag0 and diag1', () => {
        const keys = buildLines(4).map(l => l.key)
        expect(keys).toContain('diag0')
        expect(keys).toContain('diag1')
    })
})

describe('seedCompletedLines', () => {
    it('seeds a complete row', () => {
        const cells = makeGrid(4)
        const checked = new Set([id(0, 0), id(0, 1), id(0, 2), id(0, 3)])
        const completed = new Set<string>()
        seedCompletedLines(cells, checked, 4, completed)
        expect(completed).toContain('row0')
    })

    it('seeds a complete column', () => {
        const cells = makeGrid(4)
        const checked = new Set([id(0, 2), id(1, 2), id(2, 2), id(3, 2)])
        const completed = new Set<string>()
        seedCompletedLines(cells, checked, 4, completed)
        expect(completed).toContain('col2')
    })

    it('seeds main diagonal (diag0)', () => {
        const cells = makeGrid(4)
        const checked = new Set([id(0, 0), id(1, 1), id(2, 2), id(3, 3)])
        const completed = new Set<string>()
        seedCompletedLines(cells, checked, 4, completed)
        expect(completed).toContain('diag0')
    })

    it('seeds anti-diagonal (diag1)', () => {
        const cells = makeGrid(4)
        const checked = new Set([id(0, 3), id(1, 2), id(2, 1), id(3, 0)])
        const completed = new Set<string>()
        seedCompletedLines(cells, checked, 4, completed)
        expect(completed).toContain('diag1')
    })

    it('does not seed incomplete lines', () => {
        const cells = makeGrid(4)
        const checked = new Set([id(0, 0), id(0, 1), id(0, 2)]) // row0 missing one
        const completed = new Set<string>()
        seedCompletedLines(cells, checked, 4, completed)
        expect(completed).not.toContain('row0')
    })
})

describe('checkBingo', () => {
    let cells: Cell[]
    let completed: Set<string>

    beforeEach(() => {
        cells = makeGrid(4)
        completed = new Set()
    })

    it('detects a new row bingo', () => {
        const checked = new Set([id(1, 0), id(1, 1), id(1, 2), id(1, 3)])
        const result = checkBingo(cells, checked, 4, completed)
        expect(result.newLines.map(l => l.key)).toContain('row1')
        expect(completed).toContain('row1')
    })

    it('detects a new column bingo', () => {
        const checked = new Set([id(0, 0), id(1, 0), id(2, 0), id(3, 0)])
        const result = checkBingo(cells, checked, 4, completed)
        expect(result.newLines.map(l => l.key)).toContain('col0')
    })

    it('detects main diagonal bingo', () => {
        const checked = new Set([id(0, 0), id(1, 1), id(2, 2), id(3, 3)])
        const result = checkBingo(cells, checked, 4, completed)
        expect(result.newLines.map(l => l.key)).toContain('diag0')
    })

    it('detects anti-diagonal bingo', () => {
        const checked = new Set([id(0, 3), id(1, 2), id(2, 1), id(3, 0)])
        const result = checkBingo(cells, checked, 4, completed)
        expect(result.newLines.map(l => l.key)).toContain('diag1')
    })

    it('does not re-report an already-completed line', () => {
        const checked = new Set([id(0, 0), id(0, 1), id(0, 2), id(0, 3)])
        completed.add('row0')
        const result = checkBingo(cells, checked, 4, completed)
        expect(result.newLines.map(l => l.key)).not.toContain('row0')
    })

    it('removes a line from completed when a cell is unchecked', () => {
        completed.add('row0')
        // row0 is now incomplete (only 3/4 checked)
        const checked = new Set([id(0, 0), id(0, 1), id(0, 2)])
        const result = checkBingo(cells, checked, 4, completed)
        expect(result.removedKeys).toContain('row0')
        expect(completed).not.toContain('row0')
    })

    it('reports multiple simultaneous bingos', () => {
        // row0 + col0 both complete at once via corner
        const checked = new Set([
            id(0, 0), id(0, 1), id(0, 2), id(0, 3), // row0
            id(1, 0), id(2, 0), id(3, 0),             // rest of col0
        ])
        const result = checkBingo(cells, checked, 4, completed)
        const keys = result.newLines.map(l => l.key)
        expect(keys).toContain('row0')
        expect(keys).toContain('col0')
    })

    it('returns empty arrays when nothing changed', () => {
        const result = checkBingo(cells, new Set(), 4, completed)
        expect(result.newLines).toHaveLength(0)
        expect(result.removedKeys).toHaveLength(0)
    })
})
