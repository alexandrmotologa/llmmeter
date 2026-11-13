const fs = require('fs');
const path = require('path');
const { Resvg } = require('@resvg/resvg-js');

function buildLogoSvg() {
  return `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 1024 1024" width="1024" height="1024">
  <defs>
    <clipPath id="squircle-clip">
      <rect x="24" y="24" width="976" height="976" rx="220" />
    </clipPath>

    <!-- Emerald / Cyan FinOps Telemetry Glow -->
    <linearGradient id="emerald-optic" x1="0%" y1="0%" x2="100%" y2="100%">
      <stop offset="0%" stop-color="#10b981"/>
      <stop offset="100%" stop-color="#06b6d4"/>
    </linearGradient>

    <linearGradient id="emerald-accent" x1="0%" y1="0%" x2="0%" y2="100%">
      <stop offset="0%" stop-color="#34d399"/>
      <stop offset="100%" stop-color="#059669"/>
    </linearGradient>

    <linearGradient id="mantle-grad-l" x1="0%" y1="0%" x2="100%" y2="100%">
      <stop offset="0%" stop-color="#334155"/>
      <stop offset="100%" stop-color="#0f172a"/>
    </linearGradient>

    <linearGradient id="mantle-grad-r" x1="100%" y1="0%" x2="0%" y2="100%">
      <stop offset="0%" stop-color="#475569"/>
      <stop offset="100%" stop-color="#1e293b"/>
    </linearGradient>

    <filter id="subtle-shadow" x="-10%" y="-10%" width="120%" height="120%">
      <feDropShadow dx="0" dy="14" stdDeviation="18" flood-color="#000000" flood-opacity="0.14" />
    </filter>
  </defs>

  <!-- Luxury White Squircle Container (1024x1024, rx=220) -->
  <rect x="24" y="24" width="976" height="976" rx="220" fill="#ffffff" stroke="#e2e8f0" stroke-width="6" />

  <g clip-path="url(#squircle-clip)">
    <g transform="translate(512, 512)" filter="url(#subtle-shadow)">

      <!-- Hexagonal Architectural Gateway -->
      <polygon points="
        0,-400
        346,-200
        346,200
        0,400
        -346,200
        -346,-200
      " fill="none" stroke="#0f172a" stroke-width="40" stroke-linejoin="round" />

      <!-- Dashed Emerald Telemetry Ring -->
      <polygon points="
        0,-365
        316,-182
        316,182
        0,365
        -316,182
        -316,-182
      " fill="none" stroke="#10b981" stroke-width="5" opacity="0.5" stroke-dasharray="18, 14" />

      <!-- Symmetrical Origami Sentinel Owl (Low-Poly FinOps Mascot) -->
      <!-- 1. Watertight Solid Base Silhouette (Zero-Leak Rule) -->
      <path d="
        M 0 -330
        L 85 -245
        L 175 -315
        L 215 -180
        L 245 -70
        L 230 80
        L 180 230
        L 95 315
        L 0 345
        L -95 315
        L -180 230
        L -230 80
        L -245 -70
        L -215 -180
        L -175 -315
        L -85 -245
        Z
      " fill="#0b0f19" />

      <!-- 2. Mantle & Shoulders (Lower Body Facets) -->
      <!-- Left side facets -->
      <polygon points="0,345 -95,315 -60,200 0,250" fill="#0f172a" />
      <polygon points="-95,315 -180,230 -130,140 -60,200" fill="#1e293b" />
      <polygon points="-180,230 -230,80 -160,50 -130,140" fill="#0f172a" />
      <polygon points="-230,80 -245,-70 -170,-40 -160,50" fill="#1e293b" />
      
      <!-- Right side facets (with slight volumetric highlight) -->
      <polygon points="0,345 95,315 60,200 0,250" fill="#1e293b" />
      <polygon points="95,315 180,230 130,140 60,200" fill="#334155" />
      <polygon points="180,230 230,80 160,50 130,140" fill="#1e293b" />
      <polygon points="230,80 245,-70 170,-40 160,50" fill="#334155" />

      <!-- Chest Telemetry Plate (Chevron Meter Motif) -->
      <polygon points="0,250 -60,200 0,140" fill="#1e293b" />
      <polygon points="0,250 60,200 0,140" fill="#334155" />
      <polygon points="0,140 -60,200 -130,140 -50,80" fill="#0f172a" />
      <polygon points="0,140 60,200 130,140 50,80" fill="#1e293b" />

      <!-- Central Emerald FinOps Vector Core -->
      <polygon points="0,140 -50,80 0,40" fill="#065f46" />
      <polygon points="0,140 50,80 0,40" fill="#047857" />
      <polygon points="0,40 -50,80 0,-10" fill="#059669" />
      <polygon points="0,40 50,80 0,-10" fill="#10b981" />

      <!-- 3. Head & Ear Tufts (Vigilant Owl Horns) -->
      <!-- Left Ear Tuft -->
      <polygon points="-85,-245 -175,-315 -140,-190" fill="#334155" />
      <polygon points="-175,-315 -215,-180 -140,-190" fill="#1e293b" />
      <polygon points="-215,-180 -245,-70 -170,-80 -140,-190" fill="#0f172a" />

      <!-- Right Ear Tuft (Light catch) -->
      <polygon points="85,-245 175,-315 140,-190" fill="#475569" />
      <polygon points="175,-315 215,-180 140,-190" fill="#334155" />
      <polygon points="215,-180 245,-70 170,-80 140,-190" fill="#1e293b" />

      <!-- Forehead & Crown Center -->
      <polygon points="0,-330 -85,-245 0,-190" fill="#1e293b" />
      <polygon points="0,-330 85,-245 0,-190" fill="#334155" />
      <polygon points="-85,-245 -140,-190 0,-190" fill="#0f172a" />
      <polygon points="85,-245 140,-190 0,-190" fill="#1e293b" />

      <!-- Eyebrow Ridge (Vigilant Slant) -->
      <polygon points="0,-190 -140,-190 -110,-130 0,-140" fill="#475569" />
      <polygon points="0,-190 140,-190 110,-130 0,-140" fill="#64748b" />

      <!-- Facial Discs -->
      <!-- Left eye socket plate -->
      <polygon points="-110,-130 -170,-80 -130,-20 -50,-60" fill="#1e293b" />
      <polygon points="-50,-60 -130,-20 -50,-10 0,-60" fill="#0f172a" />
      <!-- Right eye socket plate -->
      <polygon points="110,-130 170,-80 130,-20 50,-60" fill="#334155" />
      <polygon points="50,-60 130,-20 50,-10 0,-60" fill="#1e293b" />

      <!-- Piercing Almond Predator Optics (Eyes) -->
      <!-- Left Eye -->
      <polygon points="-98,-105 -42,-85 -55,-65 -105,-82" fill="#090d16" />
      <polygon points="-92,-100 -48,-83 -58,-68 -98,-84" fill="url(#emerald-optic)" />
      <!-- Left Pupil Slit -->
      <polygon points="-75,-95 -68,-73 -72,-72 -79,-94" fill="#022c22" />

      <!-- Right Eye -->
      <polygon points="98,-105 42,-85 55,-65 105,-82" fill="#090d16" />
      <polygon points="92,-100 48,-83 58,-68 98,-84" fill="url(#emerald-optic)" />
      <!-- Right Pupil Slit -->
      <polygon points="75,-95 68,-73 72,-72 79,-94" fill="#022c22" />

      <!-- Owl Beak (Geometric Obsidian & Cyan Highlight) -->
      <polygon points="0,-140 -22,-60 0,-10" fill="#0f172a" />
      <polygon points="0,-140 22,-60 0,-10" fill="#1e293b" />
      <polygon points="0,-10 -16,40 0,65" fill="#065f46" />
      <polygon points="0,-10 16,40 0,65" fill="#10b981" />

      <!-- Lateral Sensor Wings (Outer Accent Lines) -->
      <line x1="-170" y1="-80" x2="-220" y2="70" stroke="#10b981" stroke-width="3" stroke-linecap="round" opacity="0.7" />
      <line x1="170" y1="-80" x2="220" y2="70" stroke="#10b981" stroke-width="3" stroke-linecap="round" opacity="0.7" />
    </g>
  </g>
</svg>`;
}

async function renderLogo() {
  const outputDir = path.resolve(__dirname, '../docs/images');
  if (!fs.existsSync(outputDir)) {
    fs.mkdirSync(outputDir, { recursive: true });
  }

  const svg = buildLogoSvg();
  const svgPath = path.join(outputDir, 'logo.svg');
  const pngPath = path.join(outputDir, 'logo.png');

  fs.writeFileSync(svgPath, svg, 'utf8');

  const resvg = new Resvg(svg, {
    fitTo: { mode: 'width', value: 1024 },
  });
  const pngData = resvg.render().asPng();
  fs.writeFileSync(pngPath, pngData);

  console.log('✓ Successfully rendered logo.svg and logo.png at 1024x1024 to ' + outputDir);
}

renderLogo().catch((err) => {
  console.error('Logo render failed:', err);
  process.exit(1);
});
