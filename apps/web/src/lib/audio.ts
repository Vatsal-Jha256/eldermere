import * as Tone from 'tone';
import { buildAtmosphereProfile, type RoomAtmosphere, type AtmosphereProfile, hashText, mulberry32 } from './atmosphere';

type Disposables = Array<{ stop?: () => void; disconnect?: () => void; dispose?: () => void }>;

export class AtmosphereAudio {
  private activeNodes: Disposables = [];
  private loops: Tone.Loop[] = [];
  private masterVolume: Tone.Volume;
  private currentKey = '';
  private currentProfile: AtmosphereProfile | null = null;
  private started = false;

  constructor() {
    this.masterVolume = new Tone.Volume(-8).toDestination();
  }

  async start() {
    if (this.started) return;
    await Tone.start();
    Tone.Transport.start();
    this.started = true;
    this.applyProfile(this.currentProfile ?? buildAtmosphereProfile(null));
  }

  stopAll() {
    for (const loop of this.loops) {
      loop.stop();
      loop.dispose();
    }
    this.loops = [];

    for (const node of this.activeNodes) {
      try {
        node.stop?.();
      } catch {
        // Ignore teardown errors from already-finished synths.
      }
      try {
        node.disconnect?.();
      } catch {
        // Ignore teardown errors from already-disconnected nodes.
      }
      try {
        node.dispose?.();
      } catch {
        // Ignore teardown errors from already-disposed nodes.
      }
    }
    this.activeNodes = [];
  }

  updateMood(room: RoomAtmosphere) {
    const profile = buildAtmosphereProfile(room);
    this.currentProfile = profile;

    if (this.currentKey === profile.key) return;
    this.currentKey = profile.key;

    if (this.started) {
      this.applyProfile(profile);
    }
  }

  playCue(kind: string, text = '') {
    if (!this.started) return;

    const seed = hashText([this.currentKey, kind, text.slice(0, 64)].join('|'));
    const rng = mulberry32(seed);
    const low = this.currentProfile?.modes.includes('void') || kind === 'error' ? 'brown' : 'sine';
    const velocity = kind === 'error' ? 0.24 : 0.34 + rng() * 0.18;
    const duration = kind === 'fight' ? '8n' : kind === 'story' ? '4n' : '16n';

    if (kind === 'fight' || kind === 'error') {
      const noise = new Tone.NoiseSynth({
        noise: { type: low === 'brown' ? 'brown' : 'pink' },
        envelope: { attack: 0.002, decay: 0.08 + rng() * 0.05, sustain: 0, release: 0.12 }
      });
      const filter = new Tone.Filter({
        type: 'highpass',
        frequency: kind === 'fight' ? 1400 : 900,
        rolloff: -24
      });
      const volume = new Tone.Volume(kind === 'fight' ? -16 : -22);
      noise.chain(filter, volume, this.masterVolume);
      this.activeNodes.push(noise, filter, volume);

      noise.triggerAttackRelease(duration, Tone.now(), velocity);
      this.disposeLater(noise, 800);
      this.disposeLater(filter, 800);
      this.disposeLater(volume, 800);
      return;
    }

    const synth = new Tone.Synth({
      oscillator: { type: kind === 'move' ? 'triangle' : 'sine' },
      envelope: {
        attack: 0.01,
        decay: 0.18,
        sustain: 0.05,
        release: 0.45
      }
    });
    const delay = new Tone.FeedbackDelay('8n', 0.28);
    const reverb = new Tone.Freeverb({ roomSize: 0.78, dampening: 2800 });
    const volume = new Tone.Volume(-24);
    synth.chain(delay, reverb, volume, this.masterVolume);
    this.activeNodes.push(synth, delay, reverb, volume);

    const note = kind === 'party'
      ? pick(['E4', 'G4', 'B4', 'D5'], rng)
      : kind === 'quest'
        ? pick(['D4', 'F4', 'A4', 'C5'], rng)
        : kind === 'story'
          ? pick(['C4', 'G4', 'C5'], rng)
          : kind === 'move'
            ? pick(['A3', 'D4', 'F4'], rng)
            : pick(['E3', 'A3', 'C4'], rng);

    synth.triggerAttackRelease(note, duration, Tone.now(), velocity);
    this.disposeLater(synth, 1200);
    this.disposeLater(delay, 1200);
    this.disposeLater(reverb, 1200);
    this.disposeLater(volume, 1200);
  }

