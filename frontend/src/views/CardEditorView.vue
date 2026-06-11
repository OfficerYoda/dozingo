<template>
  <section>
    <div class="container">
      <article class="card">
        <div class="card-header mb-3">
          <FilePlus :size="27" />
          <h2 class="mb-0">{{ $t('cardEditor.title') }}</h2>
        </div>

        <div class="form-top mb-3">
          <div class="form-title">
            <h3>{{ $t('cardEditor.sessionTitle') }}</h3>
            <input type="text" class="form-input title-input" :class="{ 'input-error': submitted && !title.trim() }" v-model="title" :placeholder="$t('cardEditor.titlePlaceholder')">
          </div>

          <div class="form-boardsize">
            <h3>{{ $t('cardEditor.layoutSelection') }}</h3>

            <div role="radiogroup" class="layout-options">
              <label :class="['btn', selectedSize === '4x4' ? 'btn-primary' : 'btn-secondary']">
                <input type="radio" name="boardsize" value="4x4" v-model="selectedSize" class="sr-only" />
                4x4
              </label>
              <label :class="['btn', selectedSize === '5x5' ? 'btn-primary' : 'btn-secondary']">
                <input type="radio" name="boardsize" value="5x5" v-model="selectedSize" class="sr-only" />
                5x5
              </label>
              <label :class="['btn', selectedSize === '6x6' ? 'btn-primary' : 'btn-secondary']">
                <input type="radio" name="boardsize" value="6x6" v-model="selectedSize" class="sr-only" />
                6x6
              </label>
            </div>
          </div>
        </div>

        <div class="form-description mb-3">
          <h3>{{ $t('cardEditor.description') }}</h3>
          <textarea rows="3" class="form-input description-input" v-model="description"
            :placeholder="$t('cardEditor.descriptionPlaceholder')"></textarea>
        </div>

        <div class="form-entries mb-3">
          <div class="entries-header mb-3">
            <div>
              <h3 class="mb-0">{{ $t('cardEditor.entries') }}</h3>
              <small :class="['entry-count-hint', entryHintReady ? 'hint-ready' : 'hint-warn']">{{ entryHintText }}</small>
            </div>
            <button class="btn btn-secondary add-entry-btn" @click="addRow">
              <CirclePlus :size="20" />
              <p class="mb-0">{{ $t('cardEditor.addRow') }}</p>
            </button>
          </div>

          <table class="entries-table">
            <thead>
              <tr class="table-header">
                <th class="col-term">{{ $t('cardEditor.termColumn') }}</th>
                <th class="col-rarity">{{ $t('cardEditor.rarityColumn') }}</th>
                <th class="col-delete"></th>
              </tr>
            </thead>

            <tbody>
              <tr v-for="(entry, index) in entries" :key="index">
                <td class="td-term">
                  <input type="text" :class="['entry-term-input', submitted && !entry.term.trim() ? 'input-error' : '']" v-model="entry.term"
                    :placeholder="$t('cardEditor.termPlaceholder')" />
                </td>
                <td class="td-rarity">
                  <select class="entry-select" v-model="entry.rarity">
                    <option value="common">{{ $t('cardEditor.rarity.common') }}</option>
                    <option value="uncommon">{{ $t('cardEditor.rarity.uncommon') }}</option>
                    <option value="rare">{{ $t('cardEditor.rarity.rare') }}</option>
                    <option value="legendary">{{ $t('cardEditor.rarity.legendary') }}</option>
                  </select>
                </td>
                <td class="td-delete">
                  <button class="btn btn-danger btn-icon" @click="removeRow(index)">
                    <Trash2 :size="16" />
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <div v-if="saveSuccess" class="alert alert-success mt-3">
          {{ $t('cardEditor.saveSuccess') }}
        </div>

        <div v-if="saveError" class="alert alert-danger mt-3">
          {{ $t('cardEditor.saveError') }}
        </div>

        <div v-if="validationError" class="alert alert-danger mt-3">
          {{ $t('cardEditor.validationError') }}
        </div>

        <div v-if="notEnoughEntriesError" class="alert alert-danger mt-3">
          {{ $t('cardEditor.notEnoughEntriesError') }}
        </div>

        <div class="save-btn-row">
          <button type="submit" class="btn btn-primary" @click="saveBoard">{{ $t('cardEditor.save') }}</button>
        </div>
      </article>
    </div>
  </section>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue';
import { useI18n } from 'vue-i18n';
import { FilePlus, CirclePlus, Trash2 } from 'lucide-vue-next';
import { usePageTitle } from '@/composables/usePageTitle'

useI18n()

interface Entry {
  term: string
  rarity: string
}

const title = ref('')
const description = ref('')
const saveSuccess = ref(false)
const saveError = ref(false)
const validationError = ref(false)
const notEnoughEntriesError = ref(false)
const submitted = ref(false)
const entries = ref<Entry[]>([{ term: '', rarity: 'common' }, { term: '', rarity: 'common' }, { term: '', rarity: 'common' }])
const selectedSize = ref('5x5')

const { t } = useI18n()

const { pageTitle } = usePageTitle(t('header.cardEditor'))

const requiredEntries = computed(() => {
  const n = parseInt(selectedSize.value)
  return n * n
})

const entryHintReady = computed(() => entries.value.filter(e => e.term.trim()).length >= requiredEntries.value)

