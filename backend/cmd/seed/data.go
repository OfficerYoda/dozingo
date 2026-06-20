package main

// Seed credentials (for local dev only):
//   maxmustermann      / password123
//   lena.schmidt       / securePass!
//   timoWerner42       / timoSecret42
//   ghostUser01        / ghostpass
//   anon_student       / anonpass
//   admin              / password123      (easy test admin)
//
// Seed session tokens (for local dev only):
//   seed-anon-fresh-token-0001      (anonymous, ~30d valid)
//   seed-anon-near-expiry-0002      (anonymous, <7d -> triggers extension)
//   seed-anon-expired-0003          (anonymous, expired -> cleanup target)
//   seed-user-max-token-0010        (bound to maxmustermann)
//   seed-user-ghost-token-0011      (bound to ghostUser01)
//   seed-user-admin-token-0012      (bound to admin)

// userData holds username and email pairs for seeding users.
type userData struct {
	Username string
	Email    string
}

// boardData holds the definition of a board to seed.
// AuthorIdx refers to an index in the users slice.
type boardData struct {
	Title       string
	Description string
	Size        int32
	AuthorIdx   int
}

// voteData holds a single vote to seed.
// UserIdx and BoardIdx refer to indices in the users/boards slices.
type voteData struct {
	UserIdx  int
	BoardIdx int
	Value    int32 // 1 or -1
}

// gameData holds a game to seed.
// PlayerIdx and BoardIdx refer to indices in the users/boards slices.
// PlayerIdx == -1 means the game is anonymous (no logged-in user) and is
// authorised via the session referenced by SessionIdx instead.
// SessionIdx == -1 means the game has no associated session (the classic
// pre-session, user-only flow); when both are set, the game has both a
// player_id and a session_id (a logged-in user playing from a session that
// was minted while they were anonymous).
// gameSessionSeed describes one play session for a seeded game.
// StartedAgoMinutes controls when the session began relative to now();
// DurationMinutes controls how long the player was active.
type gameSessionSeed struct {
	StartedAgoMinutes int
	DurationMinutes   int
}

type gameData struct {
	PlayerIdx  int
	SessionIdx int
	BoardIdx   int
	Status     string            // "active" or "abandoned"
	Sessions   []gameSessionSeed // one or more play sessions
}

// gameCellData holds a game cell to seed.
// Position is 0-indexed (row-major order).
type gameCellData struct {
	Content  string
	Position int32
	IsMarked bool
}

// --- Seed Data ---

var users = []userData{
	{Username: "maxmustermann", Email: "max.mustermann@stud.uni-heidelberg.de"},
	{Username: "lena.schmidt", Email: "lena.schmidt@stud.uni-muenchen.de"},
	{Username: "timoWerner42", Email: "timo.werner@stud.kit.edu"},
	{Username: "sophiebraun", Email: "sophie.braun@stud.tu-berlin.de"},
	{Username: "jannik_f", Email: "jannik.fischer@stud.uni-freiburg.de"},
	{Username: "klaraM", Email: "klara.mueller@stud.rwth-aachen.de"},
	{Username: "felixHahn", Email: "felix.hahn@stud.uni-stuttgart.de"},
	{Username: "annaK99", Email: "anna.koch@stud.fu-berlin.de"},
	{Username: "lukasBauer", Email: "lukas.bauer@stud.tu-darmstadt.de"},
	{Username: "emiliaWolf", Email: "emilia.wolf@stud.uni-koeln.de"},
	{Username: "ghostUser01", Email: "ghost.user01@mail.dozingo.de"},
	{Username: "anon_student", Email: "anon.student@mail.dozingo.de"},
	{Username: "admin", Email: "admin@mail.dozingo.de"},
}

// boards defines 18 boards with mixed sizes (4-6), each linked to an author.
// Size distribution: 4 boards size 4, 10 boards size 5, 4 boards size 6
// (~22% / 56% / 22%, approximating a 20/60/20 split).
var boards = []boardData{
	{Title: "Mathe 1 Bingo", Description: "Klassiker aus der Mathe 1 Vorlesung", Size: 4, AuthorIdx: 0},          // idx 0  (size 4)
	{Title: "Lineare Algebra Klassiker", Description: "Die besten Momente aus LinAlg", Size: 5, AuthorIdx: 1},    // idx 1  (size 5)
	{Title: "Statistik Grundlagen", Description: "Statistik Vorlesung Bingo", Size: 5, AuthorIdx: 4},             // idx 2  (size 5)
	{Title: "Theoretische Informatik", Description: "Turingmaschinen und mehr", Size: 4, AuthorIdx: 2},           // idx 3  (size 4)
	{Title: "Physik Vorlesung", Description: "Experimente die schiefgehen", Size: 5, AuthorIdx: 3},               // idx 4  (size 5)
	{Title: "BWL Einführung", Description: "PowerPoint-Schlachten und Buzzwords", Size: 5, AuthorIdx: 5},         // idx 5  (size 5)
	{Title: "Algorithmen und Datenstrukturen", Description: "O-Notation und Rekursion", Size: 5, AuthorIdx: 6},   // idx 6  (size 5)
	{Title: "Softwaretechnik Bingo", Description: "UML und Agile Methoden", Size: 5, AuthorIdx: 7},               // idx 7  (size 5)
	{Title: "Datenbanken Vorlesung", Description: "SQL, Normalformen und Joins", Size: 5, AuthorIdx: 0},          // idx 8  (size 5)
	{Title: "Betriebssysteme Chaos", Description: "Deadlocks und Race Conditions", Size: 6, AuthorIdx: 8},        // idx 9  (size 6)
	{Title: "Rechnernetze Marathon", Description: "OSI-Modell und TCP/IP", Size: 6, AuthorIdx: 9},                // idx 10 (size 6)
	{Title: "Digitaltechnik Vorlesung", Description: "Flip-Flops und Gatter", Size: 6, AuthorIdx: 1},             // idx 11 (size 6)
	{Title: "Compilerbau Endgegner", Description: "Parser, Lexer und Optimierung", Size: 6, AuthorIdx: 3},        // idx 12 (size 6)
	{Title: "Verteilte Systeme Bingo", Description: "CAP-Theorem und Konsens", Size: 5, AuthorIdx: 6},            // idx 13 (size 5)
	{Title: "Maschinelles Lernen Bingo", Description: "Gradient Descent und Overfitting", Size: 5, AuthorIdx: 2}, // idx 14 (size 5)

	// Admin-authored boards (AuthorIdx: 12)
	{Title: "Programmierkurs Klassiker", Description: "Live-Coding-Fails und Stack Traces", Size: 4, AuthorIdx: 12},    // idx 15 (size 4)
	{Title: "Webtechnologien Bingo", Description: "HTTP, REST und CSS-Spezifität", Size: 5, AuthorIdx: 12},             // idx 16 (size 5)
	{Title: "Admin Debug Bingo", Description: "QA hits, prod fires, and questionable commits", Size: 4, AuthorIdx: 12}, // idx 17 (size 4)
}

