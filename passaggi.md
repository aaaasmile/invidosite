# Invido Site
Tool per generare il sito completamente statico invido.it. Questo rimpiazza webgen.


### Creare il contenuto Html
Nella sottodirectory Content metto tutti i vari post in una directory singola.
Qui ho il mio file _mdhtml_ e i vari files delle immagini relative al post. 
Il file mdhtml contiene solo la parte all'interno del tag article principale.

L'ho chiamato mdhtml in quanto, come nei files md, c'è una parte di dati in testa seguita da
una parte in html. Il testo dell'articolo lo edito in html per avere la massima flessibilità
di generazione del codice html. La parte di dati mi serve solo per quei campi che hanno bisogno
di un valore eplicito, altrimenti ci sono spazi per delle ambiguità nella generazione del codice. 

Per quanto mi riguarda, usando un html strettamente semantico, non vedo il bisogno di editare il post
in md con tutte le restrizioni del caso. Qui vengono generati i files statici della directory _posts_.

I contenuti della directory _pages_ li creo manualmente.

### Editare un post
In src/Content si lancia il watcher, che agisce quando il file mdhtml cambia. Oppure viene inserita un'immagine o rinominata. Per esempio se voglio modificare il file 24-11-08-ProssimaGara.mdhtml:

    go run .\main.go -editpost -date "2023-01-04"
Poi mentre cambio il file 24-11-08-ProssimaGara.mdhtml, mi piazzo col browser su:

    http://localhost:5572/posts/2024/11/24-11-08-ProssimaGara/
per vedere il cambiamenti nell'output statico dopo un browser reload


## Formato mdhtml
È un file che ha una sezione per i dati come i files md e una con il contenuto.
Nella parte del contenuto uso il codice html. Per velocizzare la generazione dei tag, uso
un preprocessor che mi genera un codice html. Esso supporta queste macro:

- link
- linkcap
- linknext
- linkimgnext
- img_link_run
- figstack
- youtube
- latest_posts
- archive_posts
- tag_posts
- single_taggedposts

Tutti i comandi sono compresi tra parantesi quadre. La lista la trovo nel file _lexer-builtin-func.go_.

### Link
Il comando _link_ serve per avere un <a href> con il link url uguale al testo mostrato.
Esempio:

    [link 'https://wien-rundumadum-2024-130k.legendstracking.com/']
genera:

    <a href='https://wien-rundumadum-2024-130k.legendstracking.com/' traget="_blank">https://wien-rundumadum-2024-130k.legendstracking.com/ </a> 

### Link caption
Un link che però ha anche la caption.
Esempio:

    [linkcap 'Tracker', 'https://wien-rundumadum-2024-130k.legendstracking.com/']
genera:

    <a href='https://wien-rundumadum-2024-130k.legendstracking.com/' traget="_blank">Tracker</a> 

### Link Next
Come linkcap ma non apre una  nuova page

    [linknext 'Tracker', 'https://wien-rundumadum-2024-130k.legendstracking.com/']

    <a href='https://wien-rundumadum-2024-130k.legendstracking.com/'>Tracker</a> 

### linkimgnext
Come linknext, ma al posto del link di testo ho un'immagine

    [linkimgnext, 'foto02_320.jpg', '/pages/tressette']
che genera:

    <a href="/pages/tressette"><img src="foto02_320.jpg" /></a>

### figstack
Serve per creare velocemente una galleria di immagini.
Esempio:

    [figstack
        'AustriaBackyardUltra2024011.jpg', 'Partenza mondiale Backyard',
        'backyard_award.png', 'Certificato finale'
    ]
Ogni coppia è rappresentata dal nome del file dell'immagine integrale e dal titolo.
Il codice html generato lo trovo di seguito. Col file dell'immagine integrale 
cosidero per dato il file d'immagine in formato ridotto di larghezza 320 pixel.

