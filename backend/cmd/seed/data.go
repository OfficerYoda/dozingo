package main

// Seed credentials (for local dev only):
//   maxmustermann      / password123
//   lena.schmidt       / securePass!
//   timoWerner42       / timoSecret42
//   ghostUser01        / ghostpass        (no email)
//   anon_student       / anonpass         (no email)
//
// Seed session tokens (for local dev only):
//   seed-anon-fresh-token-0001      (anonymous, ~30d valid)
//   seed-anon-near-expiry-0002      (anonymous, <7d -> triggers extension)
//   seed-anon-expired-0003          (anonymous, expired -> cleanup target)
//   seed-user-max-token-0010        (bound to maxmustermann)
//   seed-user-ghost-token-0011      (bound to ghostUser01)

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
type gameData struct {
	PlayerIdx  int
	SessionIdx int
	BoardIdx   int
	Status     string // "active", "completed", or "abandoned"
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
	{Username: "ghostUser01", Email: ""},
	{Username: "anon_student", Email: ""},
}

// boards defines 15 boards with mixed sizes (3-7), each linked to an author.
var boards = []boardData{
	// Size 3 boards (n²=9, pool=12 cells each)
	{Title: "Mathe 1 Bingo", Description: "Klassiker aus der Mathe 1 Vorlesung", Size: 3, AuthorIdx: 0},
	{Title: "Lineare Algebra Klassiker", Description: "Die besten Momente aus LinAlg", Size: 3, AuthorIdx: 1},
	{Title: "Statistik Grundlagen", Description: "Statistik Vorlesung Bingo", Size: 3, AuthorIdx: 4},

	// Size 4 boards (n²=16, pool=21 cells each)
	{Title: "Theoretische Informatik", Description: "Turingmaschinen und mehr", Size: 4, AuthorIdx: 2},
	{Title: "Physik Vorlesung", Description: "Experimente die schiefgehen", Size: 4, AuthorIdx: 3},
	{Title: "BWL Einführung", Description: "PowerPoint-Schlachten und Buzzwords", Size: 4, AuthorIdx: 5},

	// Size 5 boards (n²=25, pool=32 cells each)
	{Title: "Algorithmen und Datenstrukturen", Description: "O-Notation und Rekursion", Size: 5, AuthorIdx: 6},
	{Title: "Softwaretechnik Bingo", Description: "UML und Agile Methoden", Size: 5, AuthorIdx: 7},
	{Title: "Datenbanken Vorlesung", Description: "SQL, Normalformen und Joins", Size: 5, AuthorIdx: 0},

	// Size 6 boards (n²=36, pool=45 cells each)
	{Title: "Betriebssysteme Chaos", Description: "Deadlocks und Race Conditions", Size: 6, AuthorIdx: 8},
	{Title: "Rechnernetze Marathon", Description: "OSI-Modell und TCP/IP", Size: 6, AuthorIdx: 9},
	{Title: "Digitaltechnik Vorlesung", Description: "Flip-Flops und Gatter", Size: 6, AuthorIdx: 1},

	// Size 7 boards (n²=49, pool=62 cells each)
	{Title: "Compilerbau Endgegner", Description: "Parser, Lexer und Optimierung", Size: 7, AuthorIdx: 3},
	{Title: "Verteilte Systeme Bingo", Description: "CAP-Theorem und Konsens", Size: 7, AuthorIdx: 6},
	{Title: "Maschinelles Lernen Bingo", Description: "Gradient Descent und Overfitting", Size: 7, AuthorIdx: 2},
}

