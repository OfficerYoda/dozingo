<template>
  <section>
    <div class="container">
      <article className="card">
        <div class="header mb-3">
          <FilePlus :size="27" />
          <h2 class="mb-0">Create a new board</h2>
        </div>

        <div class="firstrow mb-3">
          <div class="title">
            <h3>Session title</h3>
            <input type="text" class="inputfield inputfield-title" placeholder="e.g. Intoduction to Game Theory">
          </div>

          <div class="boardsize">
            <h3>Layout selection</h3>
            <div role="radiogroup" class="layout-options">
              <button :class="['btn', selectedSize === '3x3' ? 'btn-primary' : 'btn-secondary']" @click="selectedSize = '3x3'">3x3</button>
              <button :class="['btn', selectedSize === '4x4' ? 'btn-primary' : 'btn-secondary']" @click="selectedSize = '4x4'">4x4</button>
              <button :class="['btn', selectedSize === '5x5' ? 'btn-primary' : 'btn-secondary']" @click="selectedSize = '5x5'">5x5</button>
            </div>
          </div>
        </div>

        <div class="secondrow mb-3">
          <h3>Description</h3>
          <input type="text" class="inputfield describtion-input"
            placeholder="Briefly describe the theme of this bingo session">
        </div>

        <div class="thirdrow mb-3">
          <div class="headerthirdrow mb-3">
            <div>
              <h3 class="mb-0">Session entries</h3>
              <small class="additional-info-cells">8 entries added. You need 17 more for a 5x5 layout.</small>
            </div>
            <button class="btn btn-secondary add-row-button" @click="addRow">
              <CirclePlus :size="20" />
              <p class="mb-0">Add new row</p>
            </button>
          </div>

          <table class="entries-table">
            <thead>
              <tr class="tableheader">
                <th class="table-quote">Quote / Term</th>
                <th class="table-rarity">Rarity</th>
                <th class="table-delete">Löschen</th>
              </tr>
            </thead>

            <tbody>
              <tr v-for="(entry, index) in entries" :key="index">
                <td class="table-td-input">
                  <input type="text" class="entry-input" v-model="entry.term" placeholder="Enter quote or term" />
                </td>
                <td class="table-td-select">
                  <select class="selection-table entry-select" v-model="entry.rarity">
                    <option value="common">Common</option>
                    <option value="uncommon">Uncommon</option>
                    <option value="rare">Rare</option>
                  </select>
                </td>
                <td class="table-td-button">
                  <button class="btn-icon btn btn-danger" @click="removeRow(index)">
                    <Trash2 :size="16" />
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <div class="savebutton">
          <button class="btn btn-primary">Save board to your templates</button>
        </div>
      </article>
    </div>
  </section>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { FilePlus, CirclePlus, Trash2 } from 'lucide-vue-next';

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

.savebutton {
  display: flex;
  justify-content: center;
}

.header {
  display: flex;
  flex-direction: row;
  align-items: center;
  gap: 10px;
}

.entry-input {
  width: 100%;
  border-color: transparent;
}

.entry-select {
  width: 100%;
}

.table-quote {
  width: 60%;
}

.table-rarity {
  width: 20%;
}

.table-delete {
  width: 5%;
  text-align: center;
}

.table-td-button {
  text-align: center;
}

.entries-table tbody tr td {
  border-top: 1px solid #6B7280;
}

.table-td-input {
  padding: 5px;
}


.table-td-select {
  padding: 5px;
}

th {
  color: #5A5781;
  font-weight: 350;
  padding: 10px;
}

.tableheader th {
  background-color: #E3DFFF;
}

.tableheader th:first-child {
  border-top-left-radius: var(--radius-sm);
}

.tableheader th:last-child {
  border-top-right-radius: var(--radius-sm);
}


.entries-table {
  width: 100%;
  border-collapse: separate;
  border-spacing: 0;


  border: 1px solid #6B7280;
  border-radius: var(--radius-sm);
}

h2 {
  color: #2C2A51;
}

.header svg {
  color: #2C2A51;
}

h3 {
  color: #5A5781;
  font-weight: 300;
  font-size: 1.25rem;
}

.firstrow {
  display: flex;
  flex-direction: row;
  flex-wrap: wrap;
}

.layout-options {
  display: flex;
  gap: 10px;
}

.layout-options .btn {
  flex: 1;
}

.title {
  width: 50%;
  padding-right: 5px;
}

.boardsize {
  width: 50%;
  padding-left: 5px;
}

.inputfield-title {
  width: 100%;
  height: 40px;
}

.inputfield {
  border: transparent;
  border-radius: 8px;
  background-color: #E3DFFF;
}

.selection-table {
  border: transparent;
  border-radius: 8px;
}

.describtion-input {
  width: 100%;
  height: 80px;
}

.add-row-button {
  display: flex;
  flex-direction: row;
  gap: 8px;
  align-items: center;
  height: 40px;
}

.headerthirdrow {
  display: flex;
  justify-content: space-between;
}

input::-webkit-input-placeholder {
  position: absolute;
  left: 5px;
  top: 5px;
}

.additional-info-cells {
  color: var(--color-accent-red);
  font-weight: 500;
}
</style>