const entryHintText = computed(() => {
  const count = entries.value.filter(e => e.term.trim()).length
  const size = selectedSize.value
  if (entryHintReady.value) {
    return t('cardEditor.entryCountHintReady', { count, size })
  }
  return t('cardEditor.entryCountHint', { count, needed: requiredEntries.value - count, size })
})

const rarityValue: Record<string, number> = { common: 1, uncommon: 2, rare: 3, legendary: 4 }

function addRow() {
  entries.value.push({ term: '', rarity: 'common' })
}

function removeRow(index: number) {
  entries.value.splice(index, 1)
}

async function saveBoard() {
  submitted.value = true
  saveSuccess.value = false
  saveError.value = false
  validationError.value = false
  notEnoughEntriesError.value = false

  const titleEmpty = !title.value.trim()
  const hasEmptyEntries = entries.value.some(e => !e.term.trim())
  const notEnoughEntries = entries.value.filter(e => e.term.trim()).length < requiredEntries.value

  if (titleEmpty || hasEmptyEntries) {
    validationError.value = true
    return
  }
  if (notEnoughEntries) {
    notEnoughEntriesError.value = true
    return
  }

  const size = parseInt(selectedSize.value)

  const boardRes = await fetch('/api/boards', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    credentials: 'include',
    body: JSON.stringify({ title: title.value, description: description.value || undefined, size }),
  })
  if (!boardRes.ok) {
    saveError.value = true
    saveSuccess.value = false
    return
  }
  const board = await boardRes.json()

  await Promise.all(
    entries.value
      .filter(e => e.term.trim())
      .map(e =>
        fetch(`/api/boards/${board.board_id}/cells`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          credentials: 'include',
          body: JSON.stringify({ content: e.term, value: rarityValue[e.rarity] }),
        })
      )
  )

  console.log('Board created:', board.board_id)
  saveSuccess.value = true
}
</script>

<style scoped>
/* === Layout === */

.card-header {
  display: flex;
  align-items: center;
  gap: 10px;
}

.form-top {
  display: flex;
  flex-wrap: wrap;
}

.form-title {
  width: 50%;
  padding-right: 5px;
}

.form-boardsize {
  width: 50%;
  padding-left: 5px;
}

.entries-header {
  display: flex;
  justify-content: space-between;
}

.save-btn-row {
  display: flex;
  justify-content: center;
}

/* === Typography === */

h2 {
  color: var(--color-heading);
}

h3 {
  color: var(--color-subheading);
  font-weight: 300;
  font-size: 1.25rem;
}

.card-header svg {
  color: var(--color-heading);
}

/* === Form inputs === */

.form-input {
  border-color: transparent;
  border-radius: 8px;
  background-color: var(--color-input-bg);
}

.title-input {
  width: 100%;
  height: 40px;
}

.description-input {
  width: 100%;
  height: 80px;
  resize: vertical;
}

/* === Layout options (board size) === */

.layout-options {
  display: flex;
  gap: 10px;
}

.layout-options .btn {
  flex: 1;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
}

.sr-only {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border: 0;
}

/* === Entries table === */

.entries-table {
  width: 100%;
  border-collapse: separate;
  border-spacing: 0;
  border: 1px solid var(--color-text-secondary);
  border-radius: var(--radius-sm);
  overflow: hidden;
}

.entries-table th {
  color: var(--color-subheading);
  font-weight: 350;
  padding: 10px;
}

.table-header th {
  background-color: var(--color-input-bg);
}

.table-header th:first-child {
  border-top-left-radius: var(--radius-sm);
}

.table-header th:last-child {
  border-top-right-radius: var(--radius-sm);
}

.entries-table tbody td {
  border-top: 1px solid var(--color-text-secondary);
}

.col-term {
  width: 60%;
}

.col-rarity {
  width: 20%;
}

.col-delete {
  width: 5%;
  text-align: center;
}

.td-term {
  padding: 5px;
}

.td-rarity {
  padding: 0;
  height: 1px;
}

.td-delete {
  text-align: center;
}

.entry-term-input {
  width: 100%;
  border-color: transparent;
  background-color: transparent;
}

.entry-select {
  width: 100%;
  height: 100%;
  background-color: transparent;
  border-top: none;
  border-bottom: none;
  border-left: 1px solid var(--color-text-secondary);
  border-right: 1px solid var(--color-text-secondary);
  padding-inline: 8px;
  padding-right: 28px;
  color: var(--color-subheading);
  appearance: none;
  background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='12' height='12' viewBox='0 0 24 24' fill='none' stroke='%23888' stroke-width='2' stroke-linecap='round' stroke-linejoin='round'%3E%3Cpolyline points='6 9 12 15 18 9'%3E%3C/polyline%3E%3C/svg%3E");
  background-repeat: no-repeat;
  background-position: right 8px center;
}

/* === Buttons === */

.add-entry-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  height: 40px;
}

/* === Other === */

.entry-count-hint {
  font-weight: 500;
}

.hint-warn {
  color: var(--color-accent-red);
}

.hint-ready {
  color: var(--color-accent-green, green);
}

.input-error {
  border-color: var(--color-accent-red) !important;
  background-color: #fff0f0;
}

.card {
  background-color: var(--color-bg-card-tinted);
}
</style>
