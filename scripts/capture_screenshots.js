const fs = require('fs');
const path = require('path');
const puppeteer = require('puppeteer-core');

const EDGE_PATH = 'C:\\Program Files (x86)\\Microsoft\\Edge\\Application\\msedge.exe';
const BASE_URL = 'http://localhost:8085';
const OUT_DIR = path.resolve(__dirname, '../docs/images');
const FRAMES_DIR = path.join(OUT_DIR, 'frames');

async function main() {
  if (!fs.existsSync(OUT_DIR)) fs.mkdirSync(OUT_DIR, { recursive: true });
  if (!fs.existsSync(FRAMES_DIR)) fs.mkdirSync(FRAMES_DIR, { recursive: true });

  console.log('Launching Edge browser at:', EDGE_PATH);
  const browser = await puppeteer.launch({
    executablePath: EDGE_PATH,
    headless: 'new',
    args: ['--no-sandbox', '--disable-setuid-sandbox', '--window-size=1440,900'],
  });

  const page = await browser.newPage();
  await page.setViewport({ width: 1440, height: 900, deviceScaleFactor: 2 });

  console.log('Navigating to', BASE_URL);
  await page.goto(BASE_URL, { waitUntil: 'networkidle0' });
  await new Promise((r) => setTimeout(r, 1500)); // wait for charts to animate

  // 1. Dashboard Overview
  console.log('Capturing 01_dashboard_overview.png...');
  await page.screenshot({ path: path.join(OUT_DIR, '01_dashboard_overview.png') });

  // 2. Tenant Margins Tab
  console.log('Capturing 02_tenant_margins.png...');
  const tenantTab = await page.$('nav button:nth-child(2)');
  if (tenantTab) await tenantTab.click();
  await new Promise((r) => setTimeout(r, 800));
  await page.screenshot({ path: path.join(OUT_DIR, '02_tenant_margins.png') });

  // 3. What-If Simulator Tab
  console.log('Capturing 03_margin_simulator.png...');
  const simTab = await page.$('nav button:nth-child(4)');
  if (simTab) await simTab.click();
  await new Promise((r) => setTimeout(r, 800));
  await page.screenshot({ path: path.join(OUT_DIR, '03_margin_simulator.png') });

  // 4. Prompt Cache ROI Tab
  console.log('Capturing 04_cache_roi.png...');
  const cacheTab = await page.$('nav button:nth-child(5)');
  if (cacheTab) await cacheTab.click();
  await new Promise((r) => setTimeout(r, 800));
  await page.screenshot({ path: path.join(OUT_DIR, '04_cache_roi.png') });

  // 5. Official Tenant Statement
  console.log('Capturing 05_tenant_statement.png...');
  await page.goto(`${BASE_URL}/api/tenants/datapulse/statement?format=html`, { waitUntil: 'networkidle0' });
  await new Promise((r) => setTimeout(r, 1000));
  await page.screenshot({ path: path.join(OUT_DIR, '05_tenant_statement.png') });

  // 6. Mobile Responsive View (390x844)
  console.log('Capturing 06_mobile_responsive.png...');
  await page.setViewport({ width: 390, height: 844, deviceScaleFactor: 2 });
  await page.goto(BASE_URL, { waitUntil: 'networkidle0' });
  await new Promise((r) => setTimeout(r, 1200));
  await page.screenshot({ path: path.join(OUT_DIR, '06_mobile_responsive.png') });

  // 7. Demo GIF Sequence (Frame captures at 1280x720)
  console.log('Capturing frames for demo GIF...');
  await page.setViewport({ width: 1280, height: 720, deviceScaleFactor: 1 });

  // Frame 1: Executive Overview
  await page.goto(BASE_URL, { waitUntil: 'networkidle0' });
  await new Promise((r) => setTimeout(r, 1200));
  await page.screenshot({ path: path.join(FRAMES_DIR, 'frame_01.png') });

  // Frame 2: Hover over negative margin tenant
  const ovTenantTab = await page.$('nav button:nth-child(2)');
  if (ovTenantTab) await ovTenantTab.click();
  await new Promise((r) => setTimeout(r, 800));
  await page.screenshot({ path: path.join(FRAMES_DIR, 'frame_02.png') });

  // Frame 3: What-If Simulator
  const simTab2 = await page.$('nav button:nth-child(4)');
  if (simTab2) await simTab2.click();
  await new Promise((r) => setTimeout(r, 800));
  await page.screenshot({ path: path.join(FRAMES_DIR, 'frame_03.png') });

  // Frame 4: Prompt Cache ROI
  const cacheTab2 = await page.$('nav button:nth-child(5)');
  if (cacheTab2) await cacheTab2.click();
  await new Promise((r) => setTimeout(r, 800));
  await page.screenshot({ path: path.join(FRAMES_DIR, 'frame_04.png') });

  // Frame 5: Downgrade Advisor
  const advTab = await page.$('nav button:nth-child(6)');
  if (advTab) await advTab.click();
  await new Promise((r) => setTimeout(r, 800));
  await page.screenshot({ path: path.join(FRAMES_DIR, 'frame_05.png') });

  // Frame 6: Live Inspector
  const inspTab = await page.$('nav button:nth-child(8)');
  if (inspTab) await inspTab.click();
  await new Promise((r) => setTimeout(r, 800));
  await page.screenshot({ path: path.join(FRAMES_DIR, 'frame_06.png') });

  await browser.close();
  console.log('✓ All real screenshots & frames captured successfully!');
}

main().catch((err) => {
  console.error('Screenshot capture failed:', err);
  process.exit(1);
});