// cellPhrases contains German lecture bingo phrases for each board index.
// Pool sizes: size 3 -> 12, size 4 -> 21, size 5 -> 32, size 6 -> 45, size 7 -> 62
var cellPhrases = map[int][]string{
	// Board 0: "Mathe 1 Bingo" (size 3, 12 cells)
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
	},

	// Board 1: "Lineare Algebra Klassiker" (size 3, 12 cells)
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
		"Gauß-Algorithmus zum dritten Mal",
		"Prof dreht sich zur falschen Tafel",
		"Inverse existiert nicht",
	},

	// Board 2: "Statistik Grundlagen" (size 3, 12 cells)
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

	// Board 4: "Physik Vorlesung" (size 4, 21 cells)
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
	},

	// Board 5: "BWL Einführung" (size 4, 21 cells)
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
		"Greedy-Algorithmus vorgestellt",
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
		"Scheduling-Algorithmus verglichen",
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

	// Board 12: "Compilerbau Endgegner" (size 7, 62 cells)
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
		"Prof zeichnet DFA für Scanner",
		"Backpatching erklärt",
		"'Das ist ein Spill'",
		"Prof sagt 'Calling Convention'",
		"Instruction Selection Phase",
		"'Da brauchen wir einen Fixpunkt'",
		"Prof vergisst Epsilon-Übergang",
		"Inline-Expansion erwähnt",
		"'Das ist Data Flow Analysis'",
		"Prof schreibt Abstract Machine Code",
		"Tail-Call-Optimierung erklärt",
		"'Das ist Strength Reduction'",
		"Jemand schläft ein beim Parsertabelle-Ausfüllen",
		"Prof sagt 'Das ist der schwierigste Teil'",
		"Closure einer Itemmenge berechnet",
		"'Das handle ich in der Übung'",
		"Prof zeichnet Abhängigkeitsgraph",
	},

	// Board 13: "Verteilte Systeme Bingo" (size 7, 62 cells)
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
		"Prof sagt 'Retry mit Backoff'",
		"'Das ist ein Ordering-Problem'",
		"Sharding wird erklärt",
		"Prof zeichnet Partitionierung",
		"'Exactly-Once-Delivery gibt es nicht'",
		"CRDT wird vorgestellt",
		"Prof sagt 'Causal Consistency'",
		"'Das ist ein Hot-Spot-Problem'",
		"Jemand fragt nach Blockchain",
		"Prof seufzt",
		"Quorum-basierte Replikation erklärt",
		"'N/2 + 1 für Mehrheit'",
		"Prof zeigt Latenz-Diagramm",
		"'Das Netzwerk ist nicht zuverlässig'",
		"Failover-Strategie diskutiert",
		"Prof sagt 'Graceful Degradation'",
		"'Das ist ein Thundering-Herd-Problem'",
		"Event Sourcing erwähnt",
		"Prof zeichnet Zustandsmaschine",
		"'Das braucht ein Distributed Lock'",
		"Lease-Mechanismus erklärt",
		"Prof sagt 'Observability ist wichtig'",
		"'Distributed Tracing hilft'",
		"Bulkhead-Pattern vorgestellt",
		"Prof vergisst Pfeil im Diagramm",
		"'Das ist ein Consistency-Latency-Tradeoff'",
		"Jemand fragt 'Reicht nicht eine Datenbank?'",
		"Prof sagt 'Es kommt darauf an'",
		"Chaos Engineering erwähnt",
		"'Netflix hat das erfunden'",
	},

	// Board 14: "Maschinelles Lernen Bingo" (size 7, 62 cells)
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
		"k-Means Algorithmus vorgestellt",
		"Dimensionsreduktion erwähnt",
		"'PCA ist Ihr Freund'",
		"Prof sagt 'Regularisierung'",
		"Dropout wird erklärt",
		"'Batch Size beeinflusst das Training'",
		"Prof zeigt Loss-Kurve",
		"'Das Modell konvergiert'",
		"Aktivierungsfunktion erklärt",
		"'ReLU statt Sigmoid'",
		"Prof sagt 'Vanishing Gradient'",
		"CNN für Bilderkennung erwähnt",
		"'Das ist Transfer Learning'",
		"Prof zeigt vortrainiertes Modell",
		"'Preprocessing ist 80% der Arbeit'",
		"One-Hot-Encoding für Kategorien",
		"Prof sagt 'Ensemble-Methoden'",
		"Precision vs Recall diskutiert",
		"'F1-Score als Kompromiss'",
		"Prof zeichnet ROC-Kurve",
		"'AUC sollte nahe 1 sein'",
		"Jemand fragt nach ChatGPT",
		"Prof sagt 'Das ist ein anderes Thema'",
		"Recurrent Neural Network erwähnt",
		"'Attention is All You Need'",
		"Prof zeigt MNIST-Beispiel",
		"'Das sind zu viele Epochen'",
		"Data Augmentation erklärt",
		"Prof sagt 'Generalisierung ist das Ziel'",
		"'Das Modell ist unterbestimmt'",
		"Feature Engineering diskutiert",
		"Prof zeigt Heatmap",
		"'Korrelation der Features prüfen'",
		"Gradient Boosting vorgestellt",
		"'XGBoost gewinnt Kaggle-Wettbewerbe'",
		"Prof sagt 'Explainability'",
		"SHAP-Values erwähnt",
		"'Das ist ein Black-Box-Modell'",
		"Prof zeichnet Decision Boundary",
		"'Nächste Woche: Deep Learning'",
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
}

