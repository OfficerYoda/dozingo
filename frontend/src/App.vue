<template>
  <div class="app-layout">
    <!-- REVIEW:
    header und main verwenden aktuell beide grid-row: 1. Dadurch überlagern sich die Bereiche im Grid. 
    Schaue dir über die DevTools die größe der Header und Main Boxen an um zu verstehen, was ich meine.
    Wenn du in <h1> einen Test Titel hinein schreibst, siehst du auch, dass das ganze nicht korrekt dargestellt wird.
    -->
    <header>
      <h1><!-- Page Title --></h1>
    </header>

    <aside>
      <div class="sidebar-header">
        <!-- REVIEW:
        Der Markentitel sollte kein <h_>-Tag sein, da er keine inhaltliche Seitenüberschrift darstellt, sondern eher ein Logo oder eine Markenidentität. Stattdessen könnte man besser einen span verwenden, um die gewünschte Schriftgröße und -gewicht zu erreichen, ohne die semantische Struktur der Seite zu beeinträchtigen.
        -->
        <h3>DOZINGO</h3>
        <small>Schoolary Playground</small>
      </div>

      <!-- REVIEW:
      small wird mehrfach für normalen Text genutzt.
      Semantisch ist <small> eher für Nebeninformationen gedacht.
      Besser entweder die globale Schriftgröße anpassen oder alternativ die Sidebar-Textgröße in einer CSS-Klasse ändern, um die Konsistenz zu wahren.
      -->
      <div class="sidebar-content">
        <nav>
          <ul>
            <li><RouterLink to="/" class="sidebar-buttons upper-buttons"><Home :size="20" /><small>Home</small></RouterLink></li>
            <li><RouterLink to="/about" class="sidebar-buttons upper-buttons"><LayoutGrid :size="20"/><small>About</small></RouterLink></li>
          </ul>
        </nav>
      </div>

      <!-- REVIEW:
      Eventuelle horizontale Trennung zwischen Hauptnavigation (Home, About) und Nutzeraktionen (Settings, Sign Out, Profile) könnte die Übersicht verbessern (<hr> mit Border-Anpassung). Aktuell sind alle Einträge ohne visuelle Unterscheidung in einer Liste. Eine klare Trennung oder Gruppierung könnte hier helfen. 
      Ist übrigens auch so mit <hr> in Figma bereits zu sehen.
      -->
      <div class="sidebar-footer">
        <nav>
          <ul>
            <li><RouterLink to="/settings" class="sidebar-buttons" ><Settings :size="20"/><small>Settings</small></RouterLink></li>
            <!-- REVIEW:
            Der "Sign Out"-Eintrag ist kein RouterLink und aktuell auch kein Button.
            Für Accessibility und Funktion besser als <button> oder <RouterLink> umsetzen, je nach Bedarf.
            -->
            <li class="sidebar-buttons"><LogOut :size="20" class="sidebar-footer-signout"/><small class="sidebar-footer-signout">Sign Out</small> </li>
            <li><RouterLink class="sidebar-buttons" to="/profile"><UserCircle :size="20"/><small>Profile</small></RouterLink></li>
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
  </div>
</template>

<script setup lang="ts">
import { Home, LayoutGrid, Settings, LogOut, UserCircle} from 'lucide-vue-next'
</script>

<style scoped>

.app-layout {
  display: grid;
  height: 100vh;
  grid-template-columns: 250px 1fr;
  grid-template-rows: auto 1fr auto;
}

nav li {
  list-style-type:none;
  color: var(--color-primary-600);
  font-size: 0.875rem;
  font-weight: 700;
}

.sidebar-header h3 {
  color: var(--color-primary-800);
  margin-bottom: 0;
}

header {
  grid-column: 2;
  grid-row: 1;
}

aside {
  padding: 20px;
  background-color: var(--color-bg-sidebar);
  grid-column: 1;
  grid-row: 1 / -1;
  display: grid;
  grid-template-rows: auto 1fr auto;
}

.sidebar-header {
  grid-row: 1;
  margin-bottom: 30px;
}

.sidebar-content {
  grid-row: 2;
}

.sidebar-footer {
  grid-row: 3;
}

main {
  grid-column: 2;
  grid-row: 1;
}

footer {
  grid-column: 2;
  grid-row: 3;
}

.sidebar-buttons {
  display: flex;
  flex-direction: row;
  align-items: center;
  gap: 8px;
  padding: 10px;
}

.sidebar-buttons small {
  font-size: medium;
  font-weight: 400;
}

.sidebar-footer-signout {
  color: var(--color-accent-red);
}

.router-link-exact-active.upper-buttons {
  background-color: var(--color-primary-200);
  border-radius: var(--radius-sm);
}

</style>