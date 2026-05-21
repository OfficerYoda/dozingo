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
            <div class="layout-options">
              <button class="btn btn-primary">3x3</button>
              <button class="btn btn-secondary">4x4</button>
              <button class="btn btn-secondary">5x5</button>
            </div>
          </div>
        </div>

        <div class="secondrow mb-3">
          <h3>Description</h3>
          <input type="text" class="inputfield describtion-input"
            placeholder="Briefly describe the theme of this bingo session">
        </div>

        <div class="thirdrow">
          <div>
          <h3 class="mb-0">Session entries</h3>
          <small class="additional-info-cells">8 entries added. You need 17 more for a 5x5 layout.</small>
          </div>
          <button class="btn btn-secondary add-row-button">
            <CirclePlus :size="20" />
            <p class="mb-0">Add new row</p>
          </button>
        </div>
      </article>
    </div>
  </section>
</template>

<script setup lang="ts">
import { } from 'lucide-vue-next';
import { FilePlus, CirclePlus } from 'lucide-vue-next';
</script>

<style scoped>
.header {
  display: flex;
  flex-direction: row;
  align-items: center;
  gap: 10px;
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

.describtion-input {
  width: 100%;
  height: 80px;
}

.add-row-button {
  display: flex;
  flex-direction: row;
  gap: 8px;
  align-items: center;
}

.thirdrow {
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