// cellPhrases contains German lecture bingo phrases for each board index.
// Pool sizes: size 4 -> 21, size 5 -> 32, size 6 -> 45
var cellPhrases = map[int][]string{
	// Board 0: "Mathe 1 Bingo" (size 4, 21 cells)
	0: {
		"Prof sagt 'trivial'",
		"Beweis wird übersprungen",
		"'Das sieht man sofort'",
		"Kreide bricht ab",
		"Tafel ist voll, Prof wischt hektisch",
		"Jemand fragt 'Kommt das in der Klausur?'",
		"Prof rechnet sich an der Tafel vor",
		"Epsilon-Delta taucht auf",
		"'Übung für den Leser'",
		"Prof verwechselt Plus und Minus",
		"Studenten schauen verwirrt",
		"'Das hatten wir schon letzte Woche'",
		"Tippfehler auf der Folie",
		"Prof sucht den Tafellappen",
		"Hörsaal-Mikrofon rauscht",
		"Tutor wird widersprochen",
		"'Mit vollständiger Induktion'",
		"Beamer flackert kurz",
		"Prof zeichnet Cantor-Mengen",
		"Grenzwert wird hingeschrieben",
		"'Die Klausur wird fair'",
	},

	// Board 1: "Lineare Algebra Klassiker" (size 5, 32 cells)
	1: {
		"Matrix wird falsch multipliziert",
		"'Der Beweis ist elegant'",
		"Determinante vergessen",
		"Prof zeichnet einen Vektorraum",
		"Eigenwert-Witz",
		"Jemand schläft ein",
		"Prof sagt 'offensichtlich'",
		"Rang der Matrix unklar",
		"'Das ist ein Spezialfall'",
		"Gauß-Algorithms zum dritten Mal",
		"Prof dreht sich zur falschen Tafel",
		"Inverse existiert nicht",
		"Basiswechsel verwirrt alle",
		"Skalarprodukt falsch berechnet",
		"'Diagonalisierbar oder nicht?'",
		"Prof verliert die Kreide",
		"Jordan-Normalform erwähnt",
		"Orthogonale Projection an der Tafel",
		"'Das ist linear unabhängig'",
		"Prof vergisst Index",
		"Kern und Bild verwechselt",
		"Eigenvektor wird berechnet",
		"Spur einer Matrix erwähnt",
		"Lineare Abbildung gezeichnet",
		"Cramersche Regel angewendet",
		"Prof verwechselt Zeile und Spalte",
		"Householder-Spiegelung erwähnt",
		"Untervektorraum-Beispiel",
		"Prof zeichnet R³-Koordinaten",
		"Singulärwertzerlegung kurz angerissen",
		"'Das ist eine Bilinearform'",
		"Prof rechnet Kreuzprodukt vor",
	},

	// Board 2: "Statistik Grundlagen" (size 5, 32 cells)
	2: {
		"Normalverteilung wird gezeichnet",
		"'Korrelation ist nicht Kausalität'",
		"p-Wert wird falsch interpretiert",
		"Würfelbeispiel",
		"Prof zeigt Excel-Tabelle",
		"Standardabweichung vergessen",
		"Konfidenzintervall erklärt",
		"Jemand fragt nach der Formelsammlung",
		"Histogramm an der Tafel",
		"'In der Praxis sieht das anders aus'",
		"Bayes wird erwähnt",
		"Prof sagt 'signifikant'",
		"Hypothesentest formuliert",
		"Zentraler Grenzwertsatz zitiert",
		"Boxplot wird gezeichnet",
		"'Ausreißer ignorieren wir'",
		"Prof verwechselt Median und Mittelwert",
		"Stichprobenumfang zu klein",
		"Chi-Quadrat-Test erwähnt",
		"Streudiagramm auf der Folie",
		"'Das ist nur deskriptive Statistik'",
		"Kovarianzmatrix an der Tafel",
		"Schiefe und Kurtosis erwähnt",
		"Lineare Regression gezeichnet",
		"Prof zeigt R-Output",
		"Dichtefunktion skizziert",
		"'Das ist nur ein Schätzer'",
		"Bootstrap-Verfahren genannt",
		"Poissonverteilung als Beispiel",
		"Prof verwechselt Varianz und Standardabweichung",
		"Q-Q-Plot erwähnt",
		"'Statistik lügt nicht, Statistiker schon'",
	},

	// Board 3: "Theoretische Informatik" (size 4, 21 cells)
	3: {
		"Turingmaschine wird gezeichnet",
		"'Das ist unentscheidbar'",
		"Regulärer Ausdruck wird kompliziert",
		"Pumping Lemma Beweis",
		"Prof sagt 'Nichtdeterminismus'",
		"Automat hat zu viele Zustände",
		"Jemand fragt 'Wozu braucht man das?'",
		"Chomsky-Hierarchie an der Tafel",
		"Prof vergisst Endzustand",
		"'Das reduzieren wir auf das Halteproblem'",
		"Komplexitätsklasse P vs NP",
		"Beweis durch Widerspruch",
		"Prof schreibt Produktionsregel falsch",
		"Kellerautomat wird eingeführt",
		"Jemand sagt 'Das ist doch wie Mathe'",
		"Sprache wird nicht akzeptiert",
		"Prof zeichnet Übergangsfunktion",
		"'Das ist klausurrelevant'",
		"Diagonalisierungsargument",
		"Äquivalenzklassen werden gebildet",
		"Prof löscht die Tafel zu früh",
	},

	// Board 4: "Physik Vorlesung" (size 5, 32 cells)
	4: {
		"Prof lässt Experiment fallen",
		"'In der Realität vernachlässigen wir Reibung'",
		"Einheit vergessen",
		"Vektorpfeil fehlt",
		"Prof rechnet im Kopf falsch",
		"Newton wird zitiert",
		"'Das ist nur eine Näherung'",
		"Integral über alle Raumrichtungen",
		"Demonstration funktioniert nicht",
		"Prof sagt 'in erster Ordnung'",
		"Freikörperbild wird gezeichnet",
		"Hamiltonoperator taucht auf",
		"Student fragt nach den Einheiten",
		"Prof wechselt Koordinatensystem",
		"'Das sehen Sie im Praktikum'",
		"Schwingungsgleichung an der Tafel",
		"Prof vergisst Faktor 2",
		"'Stellen Sie sich eine Kugel vor'",
		"Maxwell-Gleichung wird hingeschrieben",
		"Prof sagt 'Das ist Schulstoff'",
		"Taschenrechner gibt falsches Ergebnis",
		"Lagrange-Formalismus erwähnt",
		"Drehimpulserhaltung diskutiert",
		"Doppelspalt-Experiment skizziert",
		"Prof verwechselt c und v",
		"Coulomb-Gesetz hingeschrieben",
		"Resonanzkurve gezeichnet",
		"'Das gilt nur im Vakuum'",
		"Prof rechnet mit kleinem Winkel",
		"Drehmoment falsch gerechnet",
		"Phasenraum erwähnt",
		"'Energie geht nie verloren'",
	},

	// Board 5: "BWL Einführung" (size 5, 32 cells)
	5: {
		"PowerPoint hat zu viel Text",
		"'In der Praxis ist das anders'",
		"Prof erzählt von seiner Firma",
		"SWOT-Analyse wird erwähnt",
		"Jemand fragt 'Kommt das in der Klausur?'",
		"Anglizismus statt deutsches Wort",
		"Stakeholder werden diskutiert",
		"Prof zeigt veraltete Grafik",
		"'Return on Investment'",
		"Fallstudie wird vorgestellt",
		"Prof sagt 'Synergie'",
		"Break-Even-Point an der Tafel",
		"Jemand tippt auf dem Laptop",
		"Prof kommt zu spät",
		"'Das ist ein Paradigmenwechsel'",
		"McKinsey wird erwähnt",
		"Prof zeigt YouTube-Video",
		"Organigramm wird gezeichnet",
		"'Der Markt regelt das'",
		"Bilanz wird erklärt",
		"Prof sagt 'Win-Win-Situation'",
		"KPI-Folie wird gezeigt",
		"Cashflow-Berechnung an der Tafel",
		"Marktanalyse aus 2018",
		"B2B vs B2C Diskussion",
		"Prof sagt 'Skaleneffekte'",
		"Porter's Five Forces erwähnt",
		"'Disruption' fällt mindestens dreimal",
		"Lieferkette wird gezeichnet",
		"Prof zeigt Excel-Pivot-Tabelle",
		"Innovationsmanagement-Buzzword",
		"'Wir denken outside the box'",
	},

	// Board 6: "Algorithmen und Datenstrukturen" (size 5, 32 cells)
	6: {
		"O-Notation wird erklärt",
		"'Das ist in O(n log n)'",
		"Binärbaum wird gezeichnet",
		"Prof sagt 'Divide and Conquer'",
		"Rekursion wird rekursiv erklärt",
		"Bubble Sort als Negativbeispiel",
		"Hashtabelle Kollision",
		"Prof vergisst Basisfall",
		"'Das ist ein bekanntes Problem'",
		"Heap wird aufgebaut",
		"Adjazenzliste vs Adjazenzmatrix",
		"Dijkstra wird falsch ausgesprochen",
		"Prof zeichnet einen Graphen",
		"Stack Overflow Witz",
		"'Das überlasse ich als Übung'",
		"Quicksort Worst Case",
		"Jemand sagt 'Einfach sortieren'",
		"Dynamische Programmierung eingeführt",
		"Prof schreibt Pseudocode",
		"Laufzeitanalyse wird kompliziert",
		"Red-Black-Tree wird erwähnt",
		"'Das geht auch effizienter'",
		"BFS vs DFS Vergleich",
		"Prof vergisst Kantenfall",
		"Amortisierte Analyse",
		"Fibonacci als Beispiel",
		"Prof zeichnet Rekursionsbaum",
		"'Merken Sie sich das für die Klausur'",
		"Greedy-Algorithms vorgestellt",
		"Jemand fragt nach Python-Implementierung",
		"Prof sagt 'Korrektheitsbeweis'",
		"Backtracking wird erklärt",
	},

	// Board 7: "Softwaretechnik Bingo" (size 5, 32 cells)
	7: {
		"UML-Diagramm wird gezeichnet",
		"'Agile ist kein Allheilmittel'",
		"Prof erwähnt Design Patterns",
		"Wasserfall-Modell als Abschreckung",
		"Jemand fragt nach Git",
		"SOLID-Prinzipien werden aufgezählt",
		"Prof sagt 'Technische Schulden'",
		"Sequenzdiagramm wird kompliziert",
		"'In der Industrie macht man das anders'",
		"Scrum-Sprint wird erklärt",
		"Unit Test Beispiel",
		"Prof zeigt Code auf Folie",
		"Refactoring wird erwähnt",
		"'Das ist ein Anti-Pattern'",
		"Code Review Diskussion",
		"Prof erzählt von seinem alten Projekt",
		"Versionskontrolle wird erklärt",
		"'Documentation is key'",
		"CI/CD Pipeline erwähnt",
		"Klassendiagramm an der Tafel",
		"Prof sagt 'Clean Code'",
		"Requirements Engineering Folie",
		"'Das wäre ein gutes Klausurthema'",
		"Singleton wird als Beispiel genutzt",
		"Prof zeigt Gantt-Chart",
		"Pair Programming erwähnt",
		"'Testen ist wichtiger als Coden'",
		"User Story wird formuliert",
		"Prof vergisst Pfeil im Diagramm",
		"Architekturentscheidung diskutiert",
		"'Das hängt vom Kontext ab'",
		"Microservices vs Monolith Debatte",
	},

	// Board 8: "Datenbanken Vorlesung" (size 5, 32 cells)
	8: {
		"ER-Diagramm wird gezeichnet",
		"'Normalisierung ist wichtig'",
		"SQL-Query an der Tafel",
		"Prof sagt 'Redundanz vermeiden'",
		"JOIN wird erklärt",
		"Primärschlüssel vergessen",
		"'Das ist dritte Normalform'",
		"NULL-Werte Diskussion",
		"Prof zeigt SELECT * FROM",
		"Transaktion wird erklärt",
		"ACID-Eigenschaften aufgezählt",
		"'In der Praxis nutzt man einen Index'",
		"Prof schreibt fehlerhaftes SQL",
		"Fremdschlüssel Beziehung",
		"GROUP BY Beispiel",
		"'Das ist ein Deadlock'",
		"Jemand fragt nach NoSQL",
		"Prof zeichnet B-Baum",
		"Relationale Algebra Notation",
		"'Denken Sie mengenorientiert'",
		"Subquery wird verschachtelt",
		"Prof erklärt Views",
		"Trigger werden erwähnt",
		"'Das optimiert der Query-Planner'",
		"Kardinalitäten werden bestimmt",
		"Prof sagt 'Schemamigration'",
		"Concurrency Control Folie",
		"HAVING vs WHERE Verwirrung",
		"Prof zeigt Ausführungsplan",
		"'Das ist klausurrelevant'",
		"Stored Procedures erwähnt",
		"Entity-Relationship-Modell erklärt",
	},

	// Board 9: "Betriebssysteme Chaos" (size 6, 45 cells)
	9: {
		"Prozess vs Thread erklärt",
		"'Das ist ein Race Condition'",
		"Deadlock wird gezeichnet",
		"Prof sagt 'Mutual Exclusion'",
		"Semaphore Beispiel",
		"Context Switch erklärt",
		"'In Linux macht man das anders'",
		"Speicherverwaltung Folie",
		"Prof zeichnet Seitentabelle",
		"Page Fault erklärt",
		"'Das ist ein klassisches Problem'",
		"Scheduling-Algorithms verglichen",
		"Round Robin Beispiel",
		"Prof sagt 'Kernel Mode'",
		"Dateisystem wird erklärt",
		"Interrupt Handler erwähnt",
		"'Stellen Sie sich einen Parkplatz vor'",
		"Philosopher-Problem an der Tafel",
		"Prof vergisst Mutex zu unlocken",
		"Virtual Memory Diagramm",
		"'Das führt zu Starvation'",
		"Systemcall wird erklärt",
		"Prof zeigt Terminal-Demo",
		"Paging vs Segmentierung",
		"'Das ist ein Producer-Consumer-Problem'",
		"I/O-Scheduling erwähnt",
		"Prof sagt 'monolithischer Kernel'",
		"Cache-Hierarchie gezeichnet",
		"'Das sehen Sie im Praktikum'",
		"Zombie-Prozess erklärt",
		"Prof erzählt von Unix-Geschichte",
		"Swapping wird diskutiert",
		"'In der Klausur kommt Scheduling dran'",
		"Belady's Anomalie erwähnt",
		"Prof zeichnet Zustandsdiagramm",
		"Fork-Bomb als Warnung",
		"TLB wird erklärt",
		"'Das ist nicht deterministisch'",
		"Spinlock vs Mutex Vergleich",
		"Prof zeigt htop",
		"Bootvorgang erklärt",
		"IPC-Mechanismen aufgezählt",
		"'Das brauchen Sie für die Übung'",
		"Memory-Mapped I/O erwähnt",
		"Prof sagt 'Das ist Betriebssystemmagie'",
	},

	// Board 10: "Rechnernetze Marathon" (size 6, 45 cells)
	10: {
		"OSI-Modell wird aufgezählt",
		"'Das ist Layer 3'",
		"TCP-Handshake gezeichnet",
		"Prof sagt 'Paketverlust'",
		"IP-Adresse wird berechnet",
		"Subnetting Beispiel",
		"'In der Praxis nutzt man Wireshark'",
		"Prof zeigt Netzwerktopologie",
		"DNS wird erklärt",
		"'Das ist ein Routing-Problem'",
		"HTTP-Statuscodes aufgezählt",
		"Prof vergisst Port-Nummer",
		"ARP-Tabelle an der Tafel",
		"Sliding Window erklärt",
		"'UDP ist schneller aber unsicher'",
		"Jemand fragt nach WLAN",
		"Prof zeichnet Ethernet-Frame",
		"Congestion Control erklärt",
		"'Das ist ein Man-in-the-Middle'",
		"Firewall-Regeln diskutiert",
		"Prof sagt 'Ende-zu-Ende-Prinzip'",
		"NAT wird erklärt",
		"'IPv6 kommt bald... seit 20 Jahren'",
		"Socket-Programmierung erwähnt",
		"Prof zeigt ping-Befehl",
		"DHCP-Ablauf gezeichnet",
		"'Das sehen Sie im Praktikum'",
		"TLS-Handshake wird erklärt",
		"Prof sagt 'Bandbreite vs Durchsatz'",
		"VLAN wird erwähnt",
		"Checksum-Berechnung an der Tafel",
		"'Das ist ein Broadcast-Sturm'",
		"Prof erzählt von altem Modem",
		"Multiplexing erklärt",
		"BGP wird kurz erwähnt",
		"'Traceroute zeigt den Weg'",
		"Prof zeichnet Protokollstapel",
		"Flow Control vs Congestion Control",
		"'Das ist ein SYN-Flood-Angriff'",
		"ICMP wird erklärt",
		"Prof sagt 'Best-Effort-Delivery'",
		"Jemand fragt nach VPN",
		"QoS wird diskutiert",
		"'Das kommt garantiert in der Klausur'",
		"Prof zeigt Wireshark-Capture",
	},

	// Board 11: "Digitaltechnik Vorlesung" (size 6, 45 cells)
	11: {
		"Wahrheitstabelle wird gezeichnet",
		"'Das ist ein NAND-Gatter'",
		"KV-Diagramm an der Tafel",
		"Prof sagt 'Don't-Care-Term'",
		"Flip-Flop wird erklärt",
		"Binärzahl wird umgerechnet",
		"'Das ist Zweierkomplement'",
		"Schaltplan wird gezeichnet",
		"Prof vergisst Taktleitung",
		"Multiplexer Beispiel",
		"'Das minimieren wir mit Karnaugh'",
		"Volladdierer wird aufgebaut",
		"Prof sagt 'Setup- und Hold-Time'",
		"Zustandsautomat wird entworfen",
		"'Das ist ein Hazard'",
		"De-Morgan-Gesetz angewendet",
		"Prof zeichnet Zeitdiagramm",
		"Register-Transfer-Ebene erklärt",
		"'Das ist ein synchrones Design'",
		"ALU wird vorgestellt",
		"Prof zeigt FPGA-Board",
		"Boolesche Algebra Vereinfachung",
		"'Propagation Delay beachten'",
		"Jemand fragt nach VHDL",
		"Prof zeichnet D-Flip-Flop",
		"Encoder vs Decoder erklärt",
		"'Das ist ein Latch, kein Flip-Flop'",
		"Taktfrequenz wird berechnet",
		"Prof sagt 'Pipeline'",
		"Komparator wird entworfen",
		"'Das Signal ist metastabil'",
		"ROM vs RAM erklärt",
		"Prof vergisst Reset-Signal",
		"Counter wird aufgebaut",
		"'Das ist ein Moore-Automat'",
		"Tri-State-Buffer erwähnt",
		"Prof zeigt Simulation",
		"Carry-Look-Ahead erklärt",
		"'Das brauchen Sie für die Übung'",
		"Schieberegister wird gezeichnet",
		"Prof sagt 'CMOS-Technologie'",
		"Glitch im Zeitdiagramm",
		"One-Hot-Encoding erklärt",
		"'Das ist der kritische Pfad'",
		"Prof zeichnet Blockschaltbild",
	},

	// Board 12: "Compilerbau Endgegner" (size 6, 45 cells)
	12: {
		"Lexer vs Parser erklärt",
		"'Das ist ein Token'",
		"Syntaxbaum wird gezeichnet",
		"Prof sagt 'kontextfreie Grammatik'",
		"FIRST- und FOLLOW-Mengen berechnet",
		"'Das ist ein Shift-Reduce-Konflikt'",
		"Prof schreibt BNF-Notation",
		"Regulärer Ausdruck für Lexer",
		"Abstract Syntax Tree erklärt",
		"'Das optimiert der Compiler weg'",
		"Prof zeichnet Parsertabelle",
		"Semantic Analysis erwähnt",
		"'Das ist ein Typfehler'",
		"Symboltabelle wird aufgebaut",
		"Prof sagt 'Zwischencode'",
		"Three-Address-Code Beispiel",
		"Code Generation Phase",
		"'Das ist tote Code-Elimination'",
		"Prof vergisst Produktionsregel",
		"LL(1) vs LR(1) Vergleich",
		"'Das ist nicht LL(1)-parsbar'",
		"Prof zeigt Compiler-Pipeline",
		"Attributierte Grammatik erklärt",
		"'Das ist ein Dangling-Else-Problem'",
		"Prof schreibt Übersetzungsregel",
		"Register Allocation erwähnt",
		"'Graph Coloring für Register'",
		"Peephole-Optimierung erklärt",
		"Prof sagt 'Bootstrapping'",
		"Garbage Collection diskutiert",
		"'Das macht LLVM für uns'",
		"Prof zeichnet Kontrollflußgraphen",
		"SSA-Form wird eingeführt",
		"'Das ist eine Schleifenoptimierung'",
		"Jemand fragt 'Warum nicht einfach Python?'",
		"Prof erklärt Activation Records",
		"Stack Frame gezeichnet",
		"'Das ist ein Scope-Problem'",
		"Operator-Priorität Tabelle",
		"Prof sagt 'Links-Rekursion eliminieren'",
		"Lookahead-Token erklärt",
		"'Das ist syntaktischer Zucker'",
		"Prof zeigt generierten Assembler",
		"Constant Folding Beispiel",
		"'Das sind implizite Typkonvertierungen'",
	},

	// Board 13: "Verteilte Systeme Bingo" (size 5, 32 cells)
	13: {
		"CAP-Theorem wird erklärt",
		"'Das ist ein verteilter Konsens'",
		"Prof zeichnet Netzwerkpartition",
		"Paxos wird erwähnt",
		"'Raft ist einfacher als Paxos'",
		"Prof sagt 'Eventual Consistency'",
		"Two-Phase-Commit erklärt",
		"'Das ist ein Byzantine Fault'",
		"Prof zeichnet Lamport-Uhr",
		"Vector Clocks vorgestellt",
		"'Das skaliert nicht'",
		"MapReduce Beispiel",
		"Prof sagt 'Idempotenz ist wichtig'",
		"RPC wird erklärt",
		"'Das ist ein Single Point of Failure'",
		"Load Balancer gezeichnet",
		"Prof erwähnt Microservices",
		"'Service Discovery braucht man auch'",
		"Jemand fragt nach Kubernetes",
		"Prof sagt 'Das ist ein Fallacy'",
		"Replikation wird diskutiert",
		"'Leader Election ist nicht trivial'",
		"Consistent Hashing erklärt",
		"Prof zeichnet Ring-Topologie",
		"'Das ist ein Split-Brain-Problem'",
		"Gossip-Protokoll vorgestellt",
		"Prof sagt 'Heartbeat-Mechanismus'",
		"'Timeouts richtig setzen ist schwer'",
		"Circuit Breaker Pattern erklärt",
		"Prof zeigt Architektur-Diagramm",
		"'Das braucht ein Message-Queue'",
		"Saga-Pattern wird erwähnt",
	},

	// Board 14: "Maschinelles Lernen Bingo" (size 5, 32 cells)
	14: {
		"'Das ist ein Klassifikationsproblem'",
		"Prof zeichnet Entscheidungsgrenze",
		"Gradient Descent erklärt",
		"'Die Lernrate ist zu hoch'",
		"Prof sagt 'Overfitting vermeiden'",
		"Confusion Matrix gezeichnet",
		"'Das ist ein Feature'",
		"Train-Test-Split erklärt",
		"Prof zeigt Jupyter Notebook",
		"'Normalisierung nicht vergessen'",
		"Bias-Variance-Tradeoff diskutiert",
		"Kreuzvalidierung vorgestellt",
		"'Das neuronale Netz lernt nicht'",
		"Prof zeichnet Perzeptron",
		"Backpropagation erklärt",
		"'Das ist ein Hyperparameter'",
		"Prof sagt 'Mehr Daten helfen'",
		"Random Forest als Beispiel",
		"'SVM mit Kernel-Trick'",
		"k-Nearest-Neighbors erklärt",
		"Prof zeigt Scatter-Plot",
		"'Das ist ein Clustering-Problem'",
		"k-Means Algorithms vorgestellt",
		"Dimensionsreduktion erwähnt",
		"'PCA ist Ihr Freund'",
		"Prof sagt 'Regularisierung'",
		"Dropout wird erklärt",
		"'Batch Size beeinflusst das Training'",
		"Prof zeigt Loss-Kurve",
		"'Das Modell konvergiert'",
		"Aktivierungsfunktion erklärt",
		"'ReLU statt Sigmoid'",
	},

	// Board 15: "Programmierkurs Klassiker" (size 4, 21 cells, German)
	15: {
		"Prof vergisst Semikolon",
		"Live-Demo crasht",
		"'Das funktioniert auf meinem Rechner'",
		"Stack Overflow wird offen gezeigt",
		"IDE friert ein",
		"Prof tippt sehr langsam",
		"Variable heißt 'temp'",
		"Endlosschleife im Beispiel",
		"'Ignorieren Sie diese Warnung'",
		"Prof vergisst zu speichern",
		"Tippfehler im Methodennamen",
		"Compiler meckert",
		"'Das ist nur Pseudocode'",
		"Prof googelt während der Vorlesung",
		"Off-by-One-Fehler",
		"Hello World funktioniert nicht",
		"Prof verwechselt == und =",
		"Indentation falsch",
		"'Das pushen wir später'",
		"Linter unterringelt alles",
		"Konsolenausgabe verschwindet",
	},

	// Board 16: "Webtechnologien Bingo" (size 5, 32 cells, German)
	16: {
		"'CORS-Fehler'",
		"Prof zeigt 404-Seite",
		"JavaScript verhält sich seltsam",
		"'CSS ist auch Programmierung'",
		"REST vs SOAP Diskussion",
		"Prof öffnet DevTools",
		"'Das sieht in Safari anders aus'",
		"JSON-Parse-Error",
		"Cookie wird nicht gesetzt",
		"Prof erwähnt jQuery",
		"'Promise.then-Hölle'",
		"HTTPS-Zertifikat abgelaufen",
		"Cache-Problem im Browser",
		"Prof zeigt Lighthouse-Score",
		"'Das löst React für uns'",
		"Async/Await falsch verwendet",
		"DOM wird manipuliert",
		"'Single-Page-Application'",
		"Prof verwirrt durch z-index",
		"Flexbox vs Grid Debatte",
		"'!important verwenden'",
		"Prof zeigt Wireshark-Trace",
		"WebSocket-Verbindung bricht ab",
		"Form Submission ohne preventDefault",
		"Prof sagt 'Hydration'",
		"Bundle-Size zu groß",
		"'Tailwind oder Vanilla CSS?'",
		"TypeScript-Fehler ignoriert",
		"Prof zeigt Network-Tab",
		"'localhost:3000 ist down'",
		"Service Worker macht Probleme",
		"Prof sagt 'Progressive Enhancement'",
	},

	// Board 17: "Admin Debug Bingo" (size 4, 21 cells, English, with hidden references)
	17: {
		"Tests pass on second try",
		"`TODO fix later` from 2019",
		"Admin force-pushes to main",
		"'It works on my machine'",
		"42 unread Slack pings",
		"Prod is on fire",
		"Coffee runs out at 3 AM",
		"Someone blames the intern",
		"`rm -rf` typed too fast",
		"Linter wins the argument",
		"'Have you tried turning it off and on again?'",
		"'I'm sorry, Dave, I'm afraid I can't do that.'",
		"'There is no spoon.'",
		"Stack trace longer than the code",
		"Hotfix on a Friday at 5 PM",
		"Database is 'just slow'",
		"'It's a UNIX system, I know this!'",
		"Commit message says 'stuff'",
		"PR approved without review",
		"Cache invalidation goes wrong",
		"'Hello, IT. Have you tried forcing an unexpected reboot?'",
	},
}

