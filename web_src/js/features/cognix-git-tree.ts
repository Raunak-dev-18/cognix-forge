// Pixelated Git tree animation for the GitCode landing and auth pages.
//
// Renders onto any <canvas data-git-tree> element:
//   - a Bayer 8x8 dithered landscape (mountain ridge + drifting mist), drawn in the
//     element's `color` at low opacity, following the Cognix dither guidelines;
//   - a scrolling commit graph (lanes that fork, commit and merge) drawn as crisp pixels,
//     with the newest commit on each lane lit in the accent colour (`--gt-accent`).
//
// Rendering happens on a low-resolution cell grid that is blitted with image smoothing off,
// so the output stays pixel-sharp at any size. The loop pauses while the canvas is off-screen
// or the tab is hidden, and draws a single still frame when the user prefers reduced motion.

const BAYER8 = [
  0, 48, 12, 60, 3, 51, 15, 63,
  32, 16, 44, 28, 35, 19, 47, 31,
  8, 56, 4, 52, 11, 59, 7, 55,
  40, 24, 36, 20, 43, 27, 39, 23,
  2, 50, 14, 62, 1, 49, 13, 61,
  34, 18, 46, 30, 33, 17, 45, 29,
  10, 58, 6, 54, 9, 57, 5, 53,
  42, 26, 38, 22, 41, 25, 37, 21,
].map((v) => (v + 0.5) / 64);

const CELL = 4; // CSS pixels per grid cell
const ROW_GAP = 9; // grid cells between graph rows
const FPS = 30;

type Rgb = [number, number, number];

type GraphRow = {
  lanes: boolean[]; // lanes active at this row
  age: number[]; // rows since each lane was branched
  commit: number; // lane holding this row's commit
  fork?: [number, number]; // [from, to]: `to` branches off `from` towards the next row
  merge?: [number, number]; // [from, to]: `from` merges into `to` towards the next row
  born: number; // timestamp the row appeared
};

