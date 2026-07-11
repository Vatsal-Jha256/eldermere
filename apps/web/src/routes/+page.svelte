<script lang="ts">
  import { onMount } from 'svelte';
  import FastNoiseLite from 'fastnoise-lite';
  import { AtmosphereAudio } from '$lib/audio';
  import { buildAtmosphereProfile, paletteFor, hashText, mulberry32, type AtmosphereProfile } from '$lib/atmosphere';
  import { env } from '$env/dynamic/public';

  const noise = new FastNoiseLite();
  noise.SetNoiseType(FastNoiseLite.NoiseType.OpenSimplex2);
  noise.SetFractalType(FastNoiseLite.FractalType.FBm);

  const canvasCache = {
    key: '',
    bg: typeof document !== 'undefined' ? document.createElement('canvas') : null,
    motifs: typeof document !== 'undefined' ? document.createElement('canvas') : null,
    overlay: typeof document !== 'undefined' ? document.createElement('canvas') : null
  };


  type RoomView = {
    id: string;
    name: string;
    description: string;
    exits: Record<string, string>;
    atmosphere: {
      palette?: string;
      weather?: string;
      myth_layer?: string;
      motifs?: string[];
    };
  };

  type ServerEvent = {
    type: string;
    text: string;
    room?: RoomView;
  };

  type PlayerSession = {
    player_id: string;
    display_name: string;
    token: string;
  };

  type VisualEvent = {
    kind: string;
    seed: number;
    startedAt: number;
  };

  const sessionKey = 'eldermere.session';
  const displayNameKey = 'eldermere.displayName';
  const commands = [
    'help',
    'exits',
    'who',
    'help social',
    'quest',
    'story',
    'story eligible',
    'story start sword-test',
    'story status',
    'story next',
    'travel arthurian-core',
    'map',
    'fight',
    'odds',
    'say hail from Camelot',
    'factions'
  ];

  let command = $state('');
  let displayName = $state('Wanderer');
  let connected = $state(false);
  let connecting = $state(false);
  let room = $state<RoomView | null>(null);
  let backgroundCanvas = $state<HTMLCanvasElement | null>(null);
  let logElement = $state<HTMLDivElement | null>(null);
  let visualEvents: VisualEvent[] = [];
  let commandHistory: string[] = [];
  let historyIndex = 0;
  let reconnectTimer: ReturnType<typeof setTimeout> | null = null;
  let log = $state([
    'Opening a path to the Eldermere server...'
  ]);
  let socket: WebSocket | null = null;
  const apiBase = env.PUBLIC_API_BASE?.trim() ?? '';
  const atmosphereStyle = $derived(buildAtmosphereStyle(room));

  let audio: AtmosphereAudio | null = null;

  onMount(() => {
    audio = new AtmosphereAudio();

    const unlockAudio = async () => {
      if (audio) {
        await audio.start();
        window.removeEventListener('click', unlockAudio);
        window.removeEventListener('keydown', unlockAudio);
      }
    };
    window.addEventListener('click', unlockAudio);
    window.addEventListener('keydown', unlockAudio);

    let active = true;
    connect().catch((error) => {
      if (active) {
        log = [...log, `Connection setup failed: ${error instanceof Error ? error.message : 'unknown error'}`];
      }
    });

    return () => {
      active = false;
      if (reconnectTimer) {
        clearTimeout(reconnectTimer);
      }
      socket?.close();
      window.removeEventListener('click', unlockAudio);
      window.removeEventListener('keydown', unlockAudio);
      if (audio) {
        audio.stopAll();
      }
    };
  });

  $effect(() => {
    if (audio && room) {
      audio.updateMood(room);
    }
  });

  $effect(() => {
    if (logElement) {
      logElement.scrollTop = logElement.scrollHeight;
    }
  });

  $effect(() => {
    const canvas = backgroundCanvas;
    const currentRoom = room;
    if (!canvas) return;

    let frame = 0;
    let time = 0;
    const prefersReducedMotion = window.matchMedia('(prefers-reduced-motion: reduce)');
    const loop = (t: number) => {
      time = t;
      drawAtmosphereCanvas(canvas, currentRoom, time);
      if (prefersReducedMotion.matches) return;
      frame = requestAnimationFrame(loop);
    };

    const draw = () => {
      if (!frame) frame = requestAnimationFrame(loop);
    };

    draw();
    const observer = new ResizeObserver(() => {
      const rect = canvas.getBoundingClientRect();
      const pixelRatio = window.devicePixelRatio || 1;
      canvas.width = Math.max(1, Math.floor(rect.width * pixelRatio));
      canvas.height = Math.max(1, Math.floor(rect.height * pixelRatio));
    });
    observer.observe(canvas);

    return () => {
      cancelAnimationFrame(frame);
      frame = 0;
      observer.disconnect();
    };
  });

  async function connect(options: { freshSession?: boolean; retryStaleSession?: boolean } = {}) {
    if (connecting) return;
    connecting = true;
    connected = false;

    let session: PlayerSession;
    try {
      session = await getPlayerSession(options.freshSession);
    } catch (error) {
      connecting = false;
      throw error;
    }
    const nextSocket = new WebSocket(toWebSocketURL('/ws', session));
    let opened = false;
    socket = nextSocket;

    nextSocket.addEventListener('open', () => {
      if (socket !== nextSocket) return;
      opened = true;
      connecting = false;
      connected = true;
    });

    nextSocket.addEventListener('message', (event) => {
      if (socket !== nextSocket) return;
      const parsed = parseServerEvent(event.data);
      if (parsed.room) {
        room = parsed.room;
      }
      audio?.playCue(parsed.type, parsed.text);
      pushVisualEvent(parsed.type, parsed.text);
      log = [...log, parsed.text];
    });

    nextSocket.addEventListener('close', () => {
      if (socket !== nextSocket) return;
      connected = false;
      connecting = false;
      socket = null;

      if (!opened && options.retryStaleSession !== false && localStorage.getItem(sessionKey)) {
        localStorage.removeItem(sessionKey);
        log = [...log, 'Stored session was rejected. Creating a fresh player session...'];
        reconnectTimer = setTimeout(() => {
          reconnectTimer = null;
          connect({ freshSession: true, retryStaleSession: false }).catch((error) => {
            connecting = false;
            log = [...log, `Reconnect failed: ${error instanceof Error ? error.message : 'unknown error'}`];
          });
        }, 200);
        return;
      }

      log = [...log, 'Disconnected from the server.'];
    });

    nextSocket.addEventListener('error', () => {
      if (socket !== nextSocket) return;
      log = [...log, 'Connection error. Check the API URL and server availability.'];
    });
  }

  function submitCommand() {
    const trimmed = command.trim();
    if (!trimmed) return;
    if (!socket || socket.readyState !== WebSocket.OPEN) {
      log = [...log, 'Cannot send yet. Reconnect and wait for Connected.'];
      return;
    }

    log = [...log, `> ${trimmed}`];
    commandHistory = [...commandHistory, trimmed].slice(-40);
    historyIndex = commandHistory.length;
    socket?.send(JSON.stringify({ command: trimmed }));
    command = '';
  }

  function runCommand(value: string) {
    command = value;
    requestAnimationFrame(() => submitCommand());
  }

  function reconnect(freshSession = false) {
    if (reconnectTimer) {
      clearTimeout(reconnectTimer);
      reconnectTimer = null;
    }
    if (freshSession) {
      localStorage.removeItem(sessionKey);
    }
    if (socket && socket.readyState < WebSocket.CLOSING) {
      socket.close();
    }
    socket = null;
    connected = false;
    connecting = false;
    log = [...log, freshSession ? 'Starting a fresh player session...' : 'Reconnecting...'];
    connect({ freshSession, retryStaleSession: true }).catch((error) => {
      connecting = false;
      log = [...log, `Reconnect failed: ${error instanceof Error ? error.message : 'unknown error'}`];
    });
  }

  function handleCommandKeydown(event: KeyboardEvent) {
    if (event.key === 'ArrowUp') {
      if (commandHistory.length === 0) return;
      event.preventDefault();
      historyIndex = Math.max(0, historyIndex - 1);
      command = commandHistory[historyIndex] ?? command;
    } else if (event.key === 'ArrowDown') {
      if (commandHistory.length === 0) return;
      event.preventDefault();
      historyIndex = Math.min(commandHistory.length, historyIndex + 1);
      command = commandHistory[historyIndex] ?? '';
    }
  }

  function parseServerEvent(data: string): ServerEvent {
    try {
      return JSON.parse(data) as ServerEvent;
    } catch {
      return { type: 'system', text: data };
    }
  }

  function pushVisualEvent(kind: string, text: string) {
    const visualKinds = new Set(['fight', 'recruit', 'quest', 'story', 'move', 'party', 'error']);
    if (!visualKinds.has(kind)) return;

    const event = {
      kind,
      seed: hashText([room?.id ?? 'eldermere', kind, text.slice(0, 80)].join('|')),
      startedAt: performance.now()
    };
    visualEvents = [...visualEvents.filter((item) => performance.now() - item.startedAt < 2200), event].slice(-8);
  }

  async function getPlayerSession(freshSession = false) {
    if (freshSession) {
      localStorage.removeItem(sessionKey);
    }
    const existing = localStorage.getItem(sessionKey);
    if (existing) {
      try {
        const session = JSON.parse(existing) as PlayerSession;
        if (isPlayerSession(session)) {
          displayName = session.display_name || 'Wanderer';
          return session;
        }
      } catch {
        // Fall through to create a clean session.
      }
      localStorage.removeItem(sessionKey);
    }

    const savedName = localStorage.getItem(displayNameKey);
    const name = normalizeDisplayName(savedName ?? displayName);
    displayName = name;
    const response = await fetch(toApiURL('/api/v1/sessions'), {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ display_name: name })
    });

    if (!response.ok) {
      throw new Error(`session request failed with ${response.status}`);
    }

    const session = (await response.json()) as PlayerSession;
    localStorage.setItem(sessionKey, JSON.stringify(session));
    return session;
  }

  function isPlayerSession(value: unknown): value is PlayerSession {
    if (!value || typeof value !== 'object') return false;
    const session = value as PlayerSession;
    return Boolean(session.player_id && session.token && session.display_name);
  }

  function saveDisplayName() {
    const name = normalizeDisplayName(displayName);
    displayName = name;
    localStorage.setItem(displayNameKey, name);
    log = [...log, `Player name set to ${name}.`];
    reconnect(true);
  }

  function normalizeDisplayName(value: string) {
    const trimmed = value.trim().replace(/\s+/g, ' ');
    return trimmed.length > 0 ? trimmed.slice(0, 28) : 'Wanderer';
  }

  function toApiURL(path: string) {
    const base = apiBase || window.location.origin;
    return new URL(path, base);
  }

  function toWebSocketURL(path: string, session: PlayerSession) {
    const base = apiBase || window.location.origin;
    const url = new URL(path, base);
    url.protocol = url.protocol === 'https:' ? 'wss:' : 'ws:';
    url.searchParams.set('player_id', session.player_id);
    url.searchParams.set('token', session.token);
    return url.toString();
  }

  function buildAtmosphereStyle(current: RoomView | null) {
    const palette = paletteFor(current?.atmosphere?.palette);
    const motifSeed = hashText(current?.atmosphere?.motifs?.join('|') ?? current?.id ?? 'eldermere');
    const mistAngle = 25 + (motifSeed % 80);
    const glowX = 18 + (motifSeed % 64);
    const glowY = 12 + ((motifSeed >> 3) % 48);

    return [
      `--bg-a: ${palette[0]}`,
      `--bg-b: ${palette[1]}`,
      `--bg-c: ${palette[2]}`,
      `--mist-angle: ${mistAngle}deg`,
      `--glow-x: ${glowX}%`,
      `--glow-y: ${glowY}%`
    ].join(';');
  }

  function drawAtmosphereCanvas(canvas: HTMLCanvasElement, current: RoomView | null, time: number = 0) {
    const rect = canvas.getBoundingClientRect();
    const pixelRatio = window.devicePixelRatio || 1;
    const width = Math.max(1, Math.floor(rect.width * pixelRatio));
    const height = Math.max(1, Math.floor(rect.height * pixelRatio));
    if (canvas.width !== width || canvas.height !== height) {
      canvas.width = width;
      canvas.height = height;
    }

    const context = canvas.getContext('2d');
    if (!context) return;

    const profile = buildAtmosphereProfile(current);
    const w = rect.width;
    const h = rect.height;
    const paletteStr = profile.palette;
    const palette = paletteFor(paletteStr);
    const colors = palette.map(hexToRgb);
    const motifsList = profile.motifs;
    const weatherStr = profile.weather;

    const cacheKey = `${profile.seed}-${width}-${height}`;

    if (canvasCache.key !== cacheKey && canvasCache.bg && canvasCache.motifs && canvasCache.overlay) {
      canvasCache.key = cacheKey;

      const setupOffscreen = (c: HTMLCanvasElement) => {
        c.width = width;
        c.height = height;
        const ctx = c.getContext('2d');
        ctx?.setTransform(pixelRatio, 0, 0, pixelRatio, 0, 0);
        return ctx;
      };

      const bgCtx = setupOffscreen(canvasCache.bg);
      if (bgCtx) {
        const sky = bgCtx.createLinearGradient(0, 0, w, h);
        sky.addColorStop(0, rgb(colors[0]));
        sky.addColorStop(0.62, rgb(mix(colors[0], colors[1], 0.62)));
        sky.addColorStop(1, rgb(colors[1]));
        bgCtx.fillStyle = sky;
        bgCtx.fillRect(0, 0, w, h);

        const randomBg = mulberry32(profile.seed);
        drawGlow(bgCtx, w, h, colors[2], 0.18 + randomBg() * 0.14, 0.16 + randomBg() * 0.18, Math.min(w, h) * 0.5);
        drawGlow(bgCtx, w, h, colors[1], 0.72 + randomBg() * 0.16, 0.2 + randomBg() * 0.18, Math.min(w, h) * 0.38);
        drawTerrainField(bgCtx, profile, w, h, colors);
        drawCellularLayer(bgCtx, profile, w, h, colors);
      }

      const motifsCtx = setupOffscreen(canvasCache.motifs);
      if (motifsCtx) {
        const randomMotifs = mulberry32(profile.seed);
        drawStructuredPattern(motifsCtx, profile, w, h, colors, randomMotifs);
        drawMotifs(motifsCtx, w, h, colors, motifsList, randomMotifs);
      }

      const overlayCtx = setupOffscreen(canvasCache.overlay);
      if (overlayCtx) {
        const randomOverlay = mulberry32(profile.seed);
        drawGrain(overlayCtx, w, h, colors[2], randomOverlay);

        const vignette = overlayCtx.createRadialGradient(w * 0.45, h * 0.42, Math.min(w, h) * 0.1, w * 0.5, h * 0.5, Math.max(w, h) * 0.75);
        vignette.addColorStop(0, 'rgba(0, 0, 0, 0)');
        vignette.addColorStop(1, 'rgba(0, 0, 0, 0.72)');
        overlayCtx.fillStyle = vignette;
        overlayCtx.fillRect(0, 0, w, h);
      }
    }

    context.setTransform(1, 0, 0, 1, 0, 0);
    if (canvasCache.bg) context.drawImage(canvasCache.bg, 0, 0);

    context.setTransform(pixelRatio, 0, 0, pixelRatio, 0, 0);
    noise.SetSeed(profile.seed);
    const dynRandom = mulberry32(profile.seed);

    drawProceduralMood(context, w, h, paletteStr ?? '', colors, time, dynRandom);
    drawHorizon(context, w, h, colors, time);

    context.setTransform(1, 0, 0, 1, 0, 0);
    if (canvasCache.motifs) context.drawImage(canvasCache.motifs, 0, 0);

    context.setTransform(pixelRatio, 0, 0, pixelRatio, 0, 0);
    drawWeather(context, profile, w, h, weatherStr, colors, dynRandom, time);
    drawVisualEvents(context, w, h, colors, time);

    context.setTransform(1, 0, 0, 1, 0, 0);
    if (canvasCache.overlay) context.drawImage(canvasCache.overlay, 0, 0);
  }

  function drawTerrainField(context: CanvasRenderingContext2D, profile: AtmosphereProfile, w: number, h: number, colors: RGB[]) {
    const cell = Math.max(18, profile.visual.tileSize);
    const cols = Math.ceil(w / cell);
    const rows = Math.ceil(h / cell);
    const terrainNoise = new FastNoiseLite(profile.seed ^ 0x9e3779b9);
    terrainNoise.SetNoiseType(FastNoiseLite.NoiseType.OpenSimplex2);
    terrainNoise.SetFractalType(FastNoiseLite.FractalType.FBm);
    terrainNoise.SetFractalOctaves(4);

    context.save();
    for (let y = 0; y < rows; y += 1) {
      for (let x = 0; x < cols; x += 1) {
        const elevation = (terrainNoise.GetNoise(x * profile.visual.terrainScale, y * profile.visual.terrainScale) + 1) / 2;
        const moisture = (terrainNoise.GetNoise((x + 80) * 0.35, (y - 30) * 0.35) + 1) / 2;
        const color = classifyTerrainColor(profile, colors, elevation, moisture);
        context.fillStyle = rgba(color, 0.1 + elevation * 0.16);
        context.fillRect(x * cell, y * cell, cell + 1, cell + 1);
      }
    }
    context.restore();
  }

  function drawCellularLayer(context: CanvasRenderingContext2D, profile: AtmosphereProfile, w: number, h: number, colors: RGB[]) {
    if (!['cave', 'void', 'water', 'forest'].includes(profile.biome)) return;

    const cell = Math.max(16, Math.floor(profile.visual.tileSize * 0.85));
    const cols = Math.ceil(w / cell) + 2;
    const rows = Math.ceil(h / cell) + 2;
    let grid = seedCellularGrid(cols, rows, profile.seed, profile.visual.caveFill);

    for (let i = 0; i < profile.visual.caveIterations; i += 1) {
      grid = stepCellularGrid(grid, cols, rows);
    }

    context.save();
    context.fillStyle = rgba(profile.biome === 'water' ? colors[1] : colors[0], profile.biome === 'void' ? 0.5 : 0.34);
    for (let y = 1; y < rows - 1; y += 1) {
      for (let x = 1; x < cols - 1; x += 1) {
        if (!grid[y * cols + x]) continue;
        const px = (x - 1) * cell;
        const py = (y - 1) * cell;
        context.beginPath();
        context.roundRect(px, py, cell * 1.08, cell * 1.08, cell * 0.28);
        context.fill();
      }
    }
    context.restore();
  }

  function drawStructuredPattern(context: CanvasRenderingContext2D, profile: AtmosphereProfile, w: number, h: number, colors: RGB[], random: () => number) {
    if (!['court', 'cave', 'void', 'fey'].includes(profile.biome)) return;

    const cell = profile.visual.tileSize;
    const cols = Math.ceil(w / cell);
    const rows = Math.ceil(h / cell);
    const tiles = collapseStructureTiles(cols, rows, profile, random);

    context.save();
    context.lineWidth = 1.5;
    for (let y = 0; y < rows; y += 1) {
      for (let x = 0; x < cols; x += 1) {
        const tile = tiles[y * cols + x];
        if (tile === 0) continue;
        const px = x * cell;
        const py = y * cell;
        context.strokeStyle = rgba(colors[2], tile === 2 ? 0.24 : 0.14);
        context.fillStyle = rgba(tile === 3 ? colors[2] : colors[0], tile === 3 ? 0.08 : 0.18);
        if (tile === 1) {
          context.fillRect(px, py + cell * 0.42, cell, cell * 0.16);
        } else if (tile === 2) {
          context.strokeRect(px + cell * 0.18, py + cell * 0.18, cell * 0.64, cell * 0.64);
        } else {
          context.beginPath();
          context.arc(px + cell * 0.5, py + cell * 0.5, cell * 0.18, 0, Math.PI * 2);
          context.fill();
        }
      }
    }
    context.restore();
  }

  function drawProceduralMood(context: CanvasRenderingContext2D, w: number, h: number, palette: string, colors: RGB[], time: number, random: () => number) {
    if (palette === 'fey-realm' || palette === 'eldritch-void') {
      context.fillStyle = rgba(colors[2], 0.5);
      for (let i = 0; i < 50; i++) {
        let x = random() * w;
        let y = random() * h;
        const particleTime = time * 0.05 + random() * 1000;
        x += noise.GetNoise(x * 0.1, particleTime) * 30;
        y += noise.GetNoise(y * 0.1, particleTime + 100) * 30;
        context.beginPath();
        context.arc(x, y, 1 + random() * 2, 0, Math.PI * 2);
        context.fill();
      }
    } else if (palette === 'inferno') {
      context.fillStyle = rgba(colors[2], 0.7);
      for (let i = 0; i < 80; i++) {
        let x = random() * w;
        let y = random() * h;
        const particleTime = time * 0.08 + random() * 1000;
        x += noise.GetNoise(x * 0.2, particleTime) * 20;
        y -= (particleTime * 0.5) % h;
        if (y < 0) y += h;
        context.beginPath();
        context.arc(x, y, 1 + random() * 3, 0, Math.PI * 2);
        context.fill();
      }
    }
  }

  function drawVisualEvents(context: CanvasRenderingContext2D, w: number, h: number, colors: RGB[], time: number) {
    const now = performance.now();
    visualEvents = visualEvents.filter((event) => now - event.startedAt < 1800);
    context.save();

    for (const event of visualEvents) {
      const age = Math.max(0, now - event.startedAt);
      const progress = age / 1800;
      const random = mulberry32(event.seed);
      const cx = w * (0.22 + random() * 0.56);
      const cy = h * (0.2 + random() * 0.5);
      const pulse = Math.sin(progress * Math.PI);
      const radius = Math.min(w, h) * (0.08 + progress * 0.34);
      const accent = event.kind === 'error' ? mix(colors[2], [210, 60, 50], 0.55) : colors[2];

      context.globalAlpha = pulse * 0.32;
      context.strokeStyle = rgba(accent, 0.6);
      context.lineWidth = event.kind === 'fight' ? 3 : 1.5;
      context.beginPath();
      context.arc(cx, cy, radius, 0, Math.PI * 2);
      context.stroke();

      if (event.kind === 'fight' || event.kind === 'recruit') {
        context.fillStyle = rgba(accent, 0.22);
        for (let i = 0; i < 10; i += 1) {
          const angle = random() * Math.PI * 2 + time * 0.001;
          const distance = radius * (0.35 + random() * 0.8);
          context.beginPath();
          context.arc(cx + Math.cos(angle) * distance, cy + Math.sin(angle) * distance, 1.5 + random() * 3, 0, Math.PI * 2);
          context.fill();
        }
      }
    }

    context.restore();
  }

  function drawGlow(context: CanvasRenderingContext2D, w: number, h: number, color: RGB, x: number, y: number, radius: number) {
    const glow = context.createRadialGradient(w * x, h * y, 0, w * x, h * y, radius);
    glow.addColorStop(0, rgba(color, 0.48));
    glow.addColorStop(0.45, rgba(color, 0.14));
    glow.addColorStop(1, rgba(color, 0));
    context.fillStyle = glow;
    context.fillRect(0, 0, w, h);
  }

  function drawHorizon(context: CanvasRenderingContext2D, w: number, h: number, colors: RGB[], time: number) {
    context.save();
    context.fillStyle = rgba(mix(colors[0], colors[1], 0.42), 0.72);
    context.beginPath();
    context.moveTo(0, h);
    const timeOffset = time * 0.05;

    for (let x = 0; x <= w; x += 10) {
      // Use FastNoiseLite for rolling hills
      const yOffset = noise.GetNoise(x * 0.3, 0) * (h * 0.1);
      const y = h * 0.65 + yOffset;
      context.lineTo(x, y);
    }
    context.lineTo(w, h);
    context.closePath();
    context.fill();

    context.globalAlpha = 0.28;
    for (let i = 0; i < 8; i += 1) {
      const startY = h * (0.48 + i * 0.055);
      context.strokeStyle = rgba(colors[2], 0.2 - i * 0.015);
      context.lineWidth = 1.5;
      context.beginPath();
      for (let x = 0; x <= w; x += 20) {
        const wave = noise.GetNoise(x * 0.4 + timeOffset, i * 100) * 20;
        if (x === 0) context.moveTo(x, startY + wave);
        else context.lineTo(x, startY + wave);
      }
      context.stroke();
    }
    context.restore();
  }

  function drawMotifs(context: CanvasRenderingContext2D, w: number, h: number, colors: RGB[], motifs: string[], random: () => number) {
    const text = motifs.join(' ').toLowerCase();
    context.save();
    context.fillStyle = rgba(colors[0], 0.62);
    context.strokeStyle = rgba(colors[2], 0.34);
    context.lineWidth = 2;

    if (text.includes('banner') || text.includes('claim') || text.includes('table')) {
      for (let i = 0; i < 4; i += 1) drawBanner(context, w * (0.14 + i * 0.13), h * (0.38 + random() * 0.18), h * 0.26, colors[2]);
    }
    if (text.includes('candle') || text.includes('wax')) {
      for (let i = 0; i < 10; i += 1) drawCandle(context, w, h, w * (0.12 + random() * 0.72), h * (0.58 + random() * 0.24), 18 + random() * 42, colors[2]);
    }
    if (text.includes('bridge') || text.includes('jetty') || text.includes('ferry')) {
      drawBridge(context, w, h, colors[0], colors[2]);
    }
    if (text.includes('boat') || text.includes('shore') || text.includes('avalon')) {
      for (let i = 0; i < 3; i += 1) drawBoat(context, w * (0.28 + random() * 0.52), h * (0.58 + random() * 0.18), w * (0.08 + random() * 0.05), colors[0], colors[2]);
    }
    if (text.includes('stone') || text.includes('vault') || text.includes('mount')) {
      drawStone(context, w * 0.66, h * 0.58, Math.min(w, h) * 0.16, colors[0], colors[2]);
    }

    context.restore();
  }

  function drawBanner(context: CanvasRenderingContext2D, x: number, y: number, height: number, accent: RGB) {
    context.strokeStyle = rgba(accent, 0.35);
    context.beginPath();
    context.moveTo(x, y - height * 0.45);
    context.lineTo(x, y + height * 0.48);
    context.stroke();
    context.fillStyle = rgba(accent, 0.16);
    context.beginPath();
    context.moveTo(x, y - height * 0.42);
    context.lineTo(x + height * 0.28, y - height * 0.35);
    context.lineTo(x + height * 0.2, y - height * 0.13);
    context.lineTo(x, y - height * 0.18);
    context.closePath();
    context.fill();
  }

  function drawCandle(context: CanvasRenderingContext2D, w: number, h: number, x: number, y: number, height: number, accent: RGB) {
    context.fillStyle = 'rgba(15, 10, 6, 0.62)';
    context.fillRect(x - 3, y - height, 6, height);
    drawGlow(context, w, h, accent, x / w, (y - height) / h, height * 2.8);
  }

  function drawBridge(context: CanvasRenderingContext2D, w: number, h: number, dark: RGB, accent: RGB) {
    context.strokeStyle = rgba(accent, 0.24);
    context.lineWidth = Math.max(3, w * 0.006);
    context.beginPath();
    context.moveTo(w * 0.06, h * 0.64);
    context.quadraticCurveTo(w * 0.5, h * 0.5, w * 0.94, h * 0.66);
    context.stroke();
    context.fillStyle = rgba(dark, 0.54);
    context.fillRect(0, h * 0.67, w, h * 0.06);
  }

  function drawBoat(context: CanvasRenderingContext2D, x: number, y: number, width: number, dark: RGB, accent: RGB) {
    context.fillStyle = rgba(dark, 0.62);
    context.strokeStyle = rgba(accent, 0.3);
    context.beginPath();
    context.moveTo(x - width * 0.5, y);
    context.quadraticCurveTo(x, y + width * 0.28, x + width * 0.5, y);
    context.quadraticCurveTo(x, y + width * 0.12, x - width * 0.5, y);
    context.fill();
    context.stroke();
  }

  function drawStone(context: CanvasRenderingContext2D, x: number, y: number, radius: number, dark: RGB, accent: RGB) {
    context.fillStyle = rgba(dark, 0.5);
    context.strokeStyle = rgba(accent, 0.26);
    context.lineWidth = 2;
    context.beginPath();
    context.ellipse(x, y, radius * 0.7, radius, -0.08, 0, Math.PI * 2);
    context.fill();
    context.stroke();
  }

  function drawWeather(context: CanvasRenderingContext2D, profile: AtmosphereProfile, w: number, h: number, weather: string, colors: RGB[], random: () => number, time: number) {
    const lower = weather.toLowerCase();
    context.save();
    const timeOffset = time * 0.05;
    const density = profile.visual.particleDensity;
    if (lower.includes('rain') || lower.includes('drizzle')) {
      context.strokeStyle = rgba(colors[2], 0.25);
      context.lineWidth = 1;
      const speed = 0.8;
      const drops = Math.round(80 + density * 160);
      for (let i = 0; i < drops; i += 1) {
        let rx = random();
        let ry = random();
        let x = rx * w;
        let y = ry * h;
        y = (y + time * speed * (0.5 + random())) % h;
        // Use FastNoiseLite to skew rain slightly as if blown by wind
        const wind = noise.GetNoise(x * 0.5, y * 0.5) * 10;
        context.beginPath();
        context.moveTo(x, y);
        context.lineTo(x - 12 + wind, y + 40);
        context.stroke();
      }
    } else {
      // Draw organic mist/flow lines using FastNoiseLite and domain warping concept
      context.strokeStyle = rgba(colors[2], 0.08);
      context.lineWidth = 1.5;
      const strands = Math.round(30 + density * 70);
      for (let i = 0; i < strands; i += 1) {
        let baseRx = random();
        let baseRy = random();
        let driftRx = random();
        
        // Add time-based drift to starting coordinates so they flow across the screen
        let x = (baseRx * w + time * 0.02 * (0.5 + driftRx)) % w;
        let y = (baseRy * h + time * 0.01 * (0.5 + driftRx)) % h;
        
        context.beginPath();
        context.moveTo(x, y);
        for(let step = 0; step < 20; step++) {
          const angle = noise.GetNoise(x * 0.3 + timeOffset * 0.5, y * 0.3) * Math.PI * 2;
          x += Math.cos(angle) * 15;
          y += Math.sin(angle) * 15;
          context.lineTo(x, y);
        }
        context.stroke();
      }
    }
    context.restore();
  }

  function classifyTerrainColor(profile: AtmosphereProfile, colors: RGB[], elevation: number, moisture: number): RGB {
    if (profile.biome === 'water') return mix(colors[0], colors[1], Math.min(0.9, 0.35 + moisture * 0.55));
    if (profile.biome === 'fire') return mix(colors[1], colors[2], Math.min(0.85, elevation * 0.75));
    if (profile.biome === 'fey') return mix(colors[1], colors[2], 0.25 + moisture * 0.45);
    if (profile.biome === 'void') return mix(colors[0], colors[1], elevation * 0.25);
    if (profile.biome === 'court') return mix(colors[0], colors[2], elevation > 0.56 ? 0.28 : 0.12);
    return mix(colors[0], colors[1], 0.25 + elevation * 0.45);
  }

  function seedCellularGrid(cols: number, rows: number, seed: number, fill: number) {
    const random = mulberry32(seed ^ 0x85ebca6b);
    const grid = new Array<boolean>(cols * rows);
    for (let y = 0; y < rows; y += 1) {
      for (let x = 0; x < cols; x += 1) {
        const edge = x === 0 || y === 0 || x === cols - 1 || y === rows - 1;
        grid[y * cols + x] = edge || random() < fill;
      }
    }
    return grid;
  }

  function stepCellularGrid(grid: boolean[], cols: number, rows: number) {
    const next = new Array<boolean>(grid.length);
    for (let y = 0; y < rows; y += 1) {
      for (let x = 0; x < cols; x += 1) {
        const walls = countNeighbors(grid, cols, rows, x, y);
        next[y * cols + x] = walls >= 5;
      }
    }
    return next;
  }

  function countNeighbors(grid: boolean[], cols: number, rows: number, x: number, y: number) {
    let count = 0;
    for (let oy = -1; oy <= 1; oy += 1) {
      for (let ox = -1; ox <= 1; ox += 1) {
        if (ox === 0 && oy === 0) continue;
        const nx = x + ox;
        const ny = y + oy;
        if (nx < 0 || ny < 0 || nx >= cols || ny >= rows || grid[ny * cols + nx]) count += 1;
      }
    }
    return count;
  }

  function collapseStructureTiles(cols: number, rows: number, profile: AtmosphereProfile, random: () => number) {
    const tiles = new Array<number>(cols * rows).fill(0);
    const density = profile.visual.structureDensity;
    for (let y = 0; y < rows; y += 1) {
      for (let x = 0; x < cols; x += 1) {
        const left = x > 0 ? tiles[y * cols + x - 1] : 0;
        const up = y > 0 ? tiles[(y - 1) * cols + x] : 0;
        const coherent = left || up;
        const chance = coherent ? density * 1.35 : density;
        if (random() > chance) continue;

        if (profile.biome === 'court') {
          tiles[y * cols + x] = left === 1 || up === 1 || random() > 0.45 ? 1 : 2;
        } else if (profile.biome === 'fey') {
          tiles[y * cols + x] = random() > 0.48 ? 3 : 2;
        } else {
          tiles[y * cols + x] = coherent && random() > 0.35 ? coherent : random() > 0.58 ? 2 : 3;
        }
      }
    }
    return tiles;
  }

  function drawGrain(context: CanvasRenderingContext2D, w: number, h: number, accent: RGB, random: () => number) {
    const count = Math.min(900, Math.floor((w * h) / 1800));
    context.fillStyle = rgba(accent, 0.08);
    for (let i = 0; i < count; i += 1) {
      context.fillRect(random() * w, random() * h, 1, 1);
    }
  }

  type RGB = [number, number, number];

  function hexToRgb(hex: string): RGB {
    const normalized = hex.replace('#', '');
    return [
      Number.parseInt(normalized.slice(0, 2), 16),
      Number.parseInt(normalized.slice(2, 4), 16),
      Number.parseInt(normalized.slice(4, 6), 16)
    ];
  }

  function mix(a: RGB, b: RGB, amount: number): RGB {
    return [
      Math.round(a[0] + (b[0] - a[0]) * amount),
      Math.round(a[1] + (b[1] - a[1]) * amount),
      Math.round(a[2] + (b[2] - a[2]) * amount)
    ];
  }

  function rgb(color: RGB) {
    return `rgb(${color[0]}, ${color[1]}, ${color[2]})`;
  }

  function rgba(color: RGB, alpha: number) {
    return `rgba(${color[0]}, ${color[1]}, ${color[2]}, ${alpha})`;
  }