// votes defines vote data. Each entry is a (userIdx, boardIdx, value) tuple.
// The UNIQUE(user_id, board_id) constraint means each user can only vote once per board.
var votes = []voteData{
	{UserIdx: 0, BoardIdx: 0, Value: 1},
	{UserIdx: 1, BoardIdx: 0, Value: 1},
	{UserIdx: 2, BoardIdx: 0, Value: -1},
	{UserIdx: 3, BoardIdx: 1, Value: 1},
	{UserIdx: 4, BoardIdx: 1, Value: 1},
	{UserIdx: 5, BoardIdx: 2, Value: -1},
	{UserIdx: 6, BoardIdx: 2, Value: 1},
	{UserIdx: 0, BoardIdx: 3, Value: 1},
	{UserIdx: 1, BoardIdx: 3, Value: 1},
	{UserIdx: 2, BoardIdx: 3, Value: 1},
	{UserIdx: 7, BoardIdx: 4, Value: -1},
	{UserIdx: 8, BoardIdx: 4, Value: 1},
	{UserIdx: 9, BoardIdx: 5, Value: 1},
	{UserIdx: 3, BoardIdx: 6, Value: 1},
	{UserIdx: 4, BoardIdx: 6, Value: -1},
	{UserIdx: 5, BoardIdx: 6, Value: 1},
	{UserIdx: 6, BoardIdx: 7, Value: 1},
	{UserIdx: 7, BoardIdx: 7, Value: 1},
	{UserIdx: 8, BoardIdx: 8, Value: -1},
	{UserIdx: 9, BoardIdx: 8, Value: 1},
	{UserIdx: 0, BoardIdx: 9, Value: 1},
	{UserIdx: 1, BoardIdx: 10, Value: 1},
	{UserIdx: 2, BoardIdx: 11, Value: -1},
	{UserIdx: 3, BoardIdx: 12, Value: 1},
	{UserIdx: 4, BoardIdx: 13, Value: 1},
	{UserIdx: 5, BoardIdx: 14, Value: 1},
	{UserIdx: 6, BoardIdx: 14, Value: -1},
	{UserIdx: 7, BoardIdx: 12, Value: 1},
	{UserIdx: 8, BoardIdx: 10, Value: 1},
	{UserIdx: 9, BoardIdx: 9, Value: -1},

	// Admin (UserIdx 12) votes on roughly half the boards so the user can
	// easily find both voted and unvoted boards in the UI.
	// Voted boards: 0, 2, 6, 9, 12, 15, 16, 17 (8 of 18).
	// Unvoted boards: 1, 3, 4, 5, 7, 8, 10, 11, 13, 14.
	{UserIdx: 12, BoardIdx: 0, Value: 1},
	{UserIdx: 12, BoardIdx: 2, Value: 1},
	{UserIdx: 12, BoardIdx: 6, Value: -1},
	{UserIdx: 12, BoardIdx: 9, Value: 1},
	{UserIdx: 12, BoardIdx: 12, Value: 1},
	{UserIdx: 12, BoardIdx: 15, Value: 1}, // own board
	{UserIdx: 12, BoardIdx: 16, Value: 1}, // own board
	{UserIdx: 12, BoardIdx: 17, Value: 1}, // own board

	// A few other users vote on admin's boards so they show non-zero scores.
	{UserIdx: 0, BoardIdx: 15, Value: 1},
	{UserIdx: 1, BoardIdx: 16, Value: 1},
	{UserIdx: 3, BoardIdx: 16, Value: -1},
	{UserIdx: 4, BoardIdx: 17, Value: 1},
	{UserIdx: 7, BoardIdx: 17, Value: 1},
}

