export type RoomAtmosphere = {
  id?: string;
  atmosphere?: {
    palette?: string;
    weather?: string;
    myth_layer?: string;
    motifs?: string[];
  };
} | null;

export type SoundMode = 'rain' | 'wind' | 'fire' | 'water' | 'sacred' | 'void' | 'court' | 'field';

export type AtmosphereProfile = {
  key: string;
  seed: number;
  palette: string;
  weather: string;
  mythLayer: string;
  motifs: string[];
  modes: SoundMode[];
};

export function buildAtmosphereProfile(room: RoomAtmosphere): AtmosphereProfile {
  const palette = room?.atmosphere?.palette ?? '';
  const weather = room?.atmosphere?.weather ?? '';
  const mythLayer = room?.atmosphere?.myth_layer ?? '';
  const motifs = room?.atmosphere?.motifs ?? [];
  const key = [room?.id ?? 'eldermere', palette, weather, mythLayer, motifs.join('|')].join('|');
  return {
    key,
    seed: hashText(key),
    palette,
    weather,
    mythLayer,
    motifs,
    modes: detectModes(palette, weather, mythLayer, motifs)
  };
}

export function paletteFor(name?: string) {
  const palettes: Record<string, [string, string, string]> = {
    'rain-gold': ['#0e1612', '#7f6a32', '#d9b45f'],
    blackwater: ['#090d10', '#13232a', '#6b7f87'],
    'candle-smoke': ['#17110d', '#6f3e24', '#d9a45d'],
    'tavern-red': ['#190c0c', '#612018', '#d67b45'],
    'avalon-green': ['#081411', '#174736', '#96c7a1'],
    'relic-vault': ['#100f15', '#594c7a', '#c4a45d'],
    'coin-shadow': ['#11100b', '#5c4a1f', '#c59a3a'],
    'oracle-blue': ['#07131f', '#164466', '#8fc6d9'],
    'bronze-ash': ['#15100d', '#694421', '#c28f52'],
    'peaceful-village': ['#0f1710', '#9c5a17', '#426e2e'],
    'haunted-forest': ['#0a0b12', '#1b2f1f', '#3c1d4a'],
    'fey-realm': ['#0e171c', '#2c5259', '#7a295c'],
    'underdark': ['#02040a', '#08142b', '#123069'],
    'inferno': ['#140202', '#541010', '#ad3215'],
    'frozen-wastes': ['#121b21', '#42617a', '#a6cfeb'],
    'eldritch-void': ['#05010a', '#18072e', '#361066'],
    'ancient-ruins': ['#14120e', '#3b3323', '#2d3824'],
    storm: ['#0a0c0f', '#242b36', '#4a5669'],
    'divine-realm': ['#171407', '#5e4e16', '#cfb867']
  };
  return palettes[name ?? ''] ?? ['#101511', '#314439', '#e2b65f'];
}

export function hashText(value: string) {
  let hash = 0;
  for (let index = 0; index < value.length; index += 1) {
    hash = (hash * 31 + value.charCodeAt(index)) >>> 0;
  }
  return hash;
}

export function mulberry32(seed: number) {
  return function random() {
    let value = (seed += 0x6d2b79f5);
    value = Math.imul(value ^ (value >>> 15), value | 1);
    value ^= value + Math.imul(value ^ (value >>> 7), value | 61);
    return ((value ^ (value >>> 14)) >>> 0) / 4294967296;
  };
}

function detectModes(palette: string, weather: string, mythLayer: string, motifs: string[]): SoundMode[] {
  const text = [palette, weather, mythLayer, ...motifs].join(' ').toLowerCase();
  const modes: SoundMode[] = [];
  const add = (mode: SoundMode) => {
    if (!modes.includes(mode)) modes.push(mode);
  };

  if (text.includes('rain') || text.includes('drizzle') || text.includes('storm')) add('rain');
  if (text.includes('wind') || text.includes('mist') || text.includes('fog') || text.includes('haze')) add('wind');
  if (text.includes('fire') || text.includes('ember') || text.includes('candle') || text.includes('smoke') || text.includes('hearth')) add('fire');
  if (text.includes('water') || text.includes('river') || text.includes('shore') || text.includes('boat') || text.includes('ferry') || text.includes('sea')) add('water');
  if (text.includes('grail') || text.includes('queen') || text.includes('court') || text.includes('altar') || text.includes('choir') || text.includes('holy')) add('sacred');
  if (text.includes('void') || text.includes('crypt') || text.includes('vault') || text.includes('underdark') || text.includes('shadow') || text.includes('grave')) add('void');
  if (text.includes('table') || text.includes('gallery') || text.includes('hall') || text.includes('bench') || text.includes('counsel') || text.includes('court')) add('court');
  if (modes.length === 0) add('field');
  return modes.slice(0, 3);
}
