<template>
    <section>
        <div class="container">
            <div class="list-header mb-4">
                <h2 class="mb-0">{{ t('boards.title') }}</h2>
                <div class="header-actions">
                    <input class="btn btn-secondary" type="search" :placeholder="t('boards.searchPlaceholder')" v-model="search">
                    <select class="btn btn-secondary" v-model="appliedFiler">
                        <option value="newest" selected >{{ t('boards.sort.newest') }}</option>
                        <option value="most-liked">{{ t('boards.sort.mostLiked') }}</option>
                        <option value="most-played">{{ t('boards.sort.mostPlayed') }}</option>
                        <option value="oldest">{{ t('boards.sort.oldest') }}</option>
                        <option value="least-liked">{{ t('boards.sort.leastLiked') }}</option>
                        <option value="least-played">{{ t('boards.sort.leastPlayed') }}</option>
                    </select>
                </div>
            </div>

            <p v-if="error" class="error-text">{{ error }}</p>

            <div class="grid">
                <BoardCard
                    v-for="board in boards" :key="board.board_id"
                    :board="board"

                    class="col-4 md-6 sm-12"
                    @click="clickBoard(board.board_id)"
                />
            </div>
        </div>
    </section>

    <ModalStartGame
        v-if="selecetedBoard"
        v-model="showModal"
        :board="selecetedBoard"
        :cells="selectedCells ?? []"
        :author-name="authorName"
    />
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute } from 'vue-router'
import ModalStartGame from '@/components/ModalStartGame.vue'
import BoardCard from '@/components/BoardCard.vue'
import { usePageTitle } from '@/composables/usePageTitle'
import * as boardService from '@/services/board.service'
import * as userService from '@/services/user.service'
import type { Board, Cell } from '@/services/api.type'

const { t } = useI18n()
const route = useRoute()

const error = ref<string | null>(null)
const boards = ref<Board[]>([])
const cells = ref<Cell[]>([])
const selecetedBoard = ref<Board>()
const selectedCells = ref<Cell[]>()
const authorName = ref<string | null>(null)

const { pageTitle } = usePageTitle(t('header.boards'))

async function fetchAllBoards() {
    try {
        boards.value = await boardService.getBoards({
            sort: appliedFiler.value || undefined,
            search: search.value || undefined,
        })
    } catch {
        error.value = t('boards.error.loadBoards')
    }
}

async function fetchAllCellsForBoard(boardID: string) {
    try {
        cells.value = await boardService.getCellsForBoard(boardID)
    } catch {
        error.value = t('boards.error.loadCells') + ' ' + boardID
        return
    }

    selecetedBoard.value = boards.value.find(b => b.board_id === boardID)
    const numberOfCells = (selecetedBoard.value?.size ?? 0) ** 2
    selectedCells.value = [...cells.value].sort(() => Math.random() - 0.5).slice(0, numberOfCells)

    authorName.value = null
    if (selecetedBoard.value?.author_id) {
        try {
            const user = await userService.getUserById(selecetedBoard.value.author_id)
            authorName.value = user.username
        } catch { /* author is optional */ }
    }

    showModal.value = true
}

const appliedFiler = ref(route.query.sort ? String(route.query.sort) : 'newest')
const search = ref('')

watch([appliedFiler, search], () => {
    fetchAllBoards()
}, { immediate: true })

const showModal = ref(false)

function clickBoard(boardID: string) {
    fetchAllCellsForBoard(boardID)
}
</script>

<style scoped>
.list-header {
    display: flex;
    flex-direction: row;
    justify-content: space-between;
}

.header-actions {
    display: flex;
    flex-direction: row;
    gap: 8px;
}
</style>