// games defines game sessions. Each game is played by a user on a board.
var games = []gameData{
	// Game 0: User 0 plays board 0 (Mathe 1 Bingo, size 3 -> 9 cells) - active
	// Played once yesterday, then came back today.
	{PlayerIdx: 0, SessionIdx: -1, BoardIdx: 0, Status: "active", Sessions: []gameSessionSeed{
		{StartedAgoMinutes: 26 * 60, DurationMinutes: 8},
		{StartedAgoMinutes: 2 * 60, DurationMinutes: 12},
	}},
	// Game 1: User 1 plays board 0 (Mathe 1 Bingo) - active
	// Single short session, currently in progress.
	{PlayerIdx: 1, SessionIdx: -1, BoardIdx: 0, Status: "active", Sessions: []gameSessionSeed{
		{StartedAgoMinutes: 30, DurationMinutes: 7},
	}},
	// Game 2: User 2 plays board 3 (Theoretische Informatik, size 4 -> 16 cells) - active
	// Tried briefly 42h ago, resumed today for a longer session.
	{PlayerIdx: 2, SessionIdx: -1, BoardIdx: 3, Status: "active", Sessions: []gameSessionSeed{
		{StartedAgoMinutes: 42 * 60, DurationMinutes: 15},
		{StartedAgoMinutes: 6 * 60, DurationMinutes: 34},
	}},
	// Game 3: User 3 plays board 6 (Algorithmen und Datenstrukturen, size 5 -> 25 cells) - abandoned
	// Short single session 36h ago, then abandoned.
	{PlayerIdx: 3, SessionIdx: -1, BoardIdx: 6, Status: "abandoned", Sessions: []gameSessionSeed{
		{StartedAgoMinutes: 36 * 60, DurationMinutes: 5},
	}},
	// Game 4: User 5 plays board 1 (Lineare Algebra Klassiker, size 3 -> 9 cells) - active
	// Single session started an hour ago.
	{PlayerIdx: 5, SessionIdx: -1, BoardIdx: 1, Status: "active", Sessions: []gameSessionSeed{
		{StartedAgoMinutes: 60, DurationMinutes: 21},
	}},
	// Game 5: Anonymous (session 0) plays board 0 (Mathe 1 Bingo, size 4 -> 16 cells) - active
	// Played 18h ago, still open.
	{PlayerIdx: -1, SessionIdx: 0, BoardIdx: 0, Status: "active", Sessions: []gameSessionSeed{
		{StartedAgoMinutes: 18 * 60, DurationMinutes: 9},
	}},
	// Game 6: Anonymous (session 1) plays board 3 (Theoretische Informatik) - abandoned
	// Brief session 44h ago, abandoned.
	{PlayerIdx: -1, SessionIdx: 1, BoardIdx: 3, Status: "abandoned", Sessions: []gameSessionSeed{
		{StartedAgoMinutes: 44 * 60, DurationMinutes: 3},
	}},
	// Game 7: maxmustermann via bound session, board 8 (Datenbanken, size 5 -> 25 cells) - active
	// Played yesterday, came back for a long session this afternoon.
	{PlayerIdx: 0, SessionIdx: 3, BoardIdx: 8, Status: "active", Sessions: []gameSessionSeed{
		{StartedAgoMinutes: 20 * 60, DurationMinutes: 25},
		{StartedAgoMinutes: 4 * 60, DurationMinutes: 45},
	}},
	// Game 8: admin plays board 15 (Programmierkurs Klassiker, size 4 -> 16 cells) - active
	// Single session 12h ago.
	{PlayerIdx: 12, SessionIdx: -1, BoardIdx: 15, Status: "active", Sessions: []gameSessionSeed{
		{StartedAgoMinutes: 12 * 60, DurationMinutes: 18},
	}},
	// Game 9: admin plays board 17 (Admin Debug Bingo, size 4 -> 16 cells) - active
	// Played near the 48h boundary, then again yesterday morning.
	{PlayerIdx: 12, SessionIdx: -1, BoardIdx: 17, Status: "active", Sessions: []gameSessionSeed{
		{StartedAgoMinutes: 47 * 60, DurationMinutes: 10},
		{StartedAgoMinutes: 24 * 60, DurationMinutes: 26},
	}},
	// Game 10: admin plays board 6 (Algorithmen, size 5 -> 25 cells) via bound session - active
	// One long session recently.
	{PlayerIdx: 12, SessionIdx: 5, BoardIdx: 6, Status: "active", Sessions: []gameSessionSeed{
		{StartedAgoMinutes: 40, DurationMinutes: 38},
	}},
}

