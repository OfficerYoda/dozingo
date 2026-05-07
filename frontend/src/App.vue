<template>
  <header>
    <h1><!-- Page Title --></h1>
    <!-- Sidebar open btn later on -->
  </header>

  <aside>
    <div className="sidebar-header">
      <span>DOZINGO</span>
      <!-- Sidebar close btn later on -->
    </div>
    <div className="sidebar-content">
      <nav>
        <ul>
          <li><RouterLink to="/">Home</RouterLink></li>
          <li><RouterLink to="/about">Kategorien</RouterLink></li>
        </ul>
      </nav>
    </div>
  </aside>

  <main>
    <RouterView />
  </main>

  <footer>
    &#169 DOZINGO - Alle Rechte vorbehalten
  </footer>
</template>

<script setup lang="ts">

</script>

<style scoped></style>