let ws = null;
let reconnectTimeout = null;
let currentUsername = null;

const statusElement = document.getElementById('status');
const playersListElement = document.getElementById('playersList');
const userNameElement = document.getElementById('userName');
const userAvatarElement = document.getElementById('userAvatar');

const adjectives = ['Swift', 'Clever', 'Bold', 'Cunning', 'Mighty', 'Sharp', 'Quick', 'Brave', 'Wise', 'Fearless'];
const animals = ['Rook', 'Knight', 'Bishop', 'Queen', 'Pawn', 'Eagle', 'Tiger', 'Lion', 'Fox', 'Wolf'];
// Use an illustrated style for high-quality avatars
const dicebearStyle = 'adventurer';

function generateUsername() {
    const adj = adjectives[Math.floor(Math.random() * adjectives.length)];
    const animal = animals[Math.floor(Math.random() * animals.length)];
    return `${adj} ${animal}`;
}

function hashString(str) {
    let hash = 0;
    for (let i = 0; i < str.length; i++) {
        hash = ((hash << 5) - hash) + str.charCodeAt(i);
        hash |= 0; // Convert to 32bit integer
    }
    return Math.abs(hash);
}

function dicebearUrl(seed) {
    // High-quality illustrated art with soft backgrounds
    const bg = 'b6e3f4,c0aede,d1d4f9';
    return `https://api.dicebear.com/7.x/${dicebearStyle}/svg?seed=${encodeURIComponent(seed)}&backgroundColor=${bg}&radius=50`;
}

async function loadAvatar(imgEl, seed) {
    try {
        const res = await fetch(dicebearUrl(seed), {
            mode: 'cors',
            headers: { 'Accept': 'image/svg+xml' }
        });
        if (!res.ok) throw new Error('HTTP ' + res.status);
        const blob = await res.blob();
        const url = URL.createObjectURL(blob);
        imgEl.src = url;
    } catch (e) {
        imgEl.src = localAvatarDataUrl(seed);
    }
}

// Local fallback: generate an avatar as a data URL with initials
function localAvatarDataUrl(name) {
    const canvas = document.createElement('canvas');
    const size = 64;
    canvas.width = size;
    canvas.height = size;
    const ctx = canvas.getContext('2d');

    // Background color derived from name hash
    const h = hashString(name);
    const hue = h % 360;
    const sat = 60 + (h % 20); // 60-79
    const light = 50; // fixed
    ctx.fillStyle = `hsl(${hue} ${sat}% ${light}%)`;
    ctx.beginPath();
    ctx.arc(size/2, size/2, size/2, 0, Math.PI*2);
    ctx.fill();

    // Initials
    const parts = name.split(/\s+/).filter(Boolean);
    const initials = parts.slice(0,2).map(p => p[0].toUpperCase()).join('');
    ctx.fillStyle = 'rgba(255,255,255,0.92)';
    ctx.font = 'bold 26px system-ui, -apple-system, Segoe UI, Roboto, Ubuntu, Cantarell, Noto Sans';
    ctx.textAlign = 'center';
    ctx.textBaseline = 'middle';
    ctx.fillText(initials, size/2, size/2);

    return canvas.toDataURL('image/png');
}

// expose generator for inline onerror usage in list rendering
window.localAvatarGen = localAvatarDataUrl;

function updateUserCard() {
    if (!userNameElement || !userAvatarElement) return;
    userNameElement.textContent = currentUsername;
    // Load avatar via CORS; fallback to local initials
    loadAvatar(userAvatarElement, currentUsername);
    userAvatarElement.alt = currentUsername;
}

function updateStatus(status) {
    statusElement.className = 'status ' + status;
    statusElement.textContent = status.charAt(0).toUpperCase() + status.slice(1);
}

function connectWebSocket() {
    updateStatus('connecting');

    if (!currentUsername) {
        currentUsername = generateUsername();
        updateUserCard();
    }

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${protocol}//${window.location.host}/ws/lobby?username=${encodeURIComponent(currentUsername)}`;

    ws = new WebSocket(wsUrl);

    ws.onopen = () => {
        console.log('Connected to lobby');
        updateStatus('connected');
        if (reconnectTimeout) {
            clearTimeout(reconnectTimeout);
            reconnectTimeout = null;
        }
    };

    ws.onmessage = (event) => {
        console.log('Message from server:', event.data);

        try {
            const data = JSON.parse(event.data);

            if (data.type === 'PLAYER_LIST') {
                updatePlayersList(data.players);
            }
        } catch (err) {
            console.error('Failed to parse message:', err);
        }
    };

    ws.onerror = (error) => {
        console.error('WebSocket error:', error);
    };

    ws.onclose = () => {
        console.log('Disconnected from lobby');
        updateStatus('disconnected');
        ws = null;

        // Attempt to reconnect after 3 seconds
        reconnectTimeout = setTimeout(() => {
            console.log('Attempting to reconnect...');
            connectWebSocket();
        }, 3000);
    };
}

function updatePlayersList(players) {
    if (players.length === 0) {
        playersListElement.innerHTML = '<div class="empty-message">No players in lobby</div>';
        return;
    }

    // Build list with async avatar loading
    playersListElement.innerHTML = '';
    players.forEach(player => {
        const item = document.createElement('div');
        item.className = 'player-item';

        const img = document.createElement('img');
        img.className = 'player-avatar';
        img.alt = player;
        // Start with local placeholder, then attempt external
        img.src = localAvatarDataUrl(player);
        loadAvatar(img, player);

        const name = document.createElement('div');
        name.className = 'player-name';
        name.textContent = player;

        item.appendChild(img);
        item.appendChild(name);
        playersListElement.appendChild(item);
    });
}

// Initialize connection when page loads
window.addEventListener('load', () => {
    currentUsername = generateUsername();
    updateUserCard();
    connectWebSocket();
});
