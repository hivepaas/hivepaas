import fs from "node:fs";
import path from "node:path";
import zlib from "node:zlib";

/**
 * High-quality SVG to PNG rasterizer for simple geometric SVG icons
 * Supports circle and polygon elements with 8x8 subpixel supersampling antialiasing.
 */
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
  ihdrData[8] = 8; // bit depth
  ihdrData[9] = 6; // color type: RGBA
  ihdrData[10] = 0; // compression
  ihdrData[11] = 0; // filter
  ihdrData[12] = 0; // interlace
  const ihdr = makeChunk("IHDR", ihdrData);

  const scanlines = Buffer.alloc(height * (1 + width * 4));
  for (let y = 0; y < height; y++) {
    const rowOffset = y * (1 + width * 4);
    scanlines[rowOffset] = 0; // Filter: None
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

export function convertSvgToPng(svgPath, outPath, targetWidth = 512, targetHeight = 512) {
  const svgContent = fs.readFileSync(svgPath, "utf8");

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

  const SAMPLES = 8;
  const subStep = 1 / SAMPLES;

  // 1. Process circles
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

  // Convert to 8-bit RGBA
  const rgbaBuf = Buffer.alloc(targetWidth * targetHeight * 4);
  for (let i = 0; i < targetWidth * targetHeight; i++) {
    rgbaBuf[i * 4 + 0] = Math.round(accR[i]);
    rgbaBuf[i * 4 + 1] = Math.round(accG[i]);
    rgbaBuf[i * 4 + 2] = Math.round(accB[i]);
    rgbaBuf[i * 4 + 3] = Math.round(accA[i] * 255);
  }

  const outPng = encodePNG(targetWidth, targetHeight, rgbaBuf);
  fs.writeFileSync(outPath, outPng);
  console.log(`Generated ${outPath} (${targetWidth}x${targetHeight}) from ${svgPath}`);
}

const defaultSvg = path.resolve("src/assets/icons/logo/logo.svg");
const defaultOut = path.resolve("src/assets/icons/logo/logo-512x512.png");
convertSvgToPng(defaultSvg, defaultOut, 512, 512);
