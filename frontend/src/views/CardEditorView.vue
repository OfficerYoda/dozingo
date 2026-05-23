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
            <input type="text" class="form-input title-input" :placeholder="$t('cardEditor.titlePlaceholder')">
          </div>

          <div class="form-boardsize">
            <h3>{{ $t('cardEditor.layoutSelection') }}</h3>
            <div role="radiogroup" class="layout-options">
              <button :class="['btn', selectedSize === '3x3' ? 'btn-primary' : 'btn-secondary']" @click="selectedSize = '3x3'">3x3</button>
              <button :class="['btn', selectedSize === '4x4' ? 'btn-primary' : 'btn-secondary']" @click="selectedSize = '4x4'">4x4</button>
              <button :class="['btn', selectedSize === '5x5' ? 'btn-primary' : 'btn-secondary']" @click="selectedSize = '5x5'">5x5</button>
            </div>
          </div>
        </div>

        <div class="form-description mb-3">
          <h3>{{ $t('cardEditor.description') }}</h3>
          <textarea rows="3" class="form-input description-input" :placeholder="$t('cardEditor.descriptionPlaceholder')"></textarea>
        </div>

        <div class="form-entries mb-3">
          <div class="entries-header mb-3">
            <div>
              <h3 class="mb-0">{{ $t('cardEditor.entries') }}</h3>
              <small class="entry-count-hint">8 entries added. You need 17 more for a 5x5 layout.(Nicht übersetzt bzw. keine dynamische Anpassung)</small>
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
                  <input type="text" class="entry-term-input" v-model="entry.term" :placeholder="$t('cardEditor.termPlaceholder')" />
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

        <div class="save-btn-row">
          <button type="submit" class="btn btn-primary">{{ $t('cardEditor.save') }}</button>
        </div>
      </article>
    </div>
  </section>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { FilePlus, CirclePlus, Trash2 } from 'lucide-vue-next';

useI18n()

interface Entry {
  term: string
  rarity: string
}

const entries = ref<Entry[]>([{ term: '', rarity: 'common' }, { term: '', rarity: 'common' }, { term: '', rarity: 'common' }])
const selectedSize = ref('3x3')

function addRow() {
  entries.value.push({ term: '', rarity: 'common' })
}

function removeRow(index: number) {
  entries.value.splice(index, 1)
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
}

/* === Entries table === */

.entries-table {
  width: 100%;
  border-collapse: separate;
  border-spacing: 0;
  border: 1px solid var(--color-text-secondary);
  border-radius: var(--radius-sm);
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

.col-term  { width: 60%; }
.col-rarity { width: 20%; }
.col-delete { width: 5%; text-align: center; }

.td-term   { padding: 5px; }
.td-rarity { padding: 5px; }
.td-delete { text-align: center; }

.entry-term-input {
  width: 100%;
  border-color: transparent;
  background-color: transparent;
}

.entry-select {
  width: 100%;
  border-color: transparent;
  border-radius: 8px;
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
  color: var(--color-accent-red);
  font-weight: 500;
}

.card {
  background-color: var(--color-bg-card-tinted);
}
</style>
