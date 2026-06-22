import type { Cell } from '@/services/api.type'

export interface BingoLine {
    key: string
    indices: [number, number][]
}

/** Build the full list of lines (rows, cols, diagonals) for a given grid size. */
export function buildLines(size: number): BingoLine[] {
    const lines: BingoLine[] = []
    for (let r = 0; r < size; r++)
        lines.push({ key: `row${r}`, indices: Array.from({ length: size }, (_, c) => [r, c]) })
    for (let c = 0; c < size; c++)
        lines.push({ key: `col${c}`, indices: Array.from({ length: size }, (_, r) => [r, c]) })
    lines.push({ key: 'diag0', indices: Array.from({ length: size }, (_, i) => [i, i]) })
    lines.push({ key: 'diag1', indices: Array.from({ length: size }, (_, i) => [i, size - 1 - i]) })
    return lines
}

/**
 * Silently populate completedLines from existing checked state (no toast/modal).
 * Used on game load when restoring in-progress game state.
 */
export function seedCompletedLines(
    cells: Cell[],
    checked: Set<string>,
    size: number,
    completedLines: Set<string>,
): void {
    const isChecked = (r: number, c: number) => checked.has(cells[r * size + c]?.cell_id ?? '')
    for (const line of buildLines(size)) {
        if (line.indices.every(([r, c]) => isChecked(r, c)))
            completedLines.add(line.key)
    }
}

export interface CheckBingoResult {
    /** Line keys newly completed this call */
    newLines: BingoLine[]
    /** Line keys that were previously complete but are no longer (unchecked cell) */
    removedKeys: string[]
}

/**
 * Diff checked cells against completedLines and return what changed.
 * Mutates completedLines in place.
 */
export function checkBingo(
    cells: Cell[],
    checked: Set<string>,
    size: number,
    completedLines: Set<string>,
): CheckBingoResult {
    const isChecked = (r: number, c: number) => checked.has(cells[r * size + c]?.cell_id ?? '')
    const lines = buildLines(size)

    const removedKeys: string[] = []
    for (const line of lines) {
        if (completedLines.has(line.key) && !line.indices.every(([r, c]) => isChecked(r, c))) {
            completedLines.delete(line.key)
            removedKeys.push(line.key)
        }
    }

    const newLines: BingoLine[] = []
    for (const line of lines) {
        if (completedLines.has(line.key)) continue
        if (line.indices.every(([r, c]) => isChecked(r, c))) {
            completedLines.add(line.key)
            newLines.push(line)
        }
    }

    return { newLines, removedKeys }
}
