(function () {
  // --- URL params set by lobby.js on MATCH_START ---
  const params      = new URLSearchParams(location.search);
  const myUsername  = params.get('me')         || 'You';
  const opponent    = params.get('opponent')   || 'Opponent';
  const firstMove   = params.get('first_move') || '';
  const matchId     = params.get('match_id')   || '0';

  const amWhite = myUsername === firstMove; // true → I play white, move first
  let myTurn    = amWhite;
  let selected  = null; // { x, y } or null
  let ws        = null;

  // boardState[x][y] = { type: string, mine: bool } | null
  // Coordinates are always in the receiver's own perspective:
  //   y=0 own back rank, y=1 own pawns, y=6 opp pawns, y=7 opp back rank
  // (the backend reverses coords for black via reverseMoveDTO, so both
  //  players always work in the same logical space)
  const boardState = Array.from({ length: 8 }, () => new Array(8).fill(null));

  const BACK_RANK  = ['Rook','Knight','Bishop','Queen','King','Bishop','Knight','Rook'];
  const SYMBOLS    = { Pawn:'♟', Rook:'♜', Knight:'♞', Bishop:'♝', Queen:'♛', King:'♚' };

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
    // Screen row 0 = logical rank 7 (opponent's back rank at top)
    for (let screenRow = 0; screenRow < 8; screenRow++) {
      const y = 7 - screenRow;
      for (let x = 0; x < 8; x++) {
        const sq = document.createElement('div');
        // a1 (x=0,y=0) is a dark square in standard chess: (x+y) even → dark
        sq.className = 'square ' + ((x + y) % 2 === 0 ? 'dark' : 'light');
        sq.dataset.x = x;
        sq.dataset.y = y;

        const piece = boardState[x][y];
        if (piece) {
          const pi = document.createElement('div');
          pi.className = 'piece ' + (piece.mine ? 'mine' : 'theirs');
          pi.textContent = SYMBOLS[piece.type] || '?';
          sq.appendChild(pi);
        }

        if (selected && selected.x === x && selected.y === y) {
          sq.classList.add('selected');
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
      if (selected.x === x && selected.y === y) {
        // Deselect
        selected = null;
        renderBoard();
        return;
      }
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
  // Both players receive last_move already in their own coordinate perspective
  function applyMove(posFrom, posTo) {
    const [fx, fy] = posFrom;
    const [tx, ty] = posTo;
    boardState[tx][ty] = boardState[fx][fy];
    boardState[fx][fy] = null;
  }

  // ── UI helpers ──────────────────────────────────────────────────────────────
  function updateTurnUI() {
    const ti = document.getElementById('turnIndicator');
    if (myTurn) {
      ti.textContent  = 'Your turn';
      ti.className    = 'turn-indicator your-turn';
    } else {
      ti.textContent  = "Opponent's turn";
      ti.className    = 'turn-indicator';
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
          applyMove(d.last_move.pos_from, d.last_move.pos_to);
          myTurn = (d.turn_now === myUsername);
          renderBoard();
          updateTurnUI();
        } else if (msg.type === 'ERROR') {
          const text = typeof msg.data === 'string' ? msg.data : JSON.stringify(msg.data);
          showStatus('⚠ ' + text, 3000);
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
    updateTurnUI();
    initBoard();
    renderBoard();
    connect();
  });
})();
