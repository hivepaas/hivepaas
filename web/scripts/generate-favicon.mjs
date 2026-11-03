import fs from "node:fs";
import path from "node:path";
import zlib from "node:zlib";

function crc32(buf) {
  let crc = -1;
  for (let i = 0; i < buf.length; i++) {
    crc ^= buf[i];
    for (let j = 0; j < 8; j++) {
      crc = (crc >>> 1) ^ (-(crc & 1) & 0xedb88320);
    }
  }
  return (crc ^ -1) >>> 0;
}

function makeChunk(type, data) {
  const len = Buffer.alloc(4);
  len.writeUInt32BE(data.length, 0);
  const typeBuf = Buffer.from(type, "ascii");
  const crcBuf = Buffer.alloc(4);
  const c = crc32(Buffer.concat([typeBuf, data]));
  crcBuf.writeUInt32BE(c, 0);
  return Buffer.concat([len, typeBuf, data, crcBuf]);
}

function encodePNG(width, height, rgbaBuffer) {
  const sig = Buffer.from([0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a]);
  const ihdrData = Buffer.alloc(13);
  ihdrData.writeUInt32BE(width, 0);
  ihdrData.writeUInt32BE(height, 4);
  ihdrData[8] = 8;
  ihdrData[9] = 6;
  ihdrData[10] = 0;
  ihdrData[11] = 0;
  ihdrData[12] = 0;
  const ihdr = makeChunk("IHDR", ihdrData);

  const scanlines = Buffer.alloc(height * (1 + width * 4));
  for (let y = 0; y < height; y++) {
    const rowOffset = y * (1 + width * 4);
    scanlines[rowOffset] = 0;
    rgbaBuffer.copy(scanlines, rowOffset + 1, y * width * 4, (y + 1) * width * 4);
  }

  const idatData = zlib.deflateSync(scanlines, { level: 9 });
  const idat = makeChunk("IDAT", idatData);
  const iend = makeChunk("IEND", Buffer.alloc(0));

  return Buffer.concat([sig, ihdr, idat, iend]);
}

function hexToRgb(hex) {
  hex = hex.replace("#", "");
  if (hex.length === 3) {
    hex = hex
      .split("")
      .map((c) => c + c)
      .join("");
  }
  const num = parseInt(hex, 16);
  return [(num >> 16) & 255, (num >> 8) & 255, num & 255];
}

export function renderSvgCirclesToRgba(svgContent, targetWidth, targetHeight) {
  const vbMatch = svgContent.match(/viewBox=["']([^"']+)["']/);
  let vbX = 0,
    vbY = 0,
    vbW = 36,
    vbH = 36;
  if (vbMatch) {
    const parts = vbMatch[1].trim().split(/\s+/).map(Number);
    if (parts.length === 4) {
      [vbX, vbY, vbW, vbH] = parts;
    }
  }

  const scaleX = targetWidth / vbW;
  const scaleY = targetHeight / vbH;

  const accR = new Float32Array(targetWidth * targetHeight);
  const accG = new Float32Array(targetWidth * targetHeight);
  const accB = new Float32Array(targetWidth * targetHeight);
  const accA = new Float32Array(targetWidth * targetHeight);

  // 16x16 subpixel supersampling for high fidelity downscaling
  const SAMPLES = 16;
  const subStep = 1 / SAMPLES;

  const circleRegex = /<circle\s+([^>]+)\/?>/g;
  let match;
  while ((match = circleRegex.exec(svgContent)) !== null) {
    const attrs = match[1];
    const cxMatch = attrs.match(/cx=["']([^"']+)["']/);
    const cyMatch = attrs.match(/cy=["']([^"']+)["']/);
    const rMatch = attrs.match(/r=["']([^"']+)["']/);
    const fillMatch = attrs.match(/fill=["']([^"']+)["']/);

    if (cxMatch && cyMatch && rMatch && fillMatch) {
      const cx = (parseFloat(cxMatch[1]) - vbX) * scaleX;
      const cy = (parseFloat(cyMatch[1]) - vbY) * scaleY;
      const r = parseFloat(rMatch[1]) * scaleX;
      const [cr, cg, cb] = hexToRgb(fillMatch[1]);
      const r2 = r * r;

      const minX = Math.max(0, Math.floor(cx - r - 2));
      const maxX = Math.min(targetWidth - 1, Math.ceil(cx + r + 2));
      const minY = Math.max(0, Math.floor(cy - r - 2));
      const maxY = Math.min(targetHeight - 1, Math.ceil(cy + r + 2));

      for (let y = minY; y <= maxY; y++) {
        for (let x = minX; x <= maxX; x++) {
          let insideCount = 0;
          for (let sy = 0; sy < SAMPLES; sy++) {
            const py = y + (sy + 0.5) * subStep;
            const dy = py - cy;
            const dy2 = dy * dy;
            for (let sx = 0; sx < SAMPLES; sx++) {
              const px = x + (sx + 0.5) * subStep;
              const dx = px - cx;
              if (dx * dx + dy2 <= r2) {
                insideCount++;
              }
            }
          }

          if (insideCount > 0) {
            const alpha = insideCount / (SAMPLES * SAMPLES);
            const pIdx = y * targetWidth + x;
            const destA = accA[pIdx];
            const outA = alpha + destA * (1 - alpha);
            if (outA > 0) {
              accR[pIdx] = (cr * alpha + accR[pIdx] * destA * (1 - alpha)) / outA;
              accG[pIdx] = (cg * alpha + accG[pIdx] * destA * (1 - alpha)) / outA;
              accB[pIdx] = (cb * alpha + accB[pIdx] * destA * (1 - alpha)) / outA;
              accA[pIdx] = outA;
            }
          }
        }
      }
    }
  }

  const rgbaBuf = Buffer.alloc(targetWidth * targetHeight * 4);
  for (let i = 0; i < targetWidth * targetHeight; i++) {
    rgbaBuf[i * 4 + 0] = Math.round(accR[i]);
    rgbaBuf[i * 4 + 1] = Math.round(accG[i]);
    rgbaBuf[i * 4 + 2] = Math.round(accB[i]);
    rgbaBuf[i * 4 + 3] = Math.round(accA[i] * 255);
  }

  return encodePNG(targetWidth, targetHeight, rgbaBuf);
}

