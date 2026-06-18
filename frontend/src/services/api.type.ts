export interface User {
    user_id: string
    username: string
    email: string | null
    avatar_url: string | null
}

export interface UserSecurity {
    password_last_changed_at: string
}

export interface Board {
    board_id: string
    title: string
    description: string
    size: number
    author_id: string
    score: number
    vote_count: number
    play_count: number
}

export interface Cell {
    cell_id: string
    content: string
    value: number
}

export interface Game {
    game_id: string
    board_id: string
    status: 'active' | 'completed'
}

export interface GameCell {
    game_cell_id: string
    cell_id: string | null
    content: string
    game_id: string
    is_marked: boolean
    position: number
}

export interface Vote {
    vote_id: string
    board_id: string
    vote_value: number
    title: string
    description: string
    vote_score: number
    vote_count: number
}

export interface BoardVote {
    score: number
    user_vote: number | null
}

export interface Stats {
    bingos: number
    boards: number
    cells: number
    games: number
}

export interface APIError {
    detail: string
    errors: ApiErrorDetails[] | null
    instance: string
    status: number
    title: string
    type: string
}

export interface ApiErrorDetails {
    location: string
    message: string
    value: unknown
}