  private applyProfile(profile: AtmosphereProfile) {
    this.stopAll();

    const rng = mulberry32(profile.seed);
    const layers = profile.modes;

    const baseVolume = layers.includes('void') ? -20 : layers.includes('field') ? -26 : -23;
    const baseCutoff = layers.includes('void') ? 240 : layers.includes('sacred') ? 850 : 520;

    const bed = new Tone.Noise(layers.includes('fire') ? 'brown' : 'pink').start();
    const bedFilter = new Tone.Filter({
      type: layers.includes('sacred') ? 'bandpass' : 'lowpass',
      frequency: baseCutoff,
      rolloff: -24
    });
    const bedVolume = new Tone.Volume(baseVolume);
    const bedLfo = new Tone.LFO(0.04 + rng() * 0.08, baseCutoff * 0.7, baseCutoff * 1.5).start();
    bedLfo.connect(bedFilter.frequency);
    bed.chain(bedFilter, bedVolume, this.masterVolume);
    this.activeNodes.push(bed, bedFilter, bedVolume, bedLfo);

    if (layers.includes('rain')) {
      this.addRainLayer(profile, rng);
    }
    if (layers.includes('wind')) {
      this.addWindLayer(profile, rng);
    }
    if (layers.includes('fire')) {
      this.addFireLayer(profile, rng);
    }
    if (layers.includes('water')) {
      this.addWaterLayer(profile, rng);
    }
    if (layers.includes('sacred')) {
      this.addSacredLayer(profile, rng);
    }
    if (layers.includes('void')) {
      this.addVoidLayer(profile, rng);
    }
    if (layers.includes('court')) {
      this.addCourtLayer(profile, rng);
    }
    if (layers.length === 1 && layers[0] === 'field') {
      this.addFieldLayer(profile, rng);
    }
  }

  private addRainLayer(profile: AtmosphereProfile, rng: () => number) {
    const rain = new Tone.Noise('pink').start();
    const rainFilter = new Tone.Filter({
      type: 'lowpass',
      frequency: 1100 + rng() * 600,
      rolloff: -24
    });
    const rainPanner = new Tone.AutoPanner(0.08 + rng() * 0.12).start();
    const rainVolume = new Tone.Volume(-18);
    rain.chain(rainFilter, rainPanner, rainVolume, this.masterVolume);
    this.activeNodes.push(rain, rainFilter, rainPanner, rainVolume);

    const grain = new Tone.NoiseSynth({
      noise: { type: 'pink' },
      envelope: { attack: 0.001, decay: 0.08, sustain: 0, release: 0.1 }
    });
    const grainFilter = new Tone.Filter({
      type: 'highpass',
      frequency: 2400 + rng() * 600,
      rolloff: -12
    });
    const grainVolume = new Tone.Volume(-22);
    grain.chain(grainFilter, grainVolume, this.masterVolume);
    this.activeNodes.push(grain, grainFilter, grainVolume);

    const interval = profile.weather.toLowerCase().includes('storm') ? '16n' : '8n';
    const loop = new Tone.Loop((time) => {
      if (rng() > 0.18) {
        const duration = rng() > 0.65 ? '16n' : '8n';
        grain.triggerAttackRelease(duration, time, 0.16 + rng() * 0.22);
      }
    }, interval).start(0);
    this.loops.push(loop);
  }

  private addWindLayer(profile: AtmosphereProfile, rng: () => number) {
    const wind = new Tone.Noise('brown').start();
    const windFilter = new Tone.Filter({
      type: 'lowpass',
      frequency: 260 + rng() * 220,
      rolloff: -24
    });
    const windLfo = new Tone.LFO(0.03 + rng() * 0.06, 90, 640 + rng() * 180).start();
    windLfo.connect(windFilter.frequency);
    const windVolume = new Tone.Volume(-20);
    wind.chain(windFilter, windVolume, this.masterVolume);
    this.activeNodes.push(wind, windFilter, windLfo, windVolume);

    const gust = new Tone.NoiseSynth({
      noise: { type: 'brown' },
      envelope: { attack: 0.002, decay: 0.18, sustain: 0, release: 0.32 }
    });
    const gustFilter = new Tone.Filter({
      type: 'bandpass',
      frequency: 420 + rng() * 220,
      Q: 1.2,
      rolloff: -12
    });
    const gustVolume = new Tone.Volume(-24);
    gust.chain(gustFilter, gustVolume, this.masterVolume);
    this.activeNodes.push(gust, gustFilter, gustVolume);

    const loop = new Tone.Loop((time) => {
      if (rng() > 0.42) {
        gust.triggerAttackRelease('8n', time, 0.12 + rng() * 0.18);
      }
    }, '2n').start(0);
    this.loops.push(loop);
  }

