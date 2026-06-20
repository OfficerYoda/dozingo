import { apiFetch } from './api'
import type { Board, Cell, Game } from './api.type'

export interface BoardsParams {
    sort?: string
    search?: string
    author_id?: string
    limit?: number
}

export async function getBoards(params?: BoardsParams): Promise<Board[]> {
    const query = new URLSearchParams()
    if (params?.sort) query.set('sort', params.sort)
    if (params?.search) query.set('search', params.search)
    if (params?.author_id) query.set('author_id', params.author_id)
    if (params?.limit) query.set('limit', String(params.limit))
    const qs = query.toString() ? '?' + query.toString() : ''
    return apiFetch(`/api/boards${qs}`)
}

export async function getBoardById(boardId: string): Promise<Board> {
    return apiFetch(`/api/boards/${boardId}`)
}

export async function createBoard(title: string, size: number, description?: string): Promise<Board> {
    return apiFetch('/api/boards', {
        method: 'POST',
        body: JSON.stringify({ title, size, ...(description ? { description } : {}) }),
    })
}

export async function getCellsForBoard(boardId: string): Promise<Cell[]> {
    return apiFetch(`/api/boards/${boardId}/cells`)
}

export async function createCell(boardId: string, content: string, value: number): Promise<Cell> {
    return apiFetch(`/api/boards/${boardId}/cells`, {
        method: 'POST',
        body: JSON.stringify({ content, value }),
    })
}

export interface GameCellPosition {
    cell_id: string
    position: number
}

export async function createGame(boardId: string, cells: GameCellPosition[]): Promise<Game> {
    return apiFetch(`/api/boards/${boardId}/games`, {
        method: 'POST',
        body: JSON.stringify(cells),
    })
}

export async function deleteBoard(boardId: string): Promise<void> {
    return apiFetch(`/api/boards/${boardId}`, { method: 'DELETE' })
}
