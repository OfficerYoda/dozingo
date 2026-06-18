import { apiFetch } from './api'
import type { Stats } from './api.type'

export async function getRecentStats(duration = '168h'): Promise<Stats> {
    return apiFetch(`/api/stats/recent?duration=${duration}`)
}
