(function () {
  // ── URL params from lobby redirect ──────────────────────────────────────────
  const params     = new URLSearchParams(location.search);
  const myUsername = params.get('me')         || 'You';
  const opponent   = params.get('opponent')   || 'Opponent';
  const firstMove  = params.get('first_move') || '';
  const matchId    = params.get('match_id')   || '0';

  const amWhite = myUsername === firstMove;
  let myTurn    = amWhite;   // white moves first
  let selected  = null;      // { x, y } or null
  let lastMove  = null;      // { from:[x,y], to:[x,y] } for highlight
  let ws        = null;

  // boardState[x][y] = { type, mine } | null
  // Both players work in their own perspective: y=0 own back rank, y=7 opponent
  const boardState = Array.from({ length: 8 }, () => new Array(8).fill(null));
  const BACK_RANK  = amWhite
    ? ['Rook','Knight','Bishop','Queen','King','Bishop','Knight','Rook']
    : ['Rook','Knight','Bishop','King','Queen','Bishop','Knight','Rook'];

  // Piece colors for this player
  const myColor  = amWhite ? 'w' : 'b';
  const oppColor = amWhite ? 'b' : 'w';

  // ── Piece images (Wikimedia Commons, public domain, Colin M.L. Burnett) ───
  const IMG = {
    w: {
      King:   'https://upload.wikimedia.org/wikipedia/commons/4/42/Chess_klt45.svg',
      Queen:  'https://upload.wikimedia.org/wikipedia/commons/1/15/Chess_qlt45.svg',
      Rook:   'https://upload.wikimedia.org/wikipedia/commons/7/72/Chess_rlt45.svg',
      Bishop: 'https://upload.wikimedia.org/wikipedia/commons/b/b1/Chess_blt45.svg',
      Knight: 'https://upload.wikimedia.org/wikipedia/commons/7/70/Chess_nlt45.svg',
      Pawn:   'https://upload.wikimedia.org/wikipedia/commons/4/45/Chess_plt45.svg',
    },
    b: {
      King:   'https://upload.wikimedia.org/wikipedia/commons/f/f0/Chess_kdt45.svg',
      Queen:  'https://upload.wikimedia.org/wikipedia/commons/4/47/Chess_qdt45.svg',
      Rook:   'https://upload.wikimedia.org/wikipedia/commons/f/ff/Chess_rdt45.svg',
      Bishop: 'https://upload.wikimedia.org/wikipedia/commons/9/98/Chess_bdt45.svg',
      Knight: 'https://upload.wikimedia.org/wikipedia/commons/e/ef/Chess_ndt45.svg',
      Pawn:   'https://upload.wikimedia.org/wikipedia/commons/c/c7/Chess_pdt45.svg',
    }
  };

  // Preload all piece images
  function preloadImages() {
    for (const c of ['w', 'b'])
      for (const t of Object.keys(IMG[c]))
        new Image().src = IMG[c][t];
  }

  // ── Board initialisation ────────────────────────────────────────────────────
  function initBoard() {
    for (let x = 0; x < 8; x++) {
      boardState[x][0] = { type: BACK_RANK[x], mine: true  };
      boardState[x][1] = { type: 'Pawn',       mine: true  };
      boardState[x][6] = { type: 'Pawn',       mine: false };
      boardState[x][7] = { type: BACK_RANK[x], mine: false };
    }
  }

  // ── Rendering ───────────────────────────────────────────────────────────────
  function renderBoard() {
    const bd = document.getElementById('board');
    bd.innerHTML = '';

    for (let screenRow = 0; screenRow < 8; screenRow++) {
      const y = 7 - screenRow;
      for (let x = 0; x < 8; x++) {
        const sq = document.createElement('div');
        sq.className = 'square ' + ((x + y) % 2 === 0 ? 'dark' : 'light');
        sq.dataset.x = x;
        sq.dataset.y = y;

        // Last-move highlight
        if (lastMove) {
          const [fx, fy] = lastMove.from;
          const [tx, ty] = lastMove.to;
          if ((x === fx && y === fy) || (x === tx && y === ty))
            sq.classList.add('last-move');
        }

        // Selected highlight
        if (selected && selected.x === x && selected.y === y)
          sq.classList.add('selected');

        // Piece
        const piece = boardState[x][y];
        if (piece) {
          const color = piece.mine ? myColor : oppColor;
          const img   = document.createElement('img');
          img.className = 'piece-img';
          img.src       = IMG[color][piece.type];
          img.alt       = piece.type;
          img.draggable = false;
          sq.appendChild(img);
        }

        sq.addEventListener('click', () => onSquareClick(x, y));
        bd.appendChild(sq);
      }
    }
  }

  // ── Interaction ─────────────────────────────────────────────────────────────
  function onSquareClick(x, y) {
    if (!myTurn) return;

    const piece = boardState[x][y];

    if (selected) {
      // Deselect
      if (selected.x === x && selected.y === y) {
        selected = null;
        renderBoard();
        return;
      }
      // Re-select another own piece
      if (piece && piece.mine) {
        selected = { x, y };
        renderBoard();
        return;
      }
      // Attempt move
      sendMove(selected.x, selected.y, x, y);
      selected = null;
      renderBoard();
    } else {
      if (piece && piece.mine) {
        selected = { x, y };
        renderBoard();
      }
    }
  }

  function sendMove(fx, fy, tx, ty) {
    if (!ws || ws.readyState !== WebSocket.OPEN) return;
    ws.send(JSON.stringify({
      type: 'MOVE',
      data: { pos_from: [fx, fy], pos_to: [tx, ty] }
    }));
  }

  // ── State update ────────────────────────────────────────────────────────────
  function applyMove(posFrom, posTo) {
    const [fx, fy] = posFrom;
    const [tx, ty] = posTo;
    const movedPiece = boardState[fx][fy];
    boardState[tx][ty] = movedPiece;
    boardState[fx][fy] = null;
    lastMove = { from: [fx, fy], to: [tx, ty] };
    return movedPiece;
  }

  function readPos(pos) {
    if (!pos) return null;
    if (Array.isArray(pos) && pos.length === 2) return pos;
    if (typeof pos.X === 'number' && typeof pos.Y === 'number') return [pos.X, pos.Y];
    if (typeof pos.x === 'number' && typeof pos.y === 'number') return [pos.x, pos.y];
    return null;
  }

  function getSpecialMoveEffect(lastMovePayload) {
    if (!lastMovePayload) return null;
    const effect = lastMovePayload.SpecialMoveEffect ||
      lastMovePayload.specialMoveEffect ||
      lastMovePayload.special_move_effect ||
      null;
    if (!effect) return null;
    const effectType = effect.SpecialMoveType || effect.specialMoveType || effect.special_move_type;
    if (!effectType) return null;
    return effect;
  }

  function applySpecialMoveEffect(effect, movedPiece) {
    if (!effect) return;

    const from = readPos(effect.PosFrom || effect.pos_from);
    const to = readPos(effect.PosTo || effect.pos_to);
    const effectType = effect.SpecialMoveType || effect.specialMoveType || effect.special_move_type;
    let pieceType = effect.PieceType || effect.piece_type || effect.pieceType;
    if (!pieceType && effectType === 'promotion') {
      pieceType = 'Queen';
    }

    const fromPiece = from ? boardState[from[0]][from[1]] : null;
    const mine = (fromPiece && fromPiece.mine !== undefined)
      ? fromPiece.mine
      : (movedPiece && movedPiece.mine !== undefined ? movedPiece.mine : true);

    if (from) {
      boardState[from[0]][from[1]] = null;
    }

    if (to && to[0] !== 99 && to[1] !== 99) {
      if (pieceType) {
        boardState[to[0]][to[1]] = { type: pieceType, mine };
      }
    }
  }

  // ── UI helpers ──────────────────────────────────────────────────────────────
  function updateTurnUI() {
    const ti = document.getElementById('turnIndicator');
    if (myTurn) {
      ti.textContent = 'Your turn';
      ti.className   = 'turn-indicator your-turn';
    } else {
      ti.textContent = "Opponent's turn";
      ti.className   = 'turn-indicator';
    }
  }

  function showStatus(msg, duration) {
    const st = document.getElementById('status');
    st.textContent = msg;
    if (duration) setTimeout(() => { st.textContent = ''; }, duration);
  }

  // ── WebSocket ───────────────────────────────────────────────────────────────
  function connect() {
    const proto = location.protocol === 'https:' ? 'wss:' : 'ws:';
    ws = new WebSocket(`${proto}//${location.host}/ws/match/${matchId}`);

    ws.onopen = () => showStatus('');

    ws.onmessage = (evt) => {
      try {
        const msg = JSON.parse(evt.data);
        if (msg.type === 'MOVE_MADE') {
          const d = msg.data;
          const movedPiece = applyMove(d.last_move.pos_from, d.last_move.pos_to);
          const effect = getSpecialMoveEffect(d.last_move);
          if (effect) {
            applySpecialMoveEffect(effect, movedPiece);
          }
          myTurn = (d.turn_now === myUsername);
          selected = null;
          renderBoard();
          updateTurnUI();
        } else if (msg.type === 'ERROR') {
          const text = typeof msg.data === 'string' ? msg.data : JSON.stringify(msg.data);
          showStatus(text, 3000);
        }
      } catch (e) {}
    };

    ws.onclose = () => {
      showStatus('Disconnected — reconnecting…');
      setTimeout(connect, 3000);
    };
  }

  // ── Boot ────────────────────────────────────────────────────────────────────
  window.addEventListener('load', () => {
    document.getElementById('opponentName').textContent = opponent;
    document.getElementById('yourName').textContent     = myUsername;
    document.getElementById('opponentRole').textContent = amWhite ? 'Black' : 'White';
    document.getElementById('yourRole').textContent     = amWhite ? 'White' : 'Black';
    preloadImages();
    initBoard();
    updateTurnUI();
    renderBoard();
    connect();
  });
})();
