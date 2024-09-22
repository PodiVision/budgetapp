document.getElementById('transaction-form').addEventListener('submit', async function (event) {
    event.preventDefault();

    const typ = document.getElementById('typ').value;
    const betrag = parseFloat(document.getElementById('betrag').value);
    const kategorie = document.getElementById('kategorie').value;

    const transaktion = { typ: typ, betrag: betrag, kategorie: kategorie };

    // Angepasst, um die korrekten Typen "Einnahme" und "Ausgabe" zu berücksichtigen
    const url = typ === 'Einnahme' ? '/einnahmen' : '/ausgaben';
    try {
        const response = await fetch(`http://localhost:8081${url}`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(transaktion)
        });

        if (response.ok) {
            updateSummary();
            fetchTransaktionen(); // Aktualisiere die Transaktionsliste nach dem Hinzufügen
            document.getElementById('transaction-form').reset();
        } else {
            alert('Hinzufügen der Transaktion fehlgeschlagen');
        }
    } catch (error) {
        console.error('Error:', error);
        alert('Es ist ein Fehler aufgetreten');
    }
});

async function updateSummary() {
    try {
        const response = await fetch('http://localhost:8081/zusammenfassung');
        const data = await response.json();

        // Anzeige der Gesamtsummen
        document.getElementById('gesamteinnahmen').textContent = data.Gesamteinnahmen.toFixed(2);
        document.getElementById('gesamtausgaben').textContent = data.Gesamtausgaben.toFixed(2);
        document.getElementById('kontostand').textContent = data.Kontostand.toFixed(2);

        // Tabelle für den aktuellen Stand aktualisieren
        const tabelle = document.getElementById('daten-tabelle');
        tabelle.innerHTML = `
            <tr>
                <td>Aktuell</td> <!-- Zeiträume müssen evtl. noch dynamisch angepasst werden -->
                <td>${data.Kontostand.toFixed(2)}</td>
                <td>${data.Gesamteinnahmen.toFixed(2)}</td>
                <td>${data.Gesamtausgaben.toFixed(2)}</td>
            </tr>
        `;
    } catch (error) {
        console.error('Fehler beim Aufrufen der Zusammenfassung:', error);
    }
}

// Funktion zum Abrufen und Anzeigen aller Transaktionen
async function fetchTransaktionen() {
    try {
        const response = await fetch('http://localhost:8081/transaktionen');
        const data = await response.json();

        const tabelle = document.getElementById('transaktionen-tabelle').getElementsByTagName('tbody')[0];
        tabelle.innerHTML = ''; // Clear existing rows
        data.forEach(transaktion => {
            const row = tabelle.insertRow();
            row.innerHTML = `
                <td>${transaktion.typ}</td>
                <td>${transaktion.betrag}</td>
                <td>${transaktion.kategorie}</td>
                <td><button onclick="deleteTransaktion(${transaktion.id})">Löschen</button></td>
            `;
        });
    } catch (error) {
        console.error('Fehler beim Abrufen der Transaktionen:', error);
    }
}

// Funktion zum Löschen einer Transaktion
async function deleteTransaktion(id) {
    if (!confirm('Möchtest du diese Transaktion wirklich löschen?')) {
        return;
    }
    console.log('Löschvorgang bestätigt für ID:', id);
    
    try {
        const response = await fetch(`http://localhost:8081/transaktionen/delete?id=${id}`, {
            method: 'DELETE'  // Wichtiger Punkt: DELETE Methode
        });

        if (response.ok) {
            fetchTransaktionen(); // Aktualisiere die Liste nach dem Löschen
            updateSummary(); // Aktualisiere auch die Zusammenfassung
        } else {
            alert('Löschen der Transaktion fehlgeschlagen');
        }
    } catch (error) {
        console.error('Fehler beim Löschen der Transaktion:', error);
    }
}

updateSummary();
fetchTransaktionen();