// games defines game sessions. Each game is played by a user on a board.
var games = []gameData{
	// User 0 plays board 0 (Mathe 1 Bingo, size 3 -> 9 cells) - completed
	{PlayerIdx: 0, SessionIdx: -1, BoardIdx: 0, Status: "completed"},
	// User 1 plays board 0 (Mathe 1 Bingo) - active
	{PlayerIdx: 1, SessionIdx: -1, BoardIdx: 0, Status: "active"},
	// User 2 plays board 3 (Theoretische Informatik, size 4 -> 16 cells) - active
	{PlayerIdx: 2, SessionIdx: -1, BoardIdx: 3, Status: "active"},
	// User 3 plays board 6 (Algorithmen und Datenstrukturen, size 5 -> 25 cells) - abandoned
	{PlayerIdx: 3, SessionIdx: -1, BoardIdx: 6, Status: "abandoned"},
	// User 5 plays board 1 (Lineare Algebra Klassiker, size 3 -> 9 cells) - active
	{PlayerIdx: 5, SessionIdx: -1, BoardIdx: 1, Status: "active"},

	// Game 5: Anonymous (session 0 = "seed-anon-fresh-token-0001") plays board 0
	// (Mathe 1 Bingo, size 3 -> 9 cells) - active
	{PlayerIdx: -1, SessionIdx: 0, BoardIdx: 0, Status: "active"},
	// Game 6: Anonymous (session 1 = "seed-anon-near-expiry-0002") plays board 3
	// (Theoretische Informatik, size 4 -> 16 cells) - abandoned
	{PlayerIdx: -1, SessionIdx: 1, BoardIdx: 3, Status: "abandoned"},
	// Game 7: maxmustermann playing through his bound session
	// (session 3 = "seed-user-max-token-0010"). Has both player_id and session_id.
	// Board 8 (Datenbanken Vorlesung, size 5 -> 25 cells) - completed
	{PlayerIdx: 0, SessionIdx: 3, BoardIdx: 8, Status: "completed"},
}

// gameCells defines the game_cells for each game index.
// Each game picks n² cells from the board's cell pool and assigns positions.
// Position is 0-indexed in row-major order.
var gameCells = map[int][]gameCellData{
	// Game 0: User 0 on board 0 (Mathe 1, size 3 -> 9 cells, completed with marks)
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
	},

	// Game 1: User 1 on board 0 (Mathe 1, size 3 -> 9 cells, active with some marks)
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

	// Game 4: User 5 on board 1 (Lineare Algebra, size 3 -> 9 cells, active)
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
	},

	// Game 5: Anonymous on board 0 (Mathe 1, size 3 -> 9 cells, active, a couple of marks)
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
	{UserIdx: 10, Password: "ghostpass"}, // user with no email
	{UserIdx: 11, Password: "anonpass"},  // user with no email
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
}
