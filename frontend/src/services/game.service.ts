import { apiFetch } from './api'
import type { Game, GameCell } from './api.type'

export async function getGameById(gameId: string): Promise<Game> {
    return apiFetch(`/api/games/${gameId}`)
}

export async function getGamesByUser(userId: string): Promise<Game[]> {
    return apiFetch(`/api/users/${userId}/games`)
}

export async function getGameCells(gameId: string): Promise<GameCell[]> {
    return apiFetch(`/api/games/${gameId}/cells`)
}

export async function markGameCell(gameId: string, gameCellId: string, isMarked: boolean): Promise<void> {
    return apiFetch(`/api/games/${gameId}/cells/${gameCellId}`, {
        method: 'PUT',
        body: JSON.stringify({ is_marked: isMarked }),
    })
}

export async function completeGame(gameId: string): Promise<void> {
    return apiFetch(`/api/games/${gameId}/status`, {
        method: 'PUT',
        body: JSON.stringify({ status: 'completed' }),
    })
}