function parseColor(value: string, fallback: Rgb): Rgb {
  const v = value.trim();
  let m = /^#([0-9a-f]{6})$/i.exec(v);
  if (m) {
    const n = parseInt(m[1], 16);
    return [(n >> 16) & 255, (n >> 8) & 255, n & 255];
  }
  m = /^#([0-9a-f]{3})$/i.exec(v);
  if (m) {
    return Array.from(m[1], (c) => parseInt(c + c, 16)) as Rgb;
  }
  m = /rgba?\(\s*(\d+)[,\s]+(\d+)[,\s]+(\d+)/i.exec(v);
  if (m) return [Number(m[1]), Number(m[2]), Number(m[3])];
  return fallback;
}

// Cheap deterministic value noise.
function hash(x: number, y: number): number {
  let h = Math.imul(x, 374761393) + Math.imul(y, 668265263);
  h = Math.imul(h ^ (h >>> 13), 1274126177);
  return ((h ^ (h >>> 16)) >>> 0) / 4294967295;
}

function noise2(x: number, y: number): number {
  const xi = Math.floor(x);
  const yi = Math.floor(y);
  const xf = x - xi;
  const yf = y - yi;
  const u = xf * xf * (3 - 2 * xf);
  const v = yf * yf * (3 - 2 * yf);
  const a = hash(xi, yi);
  const b = hash(xi + 1, yi);
  const c = hash(xi, yi + 1);
  const d = hash(xi + 1, yi + 1);
  return a + (b - a) * u + (c - a) * v + (a - b - c + d) * u * v;
}

function nearest(list: number[], from: number): number {
  let best = list[0];
  for (const l of list) if (Math.abs(l - from) < Math.abs(best - from)) best = l;
  return best;
}

function smoothstep(t: number): number {
  return t * t * (3 - 2 * t);
}

class GitTree {
  canvas: HTMLCanvasElement;
  ctx: CanvasRenderingContext2D;
  buffer: HTMLCanvasElement;
  bctx: CanvasRenderingContext2D;
  image: ImageData | null = null;
  graphBuffer: HTMLCanvasElement;
  gctx: CanvasRenderingContext2D;
  graphImage: ImageData | null = null;
  cols = 0;
  rows = 0;
  laneCount = 5;
  laneX: number[] = [];
  graph: GraphRow[] = [];
  scroll = 0; // in cells
  ink: Rgb = [18, 18, 18];
  accent: Rgb = [28, 216, 106];
  inkAlpha = 0.37;
  running = false;
  visible = true;
  last = 0;
  raf = 0;
  seed = Math.random() * 1000;
  reducedMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches;

  constructor(canvas: HTMLCanvasElement) {
    this.canvas = canvas;
    this.ctx = canvas.getContext('2d')!;
    this.buffer = document.createElement('canvas');
    this.bctx = this.buffer.getContext('2d')!;
    this.graphBuffer = document.createElement('canvas');
    this.gctx = this.graphBuffer.getContext('2d')!;
    this.readColors();
    this.resize();

    new ResizeObserver(() => {
      this.resize();
      if (!this.running) this.draw(performance.now());
    }).observe(canvas);

    new IntersectionObserver((entries) => {
      this.visible = entries.some((e) => e.isIntersecting);
      this.toggle();
    }).observe(canvas);

    document.addEventListener('visibilitychange', () => this.toggle());

    // Re-read colours when the theme changes (auto theme follows the OS).
    window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', () => {
      this.readColors();
      if (!this.running) this.draw(performance.now());
    });

    this.draw(performance.now());
    this.toggle();
  }

  readColors() {
    const style = getComputedStyle(this.canvas);
    this.ink = parseColor(style.color, this.ink);
    this.accent = parseColor(style.getPropertyValue('--gt-accent'), this.accent);
    const alpha = parseFloat(style.getPropertyValue('--gt-ink-alpha'));
    if (!Number.isNaN(alpha)) this.inkAlpha = alpha;
  }

  resize() {
    const rect = this.canvas.getBoundingClientRect();
    const cols = Math.max(8, Math.ceil(rect.width / CELL));
    const rows = Math.max(8, Math.ceil(rect.height / CELL));
    if (cols === this.cols && rows === this.rows) return;
    this.cols = cols;
    this.rows = rows;
    this.canvas.width = cols * CELL;
    this.canvas.height = rows * CELL;
    this.buffer.width = cols;
    this.buffer.height = rows;
    this.image = this.bctx.createImageData(cols, rows);
    this.graphBuffer.width = cols;
    this.graphBuffer.height = rows;
    this.graphImage = this.gctx.createImageData(cols, rows);

    // Lanes sit on the right-hand two thirds so overlay copy on the left stays readable.
    const align = this.canvas.dataset.gitTreeAlign ?? 'right';
    this.laneCount = Math.max(3, Math.min(6, Math.floor(cols / 26)));
    const span = Math.min(cols * 0.55, this.laneCount * 14);
    const start = align === 'center' ? (cols - span) / 2 : cols - span - Math.max(6, cols * 0.08);
    this.laneX = Array.from({length: this.laneCount}, (_, i) => Math.round(start + (span / (this.laneCount - 1)) * i));

    this.graph = [];
    this.scroll = 0;
    const needed = Math.ceil(rows / ROW_GAP) + 3;
    for (let i = 0; i < needed; i++) this.pushRow(-1);
  }

  // Adds a row at the top (newest) of the graph.
  pushRow(born: number) {
    const prev = this.graph[0];
    const lanes = prev ? Array.from(prev.lanes) : Array.from({length: this.laneCount}, (_, i) => i === 0);
    // apply the previous row's structural change
    if (prev?.fork) lanes[prev.fork[1]] = true;
    if (prev?.merge) lanes[prev.merge[0]] = false;

    const age = prev ? prev.age.map((a) => a + 1) : Array.from({length: this.laneCount}, () => 0);
    if (prev?.fork) age[prev.fork[1]] = 0;

    const pick = <T>(list: T[]): T => list[Math.floor(Math.random() * list.length)];
    const active = lanes.map((on, i) => (on ? i : -1)).filter((i) => i >= 0);
    const free = lanes.map((on, i) => (on ? -1 : i)).filter((i) => i >= 0);
    const row: GraphRow = {lanes, age, commit: pick(active), born};

    // Branches only merge back once they have a few commits of their own.
    const mergeable = active.filter((l) => l !== 0 && age[l] >= 3);
    const forkChance = active.length < 3 ? 0.4 : 0.14;
    const r = Math.random();
    if (free.length && r < forkChance) {
      const from = pick(active);
      // one of the two nearest free lanes keeps curves short but varied
      const near = Array.from(free).sort((a, b) => Math.abs(a - from) - Math.abs(b - from)).slice(0, 2);
      row.fork = [from, pick(near)];
      row.commit = from;
    } else if (mergeable.length && r < forkChance + 0.22) {
      const from = pick(mergeable);
      const targets = active.filter((l) => l !== from);
      const to = nearest(targets, from);
      row.merge = [from, to];
      row.commit = from;
    }
    this.graph.unshift(row);
  }

  toggle() {
    const should = this.visible && !document.hidden && !this.reducedMotion;
    if (should && !this.running) {
      this.running = true;
      this.last = performance.now();
      const loop = (now: number) => {
        if (!this.running) return;
        this.raf = requestAnimationFrame(loop);
        if (now - this.last < 1000 / FPS) return;
        const dt = Math.min(0.1, (now - this.last) / 1000);
        this.last = now;
        this.step(dt, now);
        this.draw(now);
      };
      this.raf = requestAnimationFrame(loop);
    } else if (!should && this.running) {
      this.running = false;
      cancelAnimationFrame(this.raf);
    }
  }

  step(dt: number, now: number) {
    this.scroll += dt * 4.2; // cells per second
    while (this.scroll >= ROW_GAP) {
      this.scroll -= ROW_GAP;
      this.pushRow(now);
      const maxRows = Math.ceil(this.rows / ROW_GAP) + 4;
      if (this.graph.length > maxRows) this.graph.length = maxRows;
    }
  }

  draw(now: number) {
    if (!this.image || !this.graphImage) return;
    const cols = this.cols;
    const rows = this.rows;
    const data = this.image.data;
    data.fill(0);
    const gdata = this.graphImage.data;
    gdata.fill(0);
    const t = (now / 1000) * 0.08 + this.seed; // Cognix mist speed 0.08

    const [ir, ig, ib] = this.ink;
    const inkA = Math.round(this.inkAlpha * 255);

    // ---- dithered landscape ----
    const ridgeBase = rows * 0.62;
    for (let x = 0; x < cols; x++) {
      const ridge = ridgeBase -
        noise2(x * 0.018 + this.seed, 0) * rows * 0.22 -
        noise2(x * 0.07 + this.seed, 3) * rows * 0.05;
      const ridge2 = rows * 0.74 - noise2(x * 0.03 + this.seed + 9, 7) * rows * 0.14;
      for (let y = 0; y < rows; y++) {
        let v = 0;
        // distant ridge: soft fill below the line, fading downward
        if (y > ridge) v = 0.18 + Math.min(0.32, (y - ridge) / rows * 0.9);
        // near ridge: denser
        if (y > ridge2) v = Math.max(v, 0.42 + Math.min(0.4, (y - ridge2) / rows * 1.4));
        // mist wisps drifting across (strength 0.32, wisps 0.14)
        const mist = noise2(x * 0.045 + t * 6, y * 0.09 - t * 2) * noise2(x * 0.02 - t * 3, y * 0.05);
        v += (mist - 0.2) * 0.32 * 1.6;
        // 14% top fade
        const top = y / rows;
        if (top < 0.14) v *= top / 0.14;
        if (v > BAYER8[(x & 7) + (y & 7) * 8]) {
          const i = (y * cols + x) * 4;
          data[i] = ir;
          data[i + 1] = ig;
          data[i + 2] = ib;
          data[i + 3] = inkA;
        }
      }
    }

    // ---- commit graph ----
    const plot = (x: number, y: number, rgb: Rgb, a = 255) => {
      const px = Math.round(x);
      const py = Math.round(y);
      if (px < 0 || py < 0 || px >= cols || py >= rows) return;
      const i = (py * cols + px) * 4;
      gdata[i] = rgb[0];
      gdata[i + 1] = rgb[1];
      gdata[i + 2] = rgb[2];
      gdata[i + 3] = a;
    };
    const solidInk: Rgb = this.ink;
    const lineA = 230;

    // S-curve from (x0, yStart) up to (x1, yEnd), filling horizontal gaps so it stays connected.
    const curve = (x0: number, x1: number, yStart: number, yEnd: number) => {
      let prev = Math.round(x0);
      for (let y = yStart; y >= yEnd; y--) {
        const p = (yStart - y) / (yStart - yEnd || 1);
        const x = Math.round(x0 + (x1 - x0) * smoothstep(p));
        const lo = Math.min(prev, x);
        const hi = Math.max(prev, x);
        for (let xx = lo; xx <= hi; xx++) plot(xx, y, solidInk, lineA);
        prev = x;
      }
    };

    const rowY = (i: number) => Math.round(i * ROW_GAP + this.scroll + ROW_GAP * 0.5);
    const headSeen = new Set<number>();

    for (let i = 0; i < this.graph.length; i++) {
      const y0 = rowY(i);
      const older = this.graph[i + 1];
      if (!older) continue;
      const y1 = rowY(i + 1);
      // Edges run from the older row (y1, below) up to this row (y0).
      for (let lane = 0; lane < this.laneCount; lane++) {
        if (!older.lanes[lane]) continue;
        const merging = older.merge && older.merge[0] === lane;
        const target = merging ? older.merge![1] : lane;
        curve(this.laneX[lane], this.laneX[target], y1, y0);
      }
      if (older.fork) {
        const [from, to] = older.fork;
        curve(this.laneX[from], this.laneX[to], y1, y0);
      }
    }

    for (let i = 0; i < this.graph.length; i++) {
      const row = this.graph[i];
      const cx = this.laneX[row.commit];
      const cy = rowY(i);
      const isHead = !headSeen.has(row.commit);
      headSeen.add(row.commit);
      const age = row.born < 0 ? 10 : (now - row.born) / 1000;
      const color = isHead ? this.accent : solidInk;

      // halo ring for fresh commits (accent halftone, fades out)
      if (age < 1.6 && !this.reducedMotion) {
        const radius = 2 + age * 4;
        const fade = 1 - age / 1.6;
        for (let a = 0; a < Math.PI * 2; a += 0.35 / radius) {
          const hx = cx + Math.cos(a) * radius;
          const hy = cy + Math.sin(a) * radius;
          if (((Math.round(hx) + Math.round(hy)) & 1) === 0) plot(hx, hy, this.accent, Math.round(220 * fade));
        }
      }

      // commit node: 3x3 square, hollow for merges
      const hollow = Boolean(row.merge);
      for (let dx = -1; dx <= 1; dx++) {
        for (let dy = -1; dy <= 1; dy++) {
          if (hollow && dx === 0 && dy === 0) continue;
          plot(cx + dx, cy + dy, color);
        }
      }
      if (isHead) {
        // pixel "crosshair" accent on lane heads
        plot(cx - 2, cy, color, 160);
        plot(cx + 2, cy, color, 160);
        plot(cx, cy - 2, color, 160);
        plot(cx, cy + 2, color, 160);
      }
    }

    this.bctx.putImageData(this.image, 0, 0);
    this.ctx.imageSmoothingEnabled = false;
    this.ctx.clearRect(0, 0, this.canvas.width, this.canvas.height);
    // Draw each cell as a (CELL-1)px dot with a 1px gap for the dot-matrix look.
    this.ctx.drawImage(this.buffer, 0, 0, this.canvas.width, this.canvas.height);
    this.ctx.globalCompositeOperation = 'destination-out';
    this.ctx.fillStyle = this.gapPattern();
    this.ctx.fillRect(0, 0, this.canvas.width, this.canvas.height);
    this.ctx.globalCompositeOperation = 'source-over';
    // The commit graph sits on top as solid pixels (no dot gaps) so branches read clearly.
    this.gctx.putImageData(this.graphImage, 0, 0);
    this.ctx.drawImage(this.graphBuffer, 0, 0, this.canvas.width, this.canvas.height);
  }

  pattern: CanvasPattern | null = null;
  gapPattern(): CanvasPattern {
    if (this.pattern) return this.pattern;
    const tile = document.createElement('canvas');
    tile.width = CELL;
    tile.height = CELL;
    const tctx = tile.getContext('2d')!;
    tctx.fillStyle = '#000';
    tctx.fillRect(CELL - 1, 0, 1, CELL);
    tctx.fillRect(0, CELL - 1, CELL, 1);
    this.pattern = this.ctx.createPattern(tile, 'repeat')!;
    return this.pattern;
  }
}

export function initCognixGitTree() {
  for (const el of document.querySelectorAll<HTMLCanvasElement>('canvas[data-git-tree]')) {
    if (el.dataset.gitTreeReady) continue;
    el.dataset.gitTreeReady = 'true';
    new GitTree(el);
  }
}

// "/" focuses the header search, like GitHub.
export function initNavbarSearchShortcut() {
  const input = document.querySelector<HTMLInputElement>('#navbar .cx-nav-search input');
  if (!input) return;
  document.addEventListener('keydown', (e) => {
    if (e.key !== '/' || e.ctrlKey || e.metaKey || e.altKey) return;
    const target = e.target as HTMLElement;
    if (target.closest('input, textarea, select, [contenteditable="true"], .cm-editor, .monaco-editor')) return;
    e.preventDefault();
    input.focus();
    input.select();
  });
}
