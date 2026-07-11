declare module 'fastnoise-lite' {
  export default class FastNoiseLite {
    constructor(seed?: number);
    SetSeed(seed: number): void;
    SetNoiseType(type: number): void;
    SetFractalType(type: number): void;
    SetFractalOctaves(octaves: number): void;
    GetNoise(x: number, y: number): number;
    GetNoise(x: number, y: number, z: number): number;

    static NoiseType: {
      OpenSimplex2: number;
      OpenSimplex2S: number;
      Cellular: number;
      Perlin: number;
      ValueCubic: number;
      Value: number;
    };
    static FractalType: {
      None: number;
      FBm: number;
      Ridged: number;
      PingPong: number;
      DomainWarpProgressive: number;
      DomainWarpIndependent: number;
    };
  }
}