// gameCells defines the game_cells for each game index.
// Each game picks n² cells from the board's cell pool and assigns positions.
// Position is 0-indexed in row-major order.
var gameCells = map[int][]gameCellData{
	// Game 0: User 0 on board 0 (Mathe 1, size 4 -> 16 cells, completed with marks)
	0: {
		{Content: "Prof sagt 'trivial'", Position: 0, IsMarked: true},
		{Content: "Beweis wird übersprungen", Position: 1, IsMarked: true},
		{Content: "'Das sieht man sofort'", Position: 2, IsMarked: true},
		{Content: "Kreide bricht ab", Position: 3, IsMarked: true},
		{Content: "Tafel ist voll, Prof wischt hektisch", Position: 4, IsMarked: false},
		{Content: "Jemand fragt 'Kommt das in der Klausur?'", Position: 5, IsMarked: true},
		{Content: "Prof rechnet sich an der Tafel vor", Position: 6, IsMarked: true},
		{Content: "Epsilon-Delta taucht auf", Position: 7, IsMarked: true},
		{Content: "'Übung für den Leser'", Position: 8, IsMarked: true},
		{Content: "Prof verwechselt Plus und Minus", Position: 9, IsMarked: true},
		{Content: "Studenten schauen verwirrt", Position: 10, IsMarked: true},
		{Content: "'Das hatten wir schon letzte Woche'", Position: 11, IsMarked: false},
		{Content: "Tippfehler auf der Folie", Position: 12, IsMarked: true},
		{Content: "Prof sucht den Tafellappen", Position: 13, IsMarked: true},
		{Content: "Hörsaal-Mikrofon rauscht", Position: 14, IsMarked: false},
		{Content: "Tutor wird widersprochen", Position: 15, IsMarked: true},
	},

	// Game 1: User 1 on board 0 (Mathe 1, size 4 -> 16 cells, active with some marks)
	1: {
		{Content: "'Das sieht man sofort'", Position: 0, IsMarked: true},
		{Content: "Prof verwechselt Plus und Minus", Position: 1, IsMarked: false},
		{Content: "Studenten schauen verwirrt", Position: 2, IsMarked: true},
		{Content: "Kreide bricht ab", Position: 3, IsMarked: false},
		{Content: "Prof sagt 'trivial'", Position: 4, IsMarked: true},
		{Content: "'Das hatten wir schon letzte Woche'", Position: 5, IsMarked: false},
		{Content: "Beweis wird übersprungen", Position: 6, IsMarked: false},
		{Content: "Epsilon-Delta taucht auf", Position: 7, IsMarked: true},
		{Content: "Jemand fragt 'Kommt das in der Klausur?'", Position: 8, IsMarked: false},
		{Content: "'Übung für den Leser'", Position: 9, IsMarked: true},
		{Content: "Tippfehler auf der Folie", Position: 10, IsMarked: false},
		{Content: "Beamer flackert kurz", Position: 11, IsMarked: false},
		{Content: "'Mit vollständiger Induktion'", Position: 12, IsMarked: true},
		{Content: "Prof zeichnet Cantor-Mengen", Position: 13, IsMarked: false},
		{Content: "Grenzwert wird hingeschrieben", Position: 14, IsMarked: false},
		{Content: "'Die Klausur wird fair'", Position: 15, IsMarked: false},
	},

	// Game 2: User 2 on board 3 (Theoretische Informatik, size 4 -> 16 cells, active)
	2: {
		{Content: "Turingmaschine wird gezeichnet", Position: 0, IsMarked: true},
		{Content: "'Das ist unentscheidbar'", Position: 1, IsMarked: false},
		{Content: "Regulärer Ausdruck wird kompliziert", Position: 2, IsMarked: true},
		{Content: "Pumping Lemma Beweis", Position: 3, IsMarked: false},
		{Content: "Prof sagt 'Nichtdeterminismus'", Position: 4, IsMarked: true},
		{Content: "Automat hat zu viele Zustände", Position: 5, IsMarked: false},
		{Content: "Jemand fragt 'Wozu braucht man das?'", Position: 6, IsMarked: true},
		{Content: "Chomsky-Hierarchie an der Tafel", Position: 7, IsMarked: false},
		{Content: "Prof vergisst Endzustand", Position: 8, IsMarked: false},
		{Content: "'Das reduzieren wir auf das Halteproblem'", Position: 9, IsMarked: true},
		{Content: "Komplexitätsklasse P vs NP", Position: 10, IsMarked: false},
		{Content: "Beweis durch Widerspruch", Position: 11, IsMarked: false},
		{Content: "Prof schreibt Produktionsregel falsch", Position: 12, IsMarked: true},
		{Content: "Kellerautomat wird eingeführt", Position: 13, IsMarked: false},
		{Content: "Jemand sagt 'Das ist doch wie Mathe'", Position: 14, IsMarked: true},
		{Content: "Sprache wird nicht akzeptiert", Position: 15, IsMarked: false},
	},

	// Game 3: User 3 on board 6 (Algorithmen, size 5 -> 25 cells, abandoned with few marks)
	3: {
		{Content: "O-Notation wird erklärt", Position: 0, IsMarked: true},
		{Content: "'Das ist in O(n log n)'", Position: 1, IsMarked: false},
		{Content: "Binärbaum wird gezeichnet", Position: 2, IsMarked: false},
		{Content: "Prof sagt 'Divide and Conquer'", Position: 3, IsMarked: true},
		{Content: "Rekursion wird rekursiv erklärt", Position: 4, IsMarked: false},
		{Content: "Bubble Sort als Negativbeispiel", Position: 5, IsMarked: false},
		{Content: "Hashtabelle Kollision", Position: 6, IsMarked: false},
		{Content: "Prof vergisst Basisfall", Position: 7, IsMarked: false},
		{Content: "'Das ist ein bekanntes Problem'", Position: 8, IsMarked: true},
		{Content: "Heap wird aufgebaut", Position: 9, IsMarked: false},
		{Content: "Adjazenzliste vs Adjazenzmatrix", Position: 10, IsMarked: false},
		{Content: "Dijkstra wird falsch ausgesprochen", Position: 11, IsMarked: false},
		{Content: "Prof zeichnet einen Graphen", Position: 12, IsMarked: false},
		{Content: "Stack Overflow Witz", Position: 13, IsMarked: false},
		{Content: "'Das überlasse ich als Übung'", Position: 14, IsMarked: false},
		{Content: "Quicksort Worst Case", Position: 15, IsMarked: false},
		{Content: "Jemand sagt 'Einfach sortieren'", Position: 16, IsMarked: false},
		{Content: "Dynamische Programmierung eingeführt", Position: 17, IsMarked: false},
		{Content: "Prof schreibt Pseudocode", Position: 18, IsMarked: false},
		{Content: "Laufzeitanalyse wird kompliziert", Position: 19, IsMarked: false},
		{Content: "Red-Black-Tree wird erwähnt", Position: 20, IsMarked: false},
		{Content: "'Das geht auch effizienter'", Position: 21, IsMarked: false},
		{Content: "BFS vs DFS Vergleich", Position: 22, IsMarked: false},
		{Content: "Prof vergisst Kantenfall", Position: 23, IsMarked: false},
		{Content: "Amortisierte Analyse", Position: 24, IsMarked: false},
	},

	// Game 4: User 5 on board 1 (Lineare Algebra, size 5 -> 25 cells, active)
	4: {
		{Content: "Matrix wird falsch multipliziert", Position: 0, IsMarked: false},
		{Content: "'Der Beweis ist elegant'", Position: 1, IsMarked: true},
		{Content: "Determinante vergessen", Position: 2, IsMarked: false},
		{Content: "Prof zeichnet einen Vektorraum", Position: 3, IsMarked: true},
		{Content: "Eigenwert-Witz", Position: 4, IsMarked: false},
		{Content: "Jemand schläft ein", Position: 5, IsMarked: false},
		{Content: "Prof sagt 'offensichtlich'", Position: 6, IsMarked: true},
		{Content: "Rang der Matrix unklar", Position: 7, IsMarked: false},
		{Content: "'Das ist ein Spezialfall'", Position: 8, IsMarked: false},
		{Content: "Gauß-Algorithms zum dritten Mal", Position: 9, IsMarked: true},
		{Content: "Prof dreht sich zur falschen Tafel", Position: 10, IsMarked: false},
		{Content: "Inverse existiert nicht", Position: 11, IsMarked: false},
		{Content: "Basiswechsel verwirrt alle", Position: 12, IsMarked: false},
		{Content: "Skalarprodukt falsch berechnet", Position: 13, IsMarked: true},
		{Content: "'Diagonalisierbar oder nicht?'", Position: 14, IsMarked: false},
		{Content: "Kern und Bild verwechselt", Position: 15, IsMarked: false},
		{Content: "Eigenvektor wird berechnet", Position: 16, IsMarked: true},
		{Content: "Spur einer Matrix erwähnt", Position: 17, IsMarked: false},
		{Content: "Lineare Abbildung gezeichnet", Position: 18, IsMarked: false},
		{Content: "Cramersche Regel angewendet", Position: 19, IsMarked: false},
		{Content: "Prof verwechselt Zeile und Spalte", Position: 20, IsMarked: true},
		{Content: "Untervektorraum-Beispiel", Position: 21, IsMarked: false},
		{Content: "Prof zeichnet R³-Koordinaten", Position: 22, IsMarked: false},
		{Content: "'Das ist eine Bilinearform'", Position: 23, IsMarked: false},
		{Content: "Prof rechnet Kreuzprodukt vor", Position: 24, IsMarked: false},
	},

	// Game 5: Anonymous on board 0 (Mathe 1, size 4 -> 16 cells, active, a couple of marks)
	5: {
		{Content: "Prof sagt 'trivial'", Position: 0, IsMarked: true},
		{Content: "Beweis wird übersprungen", Position: 1, IsMarked: false},
		{Content: "'Das sieht man sofort'", Position: 2, IsMarked: false},
		{Content: "Kreide bricht ab", Position: 3, IsMarked: true},
		{Content: "Tafel ist voll, Prof wischt hektisch", Position: 4, IsMarked: false},
		{Content: "Jemand fragt 'Kommt das in der Klausur?'", Position: 5, IsMarked: false},
		{Content: "Prof rechnet sich an der Tafel vor", Position: 6, IsMarked: false},
		{Content: "Epsilon-Delta taucht auf", Position: 7, IsMarked: false},
		{Content: "'Übung für den Leser'", Position: 8, IsMarked: false},
		{Content: "Prof verwechselt Plus und Minus", Position: 9, IsMarked: true},
		{Content: "Studenten schauen verwirrt", Position: 10, IsMarked: false},
		{Content: "'Das hatten wir schon letzte Woche'", Position: 11, IsMarked: false},
		{Content: "Tippfehler auf der Folie", Position: 12, IsMarked: false},
		{Content: "Prof sucht den Tafellappen", Position: 13, IsMarked: false},
		{Content: "Hörsaal-Mikrofon rauscht", Position: 14, IsMarked: false},
		{Content: "Tutor wird widersprochen", Position: 15, IsMarked: false},
	},

	// Game 6: Anonymous on board 3 (Theoretische Informatik, size 4 -> 16 cells, abandoned, a few marks)
	6: {
		{Content: "Turingmaschine wird gezeichnet", Position: 0, IsMarked: true},
		{Content: "'Das ist unentscheidbar'", Position: 1, IsMarked: true},
		{Content: "Regulärer Ausdruck wird kompliziert", Position: 2, IsMarked: false},
		{Content: "Pumping Lemma Beweis", Position: 3, IsMarked: false},
		{Content: "Prof sagt 'Nichtdeterminismus'", Position: 4, IsMarked: false},
		{Content: "Automat hat zu viele Zustände", Position: 5, IsMarked: false},
		{Content: "Jemand fragt 'Wozu braucht man das?'", Position: 6, IsMarked: true},
		{Content: "Chomsky-Hierarchie an der Tafel", Position: 7, IsMarked: false},
		{Content: "Prof vergisst Endzustand", Position: 8, IsMarked: false},
		{Content: "'Das reduzieren wir auf das Halteproblem'", Position: 9, IsMarked: false},
		{Content: "Komplexitätsklasse P vs NP", Position: 10, IsMarked: false},
		{Content: "Beweis durch Widerspruch", Position: 11, IsMarked: false},
		{Content: "Prof schreibt Produktionsregel falsch", Position: 12, IsMarked: false},
		{Content: "Kellerautomat wird eingeführt", Position: 13, IsMarked: false},
		{Content: "Jemand sagt 'Das ist doch wie Mathe'", Position: 14, IsMarked: false},
		{Content: "Sprache wird nicht akzeptiert", Position: 15, IsMarked: false},
	},

	// Game 7: maxmustermann via bound session on board 8 (Datenbanken, size 5 -> 25 cells, completed, many marks)
	7: {
		{Content: "ER-Diagramm wird gezeichnet", Position: 0, IsMarked: true},
		{Content: "'Normalisierung ist wichtig'", Position: 1, IsMarked: true},
		{Content: "SQL-Query an der Tafel", Position: 2, IsMarked: true},
		{Content: "Prof sagt 'Redundanz vermeiden'", Position: 3, IsMarked: false},
		{Content: "JOIN wird erklärt", Position: 4, IsMarked: true},
		{Content: "Primärschlüssel vergessen", Position: 5, IsMarked: true},
		{Content: "'Das ist dritte Normalform'", Position: 6, IsMarked: false},
		{Content: "NULL-Werte Diskussion", Position: 7, IsMarked: true},
		{Content: "Prof zeigt SELECT * FROM", Position: 8, IsMarked: true},
		{Content: "Transaktion wird erklärt", Position: 9, IsMarked: false},
		{Content: "ACID-Eigenschaften aufgezählt", Position: 10, IsMarked: true},
		{Content: "'In der Praxis nutzt man einen Index'", Position: 11, IsMarked: true},
		{Content: "Prof schreibt fehlerhaftes SQL", Position: 12, IsMarked: false},
		{Content: "Fremdschlüssel Beziehung", Position: 13, IsMarked: true},
		{Content: "GROUP BY Beispiel", Position: 14, IsMarked: true},
		{Content: "'Das ist ein Deadlock'", Position: 15, IsMarked: false},
		{Content: "Jemand fragt nach NoSQL", Position: 16, IsMarked: true},
		{Content: "Prof zeichnet B-Baum", Position: 17, IsMarked: false},
		{Content: "Relationale Algebra Notation", Position: 18, IsMarked: true},
		{Content: "'Denken Sie mengenorientiert'", Position: 19, IsMarked: false},
		{Content: "Subquery wird verschachtelt", Position: 20, IsMarked: true},
		{Content: "Prof erklärt Views", Position: 21, IsMarked: true},
		{Content: "Trigger werden erwähnt", Position: 22, IsMarked: false},
		{Content: "'Das optimiert der Query-Planner'", Position: 23, IsMarked: true},
		{Content: "Kardinalitäten werden bestimmt", Position: 24, IsMarked: true},
	},

	// Game 8: admin on board 15 (Programmierkurs Klassiker, size 4 -> 16 cells, completed, mostly marked)
	8: {
		{Content: "Prof vergisst Semikolon", Position: 0, IsMarked: true},
		{Content: "Live-Demo crasht", Position: 1, IsMarked: true},
		{Content: "'Das funktioniert auf meinem Rechner'", Position: 2, IsMarked: true},
		{Content: "Stack Overflow wird offen gezeigt", Position: 3, IsMarked: true},
		{Content: "IDE friert ein", Position: 4, IsMarked: false},
		{Content: "Prof tippt sehr langsam", Position: 5, IsMarked: true},
		{Content: "Variable heißt 'temp'", Position: 6, IsMarked: true},
		{Content: "Endlosschleife im Beispiel", Position: 7, IsMarked: true},
		{Content: "'Ignorieren Sie diese Warnung'", Position: 8, IsMarked: false},
		{Content: "Prof vergisst zu speichern", Position: 9, IsMarked: true},
		{Content: "Tippfehler im Methodennamen", Position: 10, IsMarked: true},
		{Content: "Compiler meckert", Position: 11, IsMarked: true},
		{Content: "'Das ist nur Pseudocode'", Position: 12, IsMarked: false},
		{Content: "Prof googelt während der Vorlesung", Position: 13, IsMarked: true},
		{Content: "Off-by-One-Fehler", Position: 14, IsMarked: true},
		{Content: "Hello World funktioniert nicht", Position: 15, IsMarked: false},
	},

	// Game 9: admin on board 17 (Admin Debug Bingo, size 4 -> 16 cells, active, a few marks)
	9: {
		{Content: "Tests pass on second try", Position: 0, IsMarked: true},
		{Content: "`TODO fix later` from 2019", Position: 1, IsMarked: true},
		{Content: "Admin force-pushes to main", Position: 2, IsMarked: false},
		{Content: "'It works on my machine'", Position: 3, IsMarked: true},
		{Content: "42 unread Slack pings", Position: 4, IsMarked: false},
		{Content: "Prod is on fire", Position: 5, IsMarked: false},
		{Content: "Coffee runs out at 3 AM", Position: 6, IsMarked: true},
		{Content: "Someone blames the intern", Position: 7, IsMarked: false},
		{Content: "`rm -rf` typed too fast", Position: 8, IsMarked: false},
		{Content: "Linter wins the argument", Position: 9, IsMarked: true},
		{Content: "'Have you tried turning it off and on again?'", Position: 10, IsMarked: false},
		{Content: "'I'm sorry, Dave, I'm afraid I can't do that.'", Position: 11, IsMarked: false},
		{Content: "'There is no spoon.'", Position: 12, IsMarked: false},
		{Content: "Stack trace longer than the code", Position: 13, IsMarked: false},
		{Content: "Hotfix on a Friday at 5 PM", Position: 14, IsMarked: false},
		{Content: "Database is 'just slow'", Position: 15, IsMarked: false},
	},

	// Game 10: admin on board 6 (Algorithmen, size 5 -> 25 cells, active, early progress)
	10: {
		{Content: "O-Notation wird erklärt", Position: 0, IsMarked: true},
		{Content: "'Das ist in O(n log n)'", Position: 1, IsMarked: true},
		{Content: "Binärbaum wird gezeichnet", Position: 2, IsMarked: false},
		{Content: "Prof sagt 'Divide and Conquer'", Position: 3, IsMarked: true},
		{Content: "Rekursion wird rekursiv erklärt", Position: 4, IsMarked: true},
		{Content: "Bubble Sort als Negativbeispiel", Position: 5, IsMarked: false},
		{Content: "Hashtabelle Kollision", Position: 6, IsMarked: false},
		{Content: "Prof vergisst Basisfall", Position: 7, IsMarked: true},
		{Content: "'Das ist ein bekanntes Problem'", Position: 8, IsMarked: false},
		{Content: "Heap wird aufgebaut", Position: 9, IsMarked: true},
		{Content: "Adjazenzliste vs Adjazenzmatrix", Position: 10, IsMarked: false},
		{Content: "Dijkstra wird falsch ausgesprochen", Position: 11, IsMarked: true},
		{Content: "Prof zeichnet einen Graphen", Position: 12, IsMarked: false},
		{Content: "Stack Overflow Witz", Position: 13, IsMarked: true},
		{Content: "'Das überlasse ich als Übung'", Position: 14, IsMarked: false},
		{Content: "Quicksort Worst Case", Position: 15, IsMarked: false},
		{Content: "Jemand sagt 'Einfach sortieren'", Position: 16, IsMarked: false},
		{Content: "Dynamische Programmierung eingeführt", Position: 17, IsMarked: false},
		{Content: "Prof schreibt Pseudocode", Position: 18, IsMarked: false},
		{Content: "Laufzeitanalyse wird kompliziert", Position: 19, IsMarked: false},
		{Content: "Red-Black-Tree wird erwähnt", Position: 20, IsMarked: false},
		{Content: "'Das geht auch effizienter'", Position: 21, IsMarked: false},
		{Content: "BFS vs DFS Vergleich", Position: 22, IsMarked: false},
		{Content: "Prof vergisst Kantenfall", Position: 23, IsMarked: false},
		{Content: "Amortisierte Analyse", Position: 24, IsMarked: false},
	},
}