### youtube
Genera l'iframe che serve per contenere il video player di youtube.  
Qui c'è solo un argomento.
Esempio:

    [youtube 'IOP7RhDnLnw'] 
Dove IOP7RhDnLnw è il video ID su youtube.

Per centrare il video come le figure:

    <figure>
      [youtube 'vsC8SXH6Ffg']
      <figcaption>Il video ufficiale della gara</figcaption>
    </figure>
Oppure

    <p class="center">
        [youtube 'vsC8SXH6Ffg']
    </p>

## Immagini (html creato da figstack)
Quondo ho una serie di immagini da inserire nel post, uso il seguente html:

    <section class="vertstack">
      <figure>
        <a href="tabella.png"><img src="tabella_320.png" alt="Tabella finale" /></a>
        <figcaption>Tabella finale</figcaption>
      </figure>
      <figure>
        <a href="partenza.jpg"><img src="partenza_320.jpg" alt="Appena partiti" /></a>
        <figcaption>Appena partiti</figcaption>
      </figure>
    </section>
Per questo ho bisogno delle immagini in formato ridotto _xxx\_320_.
Qui si vede che le immagini sono nella stessa directory del post in quanto non riutilizzo mai
la stessa immagine in un altro post.

## latest_posts
Nella pagina proncipale ho bisogno di un sommario degli ultimi post. Per questo uso la macro:

    [latest_posts 'Invido Site', '7']

Dove 'Invido Site' rappresenta il titolo e '7' è il numero dei post da mettere.
Il risultato è un html con la lista degli ultimi 7 post. L'elenco viene creato leggendo il
database.

## archive_posts
Mi genera una pagina d'archivio con un link a tutti i posts.

## tag_posts
Mi genera un tratto di html con un link a tutti tags che ho utilizzato nei vari post.
Questo lo metto di solito nella pagina main.

## single_taggedposts
Esempio

    [single_taggedposts 'MaratonaGara']
genera una list di post che hanno il tag 'MaratonaGara'.
Questo di solito lo metto in una pagina dedicata apposta al tag in questione (esempio page-src/tags/WRU.mdhtml). 
Nota che queste pagine di Tag servono poi per creare la pagina del singolo tag che si crea con buildpages.
La pagina mdhtml del Tag (mdhtml source) singolo viene creata automaticamente con il flag buildtags.
Riassunto:
- scancontent aggiorna il db
- buildtags crea i sorgenti mshtml, ma non aggiorna il db
- buildpages crea le pagine html di tutti i tags partendo dai sorgenti e il db (tabelle tags e tags_to_post)

Se la tabella tags_to_post contiene dei dati invalidi per quanto riguarda post_id, allora i single_taggedposts 
generati per lo specifico tag non coincidono. Quindi quando si cancella un post, va cancellato prima 
il record in tags_to_post. L'integrità refernziale in sqlite funziona? Si, ma non è abilitata di default.
Il comando in golang:

    sql.Open("sqlite3", fmt.Sprintf("%s?_foreign_keys=1", dbname))
apre il db con il check dell'integrità referenziale.

    -- Find tags_to_post records pointing to non-existent posts
    SELECT ttp.* 
    FROM tags_to_post ttp 
    LEFT JOIN post p ON ttp.post_id = p.id 
    WHERE p.id IS NULL;
    
## img_link_run
Per lanciare i giochi nel browser ho bisogno di un'immagine di sfondo, un bottone
ed un link nel quale parte il gioco nel browser. Esempio

    [img_link_run 'foto01_320.jpg', 'https://cup.invido.it/#/', 'RUN in Browser']

Il codice generato è questo:

    <a href="https://cup.invido.it/#/" class="image-link">
      <img src="foto01_320.jpg" class="image-run" />
      <button type="button" class="run-button-always">RUN in Browser</button>
    </a>

### config_custom.toml
È il file che mi esegue un ovveride del file config.toml. 
Al momento non ha un utilizzo particolare.


