import { createHmac } from "node:crypto";

const BASE32 = "ABCDEFGHIJKLMNOPQRSTUVWXYZ234567";

function base32Decode(text: string): Buffer {
    const bytes: number[] = [];
    let bits = 0;
    let value = 0;
    for (const char of text.toUpperCase().replace(/[\s=]/g, "")) {
        const index = BASE32.indexOf(char);
        if (index < 0) throw new Error(`not base32: ${char}`);
        value = (value << 5) | index;
        bits += 5;
        if (bits >= 8) {
            bits -= 8;
            bytes.push((value >>> bits) & 0xff);
        }
    }
    return Buffer.from(bytes);
}

// totp is the code an authenticator app shows for the secret at the time:
// RFC 6238's defaults, which HivePaaS uses - SHA-1, 30 seconds, 6 digits.
export function totp(secret: string, at = Date.now()): string {
    const counter = Buffer.alloc(8);
    counter.writeBigUInt64BE(BigInt(Math.floor(at / 1000 / 30)));
    const mac = createHmac("sha1", base32Decode(secret)).update(counter).digest();
    const offset = mac[mac.length - 1]! & 0x0f;
    const code = (mac.readUInt32BE(offset) & 0x7fffffff) % 1_000_000;
    return String(code).padStart(6, "0");
}

// A code another than the one of now, nor of the periods beside it.
export function wrongTotp(secret: string): string {
    const near = [-30_000, 0, 30_000].map(shift => totp(secret, Date.now() + shift));
    for (let n = 0; ; n++) {
        const code = String(n).padStart(6, "0");
        if (!near.includes(code)) return code;
    }
}