// passwordData holds a password to seed for a user.
// UserIdx refers to an index in the users slice.
type passwordData struct {
	UserIdx  int
	Password string
}

// passwords defines which users get a password for local login.
var passwords = []passwordData{
	{UserIdx: 0, Password: "password123"},
	{UserIdx: 1, Password: "securePass!"},
	{UserIdx: 2, Password: "timoSecret42"},
	{UserIdx: 10, Password: "ghostpass"},   // user with no email
	{UserIdx: 11, Password: "anonpass"},    // user with no email
	{UserIdx: 12, Password: "password123"}, // admin / easy test account
}

// sessionData holds a session row to seed.
// UserIdx == -1 means an anonymous session (user_id NULL).
// ExpiresInHours is relative to now(); use a negative value to seed a row
// that is already expired (handy for verifying the cleanup job).
type sessionData struct {
	UserIdx        int
	Token          string
	ExpiresInHours int
}

// sessions defines a small set of sessions to seed for local development.
// The tokens are intentionally predictable so they can be pasted directly into
// a `Cookie: session_token=...` header for manual testing.
var sessions = []sessionData{
	// Anonymous, fresh
	{UserIdx: -1, Token: "seed-anon-fresh-token-0001", ExpiresInHours: 30 * 24},
	// Anonymous, close to expiry (< 7d -> triggers auto-extension on next request)
	{UserIdx: -1, Token: "seed-anon-near-expiry-0002", ExpiresInHours: 6 * 24},
	// Anonymous, already expired (cleanup-job target)
	{UserIdx: -1, Token: "seed-anon-expired-0003", ExpiresInHours: -2},
	// User-bound, fresh -- maxmustermann (password "password123")
	{UserIdx: 0, Token: "seed-user-max-token-0010", ExpiresInHours: 30 * 24},
	// User-bound, fresh -- ghostUser01 (password "ghostpass", no email)
	{UserIdx: 10, Token: "seed-user-ghost-token-0011", ExpiresInHours: 30 * 24},
	// User-bound, fresh -- admin (password "password123")
	{UserIdx: 12, Token: "seed-user-admin-token-0012", ExpiresInHours: 30 * 24},
}
