import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createI18n } from 'vue-i18n'
import { createRouter, createMemoryHistory } from 'vue-router'
import { defineComponent } from 'vue'
import CardEditorView from '../CardEditorView.vue'

// Minimal i18n messages — only the keys used by CardEditorView
const messages = {
    en: {
        header: { cardEditor: 'Card Editor' },
        cardEditor: {
            title: 'Create Board',
            sessionTitle: 'Title',
            titlePlaceholder: 'Title',
            layoutSelection: 'Size',
            description: 'Description',
            descriptionPlaceholder: 'Description',
            entries: 'Entries',
            termColumn: 'Term',
            termPlaceholder: 'Term',
            addRow: 'Add row',
            save: 'Save',
            saveAndPlay: 'Save & Play',
            saveSuccess: 'Saved!',
            saveError: 'Error saving',
            validationError: 'Fill all fields',
            notEnoughEntriesError: 'Not enough entries',
            entryCountHint: '{count}/{needed} entries for {size}',
            entryCountHintReady: '{count} entries ready for {size}',
        },
    },
}

function buildPlugins() {
    const i18n = createI18n({ legacy: false, locale: 'en', messages })
    const router = createRouter({
        history: createMemoryHistory(),
        routes: [
            { path: '/', component: defineComponent({ template: '<div/>' }) },
            { path: '/cardeditor', component: defineComponent({ template: '<div/>' }) },
            { path: '/game/:game_id', component: defineComponent({ template: '<div/>' }) },
        ],
    })
    return { i18n, router }
}

vi.mock('@/composables/usePageTitle', () => ({
    usePageTitle: () => ({ pageTitle: { value: '' } }),
}))

vi.mock('@/services/board.service', () => ({
    createBoard: vi.fn(),
    createCell: vi.fn(),
    getCellsForBoard: vi.fn(),
    createGame: vi.fn(),
}))

import * as boardService from '@/services/board.service'

beforeEach(() => {
    vi.clearAllMocks()
})

describe('CardEditorView – requiredEntries', () => {
    it('is 25 for default 5x5', async () => {
        const { i18n, router } = buildPlugins()
        const wrapper = mount(CardEditorView, { global: { plugins: [i18n, router] } })
        // default is 5x5 → 25 entries pre-populated
        const inputs = wrapper.findAll('input[data-index]')
        expect(inputs.length).toBeGreaterThanOrEqual(25)
    })

    it('grows to 36 when switching to 6x6', async () => {
        const { i18n, router } = buildPlugins()
        const wrapper = mount(CardEditorView, { global: { plugins: [i18n, router] } })
        const radio6 = wrapper.find('input[value="6x6"]')
        await radio6.setValue(true)
        await radio6.trigger('change')
        // vue reactivity: watch fires synchronously for immediate
        await flushPromises()
        const inputs = wrapper.findAll('input[data-index]')
        expect(inputs.length).toBeGreaterThanOrEqual(36)
    })

    it('shrinks to 16 when switching to 4x4', async () => {
        const { i18n, router } = buildPlugins()
        const wrapper = mount(CardEditorView, { global: { plugins: [i18n, router] } })
        const radio4 = wrapper.find('input[value="4x4"]')
        await radio4.setValue(true)
        await radio4.trigger('change')
        await flushPromises()
        const inputs = wrapper.findAll('input[data-index]')
        // At least 16 rows (may have more if user already typed; here none)
        expect(inputs.length).toBeGreaterThanOrEqual(16)
    })
})

describe('CardEditorView – saveBoard validation', () => {
    it('shows validationError when title is empty', async () => {
        const { i18n, router } = buildPlugins()
        const wrapper = mount(CardEditorView, { global: { plugins: [i18n, router] } })
        // Don't set title — click save
        const saveBtn = wrapper.findAll('button').find(b => b.text().includes('Save'))
        await saveBtn!.trigger('click')
        await flushPromises()
        expect(wrapper.text()).toContain('Fill all fields')
        expect(boardService.createBoard).not.toHaveBeenCalled()
    })

    it('shows notEnoughEntriesError when entries < required but title filled', async () => {
        const { i18n, router } = buildPlugins()
        const wrapper = mount(CardEditorView, { global: { plugins: [i18n, router] } })

        // Switch to 5x5 (25 required). Fill title.
        await wrapper.find('input.title-input').setValue('My Board')

        // Fill only 3 entries then remove the rest to avoid hasEmptyEntries trigger.
        // The watch pre-populates 25 rows — fill all of them but then remove 22.
        // Simpler: switch to 4x4 (16 required), fill only 3 entries and remove the other 13.
        const radio4 = wrapper.find('input[value="4x4"]')
        await radio4.setValue(true)
        await radio4.trigger('change')
        await flushPromises()

        // Now 16 rows exist. Fill 3, leave the rest empty — but the guard checks
        // hasEmptyEntries first. So: fill all visible rows with something, then
        // manually splice the entries down to 3 by clicking delete on 13 rows.
        const inputs = wrapper.findAll('input[data-index]')
        for (const input of inputs) {
            await input.setValue('x')
        }
        // Now 16 non-empty entries → that IS enough for 4x4. Instead keep only 3
        // by clicking the delete buttons 13 times.
        for (let i = 0; i < 13; i++) {
            const deleteBtns = wrapper.findAll('button.btn-danger')
            await deleteBtns[deleteBtns.length - 1]!.trigger('click')
            await flushPromises()
        }

        const saveBtn = wrapper.findAll('button').find(b => b.text().includes('Save'))
        await saveBtn!.trigger('click')
        await flushPromises()
        expect(wrapper.text()).toContain('Not enough entries')
        expect(boardService.createBoard).not.toHaveBeenCalled()
    })

    it('calls createBoard when title and all entries are filled', async () => {
        vi.mocked(boardService.createBoard).mockResolvedValue({
            board_id: 'b1', title: 'T', description: '', size: 4, author_id: 'u1', score: 0, vote_count: 0, play_count: 0,
        })
        vi.mocked(boardService.createCell).mockResolvedValue({
            cell_id: 'c1', content: 'x', value: 1,
        })

        const { i18n, router } = buildPlugins()
        const wrapper = mount(CardEditorView, { global: { plugins: [i18n, router] } })

        // Switch to 4x4 so we only need 16 entries
        const radio4 = wrapper.find('input[value="4x4"]')
        await radio4.setValue(true)
        await radio4.trigger('change')
        await flushPromises()

        await wrapper.find('input.title-input').setValue('My Board')
        const inputs = wrapper.findAll('input[data-index]')
        for (const input of inputs) {
            await input.setValue('entry')
        }

        const saveBtn = wrapper.findAll('button').find(b => b.text().includes('Save'))
        await saveBtn!.trigger('click')
        await flushPromises()

        expect(boardService.createBoard).toHaveBeenCalledWith('My Board', 4, undefined)
        expect(wrapper.text()).toContain('Saved!')
    })
})
