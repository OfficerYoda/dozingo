<template>
  <section>
    <div class="container">
      <article className="card">
        <div class="header">
          <FilePlus :size="27" />
          <h2 class="mb-0">Create a new board</h2>
        </div>

        <div class="firstrow">
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

        <div class="secondrow">
          <h3>Description</h3>
          <input type="text" class="inputfield describtion-input" placeholder="Briefly describe the theme of this bingo session">
        </div>

        <div class="thirdrow">
          <h3>Session entries</h3>
          <small>8 entries added. You need 17 more for a 5x5 layout.</small>
          <button>
            <CirclePlus :size="20" />
            <p>Add new row</p>
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

.firstrow {
  display: flex;
  flex-direction: row;
  flex-wrap: wrap;
}

.layout-options {
  display: flex;
  gap: 8px;
}

.layout-options .btn {
  flex: 1;
}


.title {
    width: 50%;
}

.boardsize {
  width: 50%;
}

.inputfield-title{
  width: 100%;
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

input::-webkit-input-placeholder { 
  position:absolute;
  left: 5px;
  top: 5px;
}

</style>