<template>
  <section>
    <div class="container">
      <div class="legal-wrapper">
        <h1>Datenschutzerklärung</h1>

        <div class="legal-section">
          <h2>1. Verantwortlicher</h2>
          <p>
            Dozingo-Team, Duale Hochschule Baden-Württemberg, Karlsruhe<br>
            E-Mail: <a href="mailto:kontakt@dozingo.de">kontakt@dozingo.de</a>
          </p>
        </div>

        <div class="legal-section">
          <h2>2. Erhobene Daten</h2>
          <p>
            Bei der Nutzung von Dozingo verarbeiten wir folgende personenbezogene Daten:
          </p>
          <ul>
            <li>E-Mail-Adresse und Benutzername bei der Registrierung</li>
            <li>Spielstände und erstellte Boards</li>
            <li>Technische Verbindungsdaten (IP-Adresse, Browser-Typ) in Server-Logs</li>
          </ul>
        </div>

        <div class="legal-section">
          <h2>3. Zweck der Verarbeitung</h2>
          <p>
            Ihre Daten werden ausschließlich zur Bereitstellung der Plattform-Funktionalität
            verwendet: Nutzerauthentifizierung, Speicherung von Spielständen und
            Verwaltung von Boards.
          </p>
        </div>

        <div class="legal-section">
          <h2>4. Cookies & Sessions</h2>
          <p>
            Dozingo verwendet Session-Cookies zur Authentifizierung. Diese Cookies sind
            technisch notwendig und werden nach dem Abmelden gelöscht. Es werden keine
            Tracking- oder Marketing-Cookies eingesetzt.
          </p>
        </div>

        <div class="legal-section">
          <h2>5. Datenweitergabe</h2>
          <p>
            Ihre personenbezogenen Daten werden nicht an Dritte weitergegeben.
            Eine Übermittlung in Drittländer findet nicht statt.
          </p>
        </div>

        <div class="legal-section">
          <h2>6. Ihre Rechte</h2>
          <p>
            Sie haben das Recht auf Auskunft, Berichtigung, Löschung und Einschränkung
            der Verarbeitung Ihrer Daten sowie das Recht auf Datenübertragbarkeit.
            Zur Ausübung Ihrer Rechte wenden Sie sich bitte an:
            <a href="mailto:kontakt@dozingo.de">kontakt@dozingo.de</a>
          </p>
        </div>

        <div class="legal-section">
          <h2>7. Speicherdauer</h2>
          <p>
            Kontodaten werden gespeichert, solange ein aktives Konto besteht.
            Nach Löschung des Kontos werden alle personenbezogenen Daten
            innerhalb von 30 Tagen entfernt.
          </p>
        </div>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { usePageTitle } from '@/composables/usePageTitle'
usePageTitle('Datenschutz')
</script>

<style scoped>
.legal-wrapper {
  max-width: 720px;
  padding: 40px 0 60px;
}

.legal-wrapper h1 {
  font-size: 2rem;
  font-weight: 800;
  color: var(--color-heading);
  margin-bottom: 32px;
}

.legal-section {
  margin-bottom: 28px;
}

.legal-section h2 {
  font-size: 1rem;
  font-weight: 700;
  color: var(--color-subheading);
  text-transform: uppercase;
  letter-spacing: 0.05em;
  margin-bottom: 8px;
}

.legal-section p,
.legal-section ul {
  color: var(--color-text-secondary);
  line-height: 1.7;
  margin: 0;
}

.legal-section ul {
  padding-left: 20px;
  margin-top: 6px;
}

.legal-section li {
  margin-bottom: 4px;
}

.legal-section a {
  color: var(--color-primary-600);
}
</style>
