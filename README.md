# Mattermost Juegos

Minijuegos para grupos en Mattermost, en castellano. Se lanzan con `/juegos` (o con el botón 🎲 de la
cabecera del canal) y se juegan en el propio mensaje de la partida, sin salir del canal.

| Juego | Cómo va |
|---|---|
| 🤪 **Respuestas absurdas** | Cada jugador completa dos frases en secreto; cada frase la responden dos personas y el canal vota a ciegas la más graciosa. 100 puntos por voto y 100 más si te llevas todos. |
| 🖼️ **Pie de foto** | Sale un GIF al azar de Giphy, cada uno escribe un pie y se vota a ciegas. |
| 🎨 **Teléfono dibujado** | Cada uno escribe una frase, el siguiente la dibuja en una pizarra, el siguiente describe el dibujo... Al final se ven todas las cadenas, se vota el mejor momento y los dibujos quedan en el hilo. |
| 🔢 **Número único más bajo** | Todos eligen en secreto un número del 1 al 20; gana el más bajo que nadie repita. |
| 🔴 **4 en raya** | Uno contra uno, por turnos; gana quien junte cuatro fichas en línea. |
| 🚢 **Hundir la flota** | Uno contra uno en un tablero de 10×10 con la flota española (un barco de 4, dos de 3, tres de 2 y cuatro de 1, sin tocarse). Si tocas, vuelves a disparar. |
| 🃏 **Guiñote** | Cuatro por parejas con baraja española y las reglas de la Asociación Aragonesa de Guiñote: cantes de 20 y 40, tute, cambiar el 7, arrastre (asistir y montar o fallar a la carta del contrario), diez de últimas, 101 con 30 sin cantes, o vueltas. Gana el coto quien se lleve dos juegos. |
| 👹 **Doom** | El Doom de 1993 (episodio 1 shareware) en una ventana, si el administrador lo ha montado (ver *Doom*). Sin puntuación. |
| 👾 **Lienzo de temporada** | Un lienzo de píxeles para todo el servidor: cada persona pone un píxel cada pocos minutos. Al acabar el mes se publica y empieza otro. |

- **Dos ritmos**: *en directo* (fases de un minuto, para un descanso) o *asíncrona* (fases de horas,
  entre tarea y tarea; el bot avisa en el hilo cuando toca hacer algo).
- **Uno contra uno y guiñote**: se reta a alguien (`/juegos 4enraya @usuario`) o se deja abierta al
  primero que se apunte; si nadie se apunta a tiempo, juega el bot (fácil o difícil). En el guiñote,
  el bot ocupa los huecos. Quien no juega su turno a tiempo pierde la partida (en el guiñote, su
  pareja pierde el coto). Ganar solo a bots no cuenta para el ranking.
- **Ranking por temporadas**: cada mes, global para todo el servidor. Por partida se dan 5, 3 y 2
  puntos al podio y 1 al resto de participantes. `/juegos ranking` lo enseña y, al cerrar el mes, el
  podio y el lienzo salen en el canal de anuncios.
- Los resultados de cada partida se publican en su hilo, así que se ven también en el móvil (para
  jugar hace falta el navegador o la app de escritorio).

## Comandos

| | |
|---|---|
| `/juegos` | Menú de juegos (en la app móvil, la ayuda). |
| `/juegos <juego> [directo\|async] [@usuario]` | Crea la partida directamente: `absurdas`, `pie`, `telefono`, `numero`, `4enraya`, `flota`, `guinote`. Con `@usuario`, reta a esa persona (4 en raya y flota). |
| `/juegos lienzo` | Abre el lienzo de la temporada. |
| `/juegos ranking [AAAA-MM]` | Ranking de la temporada. |
| `/juegos frase <tipo> <texto>` | Añade una frase: `absurdas` (con un hueco `___` o una pregunta), `telefono` o `pie` (tema de búsqueda de GIF, en inglés). Un administrador las quita con `/juegos frase quitar <tipo> <texto>`. |

## Ajustes

| Ajuste | Por defecto | |
|---|---|---|
| Comando | `juegos` | |
| Canal de anuncios | vacío | `equipo/canal` del podio mensual. |
| Zona horaria | `Europe/Madrid` | Cuándo cambia la temporada y hora de los avisos. |
| Fase en directo | `60` s | Cada juego la multiplica (dibujar dura el doble). |
| Fase asíncrona | `4` h | |
| Clave de Giphy | vacía | Vacía: usa la del plugin *GIF commands* si está instalado. |
| Tamaño del lienzo | `48` | Píxeles por lado, desde la siguiente temporada. |
| Espera entre píxeles | `3` min | |
| Dirección de Doom | vacía | Dónde se sirve Doom (por ejemplo `/doom/`). Vacía: no sale en el menú. |

## Doom

El plugin no lleva Doom: abre en un iframe la dirección del ajuste. Tiene que servirse en el mismo
dominio que Mattermost (el iframe solo se permite desde el mismo origen) y conviene protegerla con
la sesión de Mattermost, por ejemplo en nginx:

```nginx
location /doom/ {
    auth_request /doom-auth;
    proxy_pass http://doom.example:8080/;
}
location = /doom-auth {
    internal;
    proxy_pass http://mattermost.example:8065/api/v4/users/me;
    proxy_pass_request_body off;
    proxy_set_header Content-Length "";
}
```

