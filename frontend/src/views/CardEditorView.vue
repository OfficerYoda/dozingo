<template>
  <section>
    <div class="grid container">
      <article class="col-7 md-12">
        <div class="card">
          <h2>Forge a new Batch</h2>
          <div>
            <span>Select Grid Layout</span>
            <div>
              <button></button>
              <button></button>
              <button></button>
            </div>
          </div>


          <div>
            <div>
              <span>Choose Dozent</span>
              <div class="dropdown">
                <button href="" className="btn btn-secondary dropdown-button">Dozent</button>
                <ul class="dropdown-menu">
                  <li>Kollege Röthig</li>
                  <li>Rotzinger</li>
                  <li>usw</li>
                </ul>
              </div>
            </div>
            <div>
              <span>Context details</span>
              <input type="text">
            </div>
            <div>
              <span>Difficulty</span>
              <div class="dropdown">
                <button href="" className="btn btn-secondary dropdown-button">Choose Dificulty</button>
                <ul class="dropdown-menu">
                  <li>Einfach</li>
                  <li>Mittel</li>
                  <li>Schwer</li>
                </ul>
              </div>
            </div>
          </div>

          <button>Submit for batch review</button>
        </div>
      </article>

      <div class="col-5">
        <article>
          <div class="card">
            <h3>Live Grid Content</h3>
            <small>Hier Grid Vorschau einbauen</small>
            <small>Previewing your batch inside a 5x5 campus grid. Layout scale adjusts to grid complexity.</small>
          </div>
        </article>

        <article>
          <div class="card">
            <h3>Submission Code</h3>
            <li>
              <ul>Batch entries must follow the same category theme.</ul>
              <ul>Duplicates in a single batch will be automatically merged.</ul>
              <ul>Ensure the difficulty rating reflects the average rarity of the batch.</ul>
            </li>
          </div>
        </article>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { } from 'lucide-vue-next';
</script>

<style scoped></style>