export function createIcoFromPngs(pngList) {
  const count = pngList.length;
  const header = Buffer.alloc(6);
  header.writeUInt16LE(0, 0); // reserved
  header.writeUInt16LE(1, 2); // type: 1 = ICO
  header.writeUInt16LE(count, 4); // count

  const dirEntries = [];
  let currentOffset = 6 + count * 16;

  for (const item of pngList) {
    const entry = Buffer.alloc(16);
    entry[0] = item.width >= 256 ? 0 : item.width;
    entry[1] = item.height >= 256 ? 0 : item.height;
    entry[2] = 0; // color count
    entry[3] = 0; // reserved
    entry.writeUInt16LE(1, 4); // color planes
    entry.writeUInt16LE(32, 6); // bits per pixel
    entry.writeUInt32LE(item.buffer.length, 8); // size in bytes
    entry.writeUInt32LE(currentOffset, 12); // offset
    dirEntries.push(entry);
    currentOffset += item.buffer.length;
  }

  return Buffer.concat([header, ...dirEntries, ...pngList.map((item) => item.buffer)]);
}

// Main execution
const sourceSvg = path.resolve("/Users/tnt/go/src/github.com/hivepaas/hivepaas-website/landing/public/favicon.svg");
const fallbackSvg = path.resolve("src/assets/icons/logo/logo.svg");
const svgPath = fs.existsSync(sourceSvg) ? sourceSvg : fallbackSvg;

console.log(`Generating favicon from: ${svgPath}`);
const svgContent = fs.readFileSync(svgPath, "utf8");

const png16 = renderSvgCirclesToRgba(svgContent, 16, 16);
const png32 = renderSvgCirclesToRgba(svgContent, 32, 32);
const png48 = renderSvgCirclesToRgba(svgContent, 48, 48);

const icoBuffer = createIcoFromPngs([
  { width: 32, height: 32, buffer: png32 },
  { width: 16, height: 16, buffer: png16 },
  { width: 48, height: 48, buffer: png48 },
]);

const destinations = [
  // Website landing
  { ico: "/Users/tnt/go/src/github.com/hivepaas/hivepaas-website/landing/public/favicon.ico", png32: "/Users/tnt/go/src/github.com/hivepaas/hivepaas-website/landing/public/favicon-32x32.png", svg: "/Users/tnt/go/src/github.com/hivepaas/hivepaas-website/landing/public/favicon.svg" },
  // Website dist (if exists)
  { ico: "/Users/tnt/go/src/github.com/hivepaas/hivepaas-website/landing/dist/favicon.ico", png32: "/Users/tnt/go/src/github.com/hivepaas/hivepaas-website/landing/dist/favicon-32x32.png", svg: "/Users/tnt/go/src/github.com/hivepaas/hivepaas-website/landing/dist/favicon.svg" },
  // Website docs
  { ico: "/Users/tnt/go/src/github.com/hivepaas/hivepaas-website/docs/static/img/favicon.ico", svg: "/Users/tnt/go/src/github.com/hivepaas/hivepaas-website/docs/static/img/favicon.svg" },
  // Dashboard public
  { ico: path.resolve("public/favicon.ico"), png32: path.resolve("public/favicon-32x32.png"), svg: path.resolve("public/favicon.svg") },
];

for (const dest of destinations) {
  if (dest.ico && (fs.existsSync(path.dirname(dest.ico)) || fs.existsSync(dest.ico))) {
    fs.writeFileSync(dest.ico, icoBuffer);
    console.log(`Updated: ${dest.ico}`);
  }
  if (dest.png32 && (fs.existsSync(path.dirname(dest.png32)) || fs.existsSync(dest.png32))) {
    fs.writeFileSync(dest.png32, png32);
    console.log(`Updated: ${dest.png32}`);
  }
  if (dest.svg && (fs.existsSync(path.dirname(dest.svg)) || fs.existsSync(dest.svg))) {
    fs.writeFileSync(dest.svg, svgContent);
    console.log(`Updated: ${dest.svg}`);
  }
}

console.log("Favicon generation completed successfully!");