Para servirlo vale [cloudflare/doom-wasm](https://github.com/cloudflare/doom-wasm) (Chocolate Doom en
WebAssembly, GPL-2.0) con `DOOM1.WAD` shareware, arrancado con `-nomusic` y compilado sin
`SAFE_HEAP` (con emsdk 3.1.71 aborta con «alignment fault» al iniciar).

## Añadir un juego

Cada juego es un paquete en `server/games/<id>/` y un componente en `webapp/src/games/<id>.jsx`. El
motor (`server/engine`) se encarga de la sala de espera, los plazos de cada fase, los avisos, los
resultados en el hilo y el ranking.

1. **Servidor**: implementa `engine.Game`.
   - `Info`: nombre, emoji, descripción, mínimo y máximo de jugadores y si necesita sala de espera
     (`Lobby`: juegos por turnos o por parejas).
   - `Start(m, env)`: prepara el estado y la primera fase con `env.SetPhase(m, "fase", peso)`. El peso
     multiplica la duración base del modo.
   - `Act(m, userID, move, env)`: aplica una jugada. Si no vale, devuelve `engine.Refusal("texto para
     el jugador")`.
   - `Timeout(m, env)`: qué pasa cuando se acaba el tiempo de la fase (o el creador la salta).
   - `View(m, userID, env)`: lo que ve cada usuario. **Nunca** los secretos de los demás.
   - Los juegos por turnos (`Turns`) cuentan cada turno como una fase; con `Bots`, la sala de
     espera los rellena con `engine.Seats` (ids `bot:N`). `Options` son las opciones que elige el
     creador en la sala (el nivel del bot, a quién reta); se leen con `m.Option`. La base común de
     los uno contra uno está en `games/duel`.
   - Al acabar, `env.Finish(m, puntos, resumen, partes...)`. Las partes son respuestas en el hilo con
     ficheros (los dibujos del teléfono, por ejemplo).
   - Opcional, `Notice(m, env)`: el aviso del hilo al empezar cada fase en modo asíncrono.

   El estado propio se guarda en `m.Data` con `engine.Save` y se lee con `engine.Load[T]`. Los tests
   usan `engine/enginetest`, que juega sin Mattermost.
2. Añade una línea en `server/registry.go`.
3. **Webapp**: un componente que recibe `{match, view, send}` (`view` es lo que devuelve `View`;
   `send('move', jugada)` manda una jugada) y regístralo en `webapp/src/games/index.js`. Hay piezas
   comunes en `webapp/src/match/common.jsx` (`TextMove`, `Countdown`, `Players`) y una pizarra en
   `webapp/src/games/drawpad.jsx`.
4. Si usa frases, añade un pack en `server/content/` y su tipo en `contentKinds` (`server/content.go`).

## Añadir frases

Los textos van en `server/content/<tipo>[-<origen>].json` (`{"source": "...", "items": [...]}`), se
incluyen en el binario y se suman los de todos los ficheros de un tipo. Basta con un PR que añada
líneas o un fichero nuevo; el test comprueba que el formato es correcto. Lo que la gente añade con
`/juegos frase` se guarda en el servidor y se suma a lo del repo.

## Instalar

Requiere Mattermost 11.11 o posterior. Descarga el `.tar.gz` de la última release y súbelo en
**Consola del sistema → Complementos → Gestión de complementos**, o:

```sh
mmctl plugin install-url https://github.com/adriadam10/mattermost-plugin-juegos/releases/download/v0.2.0/com.github.adriadam10.juegos-0.2.0.tar.gz
mmctl plugin enable com.github.adriadam10.juegos
```

## Compilar

`./build.sh` pasa los tests y compila todo en contenedores de Go y Node (solo hace falta Docker).
Genera `dist/com.github.adriadam10.juegos-<versión>.tar.gz`. Al subir una etiqueta `vX.Y.Z` igual a
la `version` de `plugin.json`, se publica la release.

## Créditos

- `server/content/absurdas-cah.json`: cartas negras de [Cards Against Humanity](https://cardsagainsthumanity.com),
  [CC BY-NC-SA 2.0](https://creativecommons.org/licenses/by-nc-sa/2.0/). Las traducciones al
  castellano son de la comunidad, recogidas en
  [mvarona/CartasContraLaHumanidad](https://github.com/mvarona/CartasContraLaHumanidad); aquí
  están filtradas y retocadas. Ese fichero mantiene su licencia (uso no comercial).
- Baraja del guiñote (`server/cards/`): naipes de Heraclio Fournier de 1878, de dominio público
  ([Wikimedia Commons](https://commons.wikimedia.org/wiki/Category:Heraclio_Fournier%E2%80%99s_1878_card_deck)).
- Reglas del guiñote: [reglamento de la Asociación Aragonesa de Guiñote](https://www.xn--guiarte-6za.es/reglamentacion)
  y [Wikipedia](https://es.wikipedia.org/wiki/Gui%C3%B1ote).
- Paleta del lienzo: [PICO-8](https://www.lexaloffle.com/pico-8.php).
- GIFs: [Giphy](https://giphy.com).

## Licencia

El código es MIT; el contenido de terceros mantiene su licencia (ver *Créditos*).