  private addFireLayer(profile: AtmosphereProfile, rng: () => number) {
    const crackle = new Tone.NoiseSynth({
      noise: { type: 'brown' },
      envelope: { attack: 0.001, decay: 0.05, sustain: 0, release: 0.08 }
    });
    const crackleFilter = new Tone.Filter({
      type: 'highpass',
      frequency: 2400 + rng() * 900,
      rolloff: -24
    });
    const crackleDelay = new Tone.FeedbackDelay('32n', 0.18);
    const crackleVolume = new Tone.Volume(-20);
    crackle.chain(crackleFilter, crackleDelay, crackleVolume, this.masterVolume);
    this.activeNodes.push(crackle, crackleFilter, crackleDelay, crackleVolume);

    const ember = new Tone.Noise('pink').start();
    const emberFilter = new Tone.Filter({
      type: 'lowpass',
      frequency: 150 + rng() * 90,
      rolloff: -24
    });
    const emberLfo = new Tone.AutoFilter({
      frequency: 0.11 + rng() * 0.05,
      depth: 0.8,
      baseFrequency: 90 + rng() * 35
    }).start();
    const emberVolume = new Tone.Volume(-26);
    ember.chain(emberFilter, emberLfo, emberVolume, this.masterVolume);
    this.activeNodes.push(ember, emberFilter, emberLfo, emberVolume);

    const interval = profile.palette === 'inferno' ? '8n' : '4n';
    const loop = new Tone.Loop((time) => {
      if (rng() > 0.2) {
        crackle.triggerAttackRelease('16n', time, 0.14 + rng() * 0.22);
      }
    }, interval).start(0);
    this.loops.push(loop);
  }

  private addWaterLayer(profile: AtmosphereProfile, rng: () => number) {
    const wash = new Tone.Noise('pink').start();
    const washFilter = new Tone.Filter({
      type: 'lowpass',
      frequency: 700 + rng() * 220,
      rolloff: -24
    });
    const washLfo = new Tone.AutoFilter({
      frequency: 0.06 + rng() * 0.03,
      depth: 0.7,
      baseFrequency: 180 + rng() * 60
    }).start();
    const washVolume = new Tone.Volume(-21);
    wash.chain(washFilter, washLfo, washVolume, this.masterVolume);
    this.activeNodes.push(wash, washFilter, washLfo, washVolume);

    const drop = new Tone.NoiseSynth({
      noise: { type: 'white' },
      envelope: { attack: 0.001, decay: 0.06, sustain: 0, release: 0.18 }
    });
    const dropFilter = new Tone.Filter({
      type: 'bandpass',
      frequency: 2100 + rng() * 500,
      Q: 2.4,
      rolloff: -12
    });
    const dropVolume = new Tone.Volume(-24);
    drop.chain(dropFilter, dropVolume, this.masterVolume);
    this.activeNodes.push(drop, dropFilter, dropVolume);

    const loop = new Tone.Loop((time) => {
      if (rng() > 0.4) {
        drop.triggerAttackRelease('16n', time, 0.12 + rng() * 0.2);
      }
    }, '2n').start(0);
    this.loops.push(loop);
  }

