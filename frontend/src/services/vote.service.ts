import { apiFetch } from './api'
import type { Vote, BoardVote } from './api.type'

export async function getMyVotes(): Promise<Vote[]> {
    return apiFetch('/api/users/me/votes')
}

export async function getBoardVote(boardId: string): Promise<BoardVote> {
    return apiFetch(`/api/boards/${boardId}/vote`)
}

export async function voteBoard(boardId: string, voteValue: number): Promise<void> {
    return apiFetch(`/api/boards/${boardId}/vote`, {
        method: 'PUT',
        body: JSON.stringify({ vote_value: voteValue }),
    })
}

export async function deleteVote(boardId: string): Promise<void> {
    return apiFetch(`/api/boards/${boardId}/vote`, {
        method: 'DELETE',
    })
}