## Database
Ho separato un database per il sito.
Il database mi serve per creare i link. 
La ricerca non l'ho ancora implementata, ma potrei provare con il target js invece di db per
non dover avere un server che gestisca il sito (nota che non ho previsto commenti per il sito).

## Ricreare il sito da zero
Se per caso devo ricreare il sito (links, pages e posts)

    .\invido-site.exe -rebuildall
Nota che alla fine del processo è necessario lanciare anche 

    .\invido-site.exe -all4sync
per aggiornare i tags nel main.

## Creare una nuova page

   .\invido-site.exe -newpage "faq" -date "2025-12-23" -watch

## Cambiare una page

    .\invido-site.exe -editpage -name "cuperativa"

## Creare un nuovo Post (New)
Al momento il processo funziona con Visual Code (profilo Edit Post).
Il database sarebbe meglio scaricarlo da current su invido.it.
Per il nuovo post:

    .\invido-site.exe  -newpost "Storia di Breda" -date "2016-03-05" -watch

Ora edito il nuovo file mdhtml e vedo subito il risultato (nell'esempio di sopra su http://localhost:5572/posts/2025/04/17/25-04-17-NuovoSito/).

### Sincronizza il nuovo Post
A questo punto, se voglio preparare tutti i files per il comando rsync, uso il seguente comando:

    .\invido-site.exe -all4sync
Ora apro WSL e lancio rsync (vedi sync_blog.sh)


### Cambiare un post già pubblicato (Edit)
Uso il flag -editpost. Per esempio:

    .\invido-site.exe -editpost -date "2025-11-30"

### Cambiare solo il main (per esempio per il live)

    .\invido-site.exe -buildmain

### all4sync

Questo è quello che esegue il flag -all4sync

1) Attualizzare i links

    .\invido-site.exe -scancontent

2) Creare i posts col feed xml

    .\invido-site.exe -buildposts

3) Creare le pages che sono cambiate 

    .\invido-site.exe -buildpages

4) Creare la main page sempre

    .\invido-site.exe -buildmain


In futuro, con la funzione "cerca", il sync del db con i dati della ricerca probabilmente sarà necessario.

### Archivio
La pagina archivio situata su content/page-src/archivio non viene creata automaticamente 
all'inizio. Essa viene solo aggiornata (invero è il [archive_posts] che viene cambiato). Quindi all'inizio uso:

    .\invido-site.exe -newpage "archivio" -date "2025-12-22" -watch
Poi in questa pagina posso mettere il titolo e il testo che voglio. L'archivio
viene aggiornato usando il tag:

    [archive_posts]
Al momento la pagina d'archivio è fissa nel codice sorgente al nome 'archivio'.

## Per vedere il sito creato

    cd .\launcher
    .\launcher.exe

### Tags
Se cambi un tag in una pagina mdhmtl bisogna ricreare la pagina del tag.
Il modo più semplice è quello di ricreare tutte le pagine dopo avere fatto uno scan
per aggiornare il db con il nuovo post nel tag 

    .\invido-site.exe -scancontent
    .\invido-site.exe -buildpages -force

Nota che il comando _-buildtags_ serve per creare il nuovo tag nel db e generare il file
mdhtml (è una nuova page) che raccoglie tutti i posts che contengono il tag.
Ne mio caso è la directory content/page-src/tags/. Nota che se la directory 
content/page-src/tags/ non esiste, il file sorgente di tag non viene generato.

Se il tag esiste già, nulla viene creato o modificato, tranne se la pagina ha un md5 obsoleto.

## TODO
- con il comando:
     go run .\main.go  -buildonepage -name "briscola"
non viene creato in modo corretto il file json di photos.json (function func (mp *MdHtmlProcess) parsedToHtml() error ). [DONE]
- dopo il rebuildall i tag nel main non vengono aggiornati, occorre lanciare all4sync 