  private addSacredLayer(profile: AtmosphereProfile, rng: () => number) {
    const choir = new Tone.PolySynth(Tone.Synth, {
      oscillator: { type: 'sine' },
      envelope: { attack: 0.3, decay: 0.8, sustain: 0.25, release: 3.5 }
    });
    const delay = new Tone.FeedbackDelay('8n.', 0.45);
    const reverb = new Tone.Freeverb({ roomSize: 0.88, dampening: 3200 });
    const volume = new Tone.Volume(-23);
    choir.chain(delay, reverb, volume, this.masterVolume);
    this.activeNodes.push(choir, delay, reverb, volume);

    const notes = profile.mythLayer.toLowerCase().includes('grail')
      ? ['C5', 'D5', 'E5', 'G5', 'A5']
      : ['D5', 'F5', 'A5', 'C6'];
    const loop = new Tone.Loop((time) => {
      if (rng() > 0.24) {
        const note = pick(notes, rng);
        choir.triggerAttackRelease(note, '4n', time, 0.16 + rng() * 0.14);
      }
    }, '2n').start(0);
    this.loops.push(loop);
  }

  private addVoidLayer(profile: AtmosphereProfile, rng: () => number) {
    const drone = new Tone.FMSynth({
      harmonicity: 0.5,
      modulationIndex: 2.5,
      oscillator: { type: 'sine' },
      modulation: { type: 'triangle' },
      envelope: { attack: 0.04, decay: 0.4, sustain: 0.85, release: 3.8 }
    });
    const delay = new Tone.FeedbackDelay('2n.', 0.58);
    const reverb = new Tone.Freeverb({ roomSize: 0.96, dampening: 1400 });
    const volume = new Tone.Volume(-18);
    drone.chain(delay, reverb, volume, this.masterVolume);
    this.activeNodes.push(drone, delay, reverb, volume);

    const notes = profile.palette === 'eldritch-void' ? ['C1', 'G1', 'A1'] : ['D1', 'F1', 'C2'];
    const loop = new Tone.Loop((time) => {
      const note = pick(notes, rng);
      drone.triggerAttackRelease(note, '1m', time, 0.28 + rng() * 0.12);
    }, '2m').start(0);
    this.loops.push(loop);
  }

  private addCourtLayer(profile: AtmosphereProfile, rng: () => number) {
    const court = new Tone.PolySynth(Tone.AMSynth, {
      harmonicity: 2.2,
      oscillator: { type: 'triangle' },
      envelope: { attack: 0.08, decay: 0.45, sustain: 0.18, release: 1.8 }
    });
    const delay = new Tone.FeedbackDelay('4n', 0.33);
    const reverb = new Tone.Freeverb({ roomSize: 0.82, dampening: 2400 });
    const volume = new Tone.Volume(-25);
    court.chain(delay, reverb, volume, this.masterVolume);
    this.activeNodes.push(court, delay, reverb, volume);

    const scales = profile.palette === 'tavern-red' || profile.palette === 'coin-shadow'
      ? ['D3', 'F3', 'A3', 'C4', 'E4']
      : ['E3', 'G3', 'B3', 'D4', 'F4'];
    const loop = new Tone.Loop((time) => {
      if (rng() > 0.24) {
        const note = pick(scales, rng);
        court.triggerAttackRelease(note, '8n', time, 0.18 + rng() * 0.1);
      }
    }, '1m').start(0);
    this.loops.push(loop);
  }

  private addFieldLayer(profile: AtmosphereProfile, rng: () => number) {
    const pluck = new Tone.PolySynth(Tone.AMSynth, {
      harmonicity: 2.8,
      oscillator: { type: 'triangle' },
      envelope: { attack: 0.06, decay: 0.35, sustain: 0.08, release: 1.2 }
    });
    const delay = new Tone.FeedbackDelay('8n', 0.24);
    const reverb = new Tone.Freeverb({ roomSize: 0.75, dampening: 1800 });
    const volume = new Tone.Volume(-27);
    pluck.chain(delay, reverb, volume, this.masterVolume);
    this.activeNodes.push(pluck, delay, reverb, volume);

    const notes = ['D3', 'F3', 'A3', 'C4', 'E4'];
    const loop = new Tone.Loop((time) => {
      if (rng() > 0.28) {
        const note = pick(notes, rng);
        pluck.triggerAttackRelease(note, '8n', time, 0.14 + rng() * 0.14);
      }
    }, '2n').start(0);
    this.loops.push(loop);
  }

  private disposeLater(node: { dispose?: () => void }, delayMs: number) {
    setTimeout(() => {
      try {
        node.dispose?.();
      } catch {
        // Ignore late dispose errors.
      }
    }, delayMs);
  }
}

function pick<T>(items: T[], rng: () => number): T {
  return items[Math.floor(rng() * items.length) % items.length];
}