</script>

<svelte:head>
  <title>Eldermere</title>
  <meta
    name="description"
    content="A browser MUD creature-RPG for Arthurian legend and connected myths."
  />
</svelte:head>

<main class="shell">
  <section class="room" aria-label="Current room" style={atmosphereStyle}>
    <canvas class="room__canvas" bind:this={backgroundCanvas} aria-hidden="true"></canvas>
    <div class="room__background"></div>
    <div class="room__content">
      <p class="eyebrow">
        {room?.name ?? 'Connecting'} / {room?.atmosphere?.myth_layer ?? 'Camelot Underbelly'}
      </p>
      <h1>Eldermere</h1>
      <p class="lede">
        {room?.description ??
          "A browser MUD for connected legends. Start in Arthur's Britain, recruit strange allies, and follow rumors that should not know each other yet."}
      </p>
      {#if room?.atmosphere?.weather || room?.atmosphere?.motifs?.length}
        <p class="atmosphere">
          {room.atmosphere.weather}
          {#if room.atmosphere.weather && room.atmosphere.motifs?.length} / {/if}
          {room.atmosphere.motifs?.join(', ')}
        </p>
      {/if}
    </div>
  </section>

  <section class="console crt-terminal" aria-label="Command console">
    <div class="status" class:online={connected}>
      <span>{connected ? 'Connected' : connecting ? 'Connecting' : 'Disconnected'}</span>
      <span>{displayName}</span>
      <span>{room?.name ?? 'No room yet'}</span>
      <button type="button" onclick={() => reconnect()} disabled={connecting}>Reconnect</button>
    </div>

    <div class="console__guide" aria-label="Play guidance">
      <form class="player" onsubmit={(event) => { event.preventDefault(); saveDisplayName(); }}>
        <label for="display-name">Player</label>
        <input
          id="display-name"
          bind:value={displayName}
          autocomplete="nickname"
          maxlength="28"
        />
        <button type="submit">Set</button>
      </form>

      <div class="guide-grid">
        <section class="guide-panel" aria-labelledby="mud-heading">
          <h2 id="mud-heading">Together</h2>
          <p>Open this page in another browser or device on the same host, set a different player name, then use room speech.</p>
          <div class="mini-actions">
            <button type="button" onclick={() => runCommand('who')} disabled={!connected}>Who</button>
            <button type="button" onclick={() => (command = 'say ')} disabled={!connected}>Say</button>
          </div>
        </section>

        <section class="guide-panel" aria-labelledby="story-heading">
          <h2 id="story-heading">Arthurian Story</h2>
          <p>Start with the sword-test arc, then use story status and next to follow source-grounded Camelot branches.</p>
          <div class="mini-actions">
            <button type="button" onclick={() => runCommand('story eligible')} disabled={!connected}>Eligible</button>
            <button type="button" onclick={() => runCommand('story status')} disabled={!connected}>Status</button>
          </div>
        </section>
      </div>
    </div>

    {#if room}
      <div class="exits" aria-label="Visible exits">
        <span>Exits</span>
        {#each Object.keys(room.exits) as exit}
          <button type="button" onclick={() => runCommand(`go ${exit}`)} disabled={!connected}>{exit}</button>
        {/each}
      </div>
    {/if}

    <div class="console__log" aria-live="polite" aria-relevant="additions" bind:this={logElement}>
      {#each log as line}
        <p>{line}</p>
      {/each}
    </div>

    <form class="command" onsubmit={(event) => { event.preventDefault(); submitCommand(); }}>
      <label for="command">Command</label>
      <input
        id="command"
        bind:value={command}
        autocomplete="off"
        spellcheck="false"
        placeholder="try: look"
        disabled={!connected}
        onkeydown={handleCommandKeydown}
      />
      <button type="submit" disabled={!connected}>Send</button>
    </form>

    <div class="chips" aria-label="Example commands">
      {#each commands as item}
        <button type="button" onclick={() => (command = item)} disabled={!connected}>{item}</button>
      {/each}
    </div>
  </section>
</main>
