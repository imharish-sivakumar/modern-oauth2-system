import { Buffer } from 'node:buffer';
import crypto from 'node:crypto';

export class RsaService {
    /**
     * @param key Public key in PEM format base64 encoded
     */
    constructor(key) {
        this.passwordPublicKey = this.pemB64ToDer(key);
    }

    async encrypt(plaintext) {
        return this.encryptWithKey(plaintext, this.passwordPublicKey);
    }

    /**
     * Encrypts the plaintext with the given key.
     * @param plaintext Plaintext to encrypt
     * @param key Key to encrypt the plaintext with, in DER format
     * @returns Encrypted plaintext in base64 format
     */
    async encryptWithKey(plaintext, key) {
        const cryptoKey = await this.importKey(key);
        const plaintextBuffer = new TextEncoder().encode(plaintext);

        const encrypted = await crypto.webcrypto.subtle.encrypt(
            {
                name: 'RSA-OAEP'
            },
            cryptoKey,
            plaintextBuffer
        );

        return Buffer.from(encrypted).toString('base64');
    }

    /**
     * Decodes a base64 encoded string.
     * @param str base64 encoded string
     * @returns Decoded string in latin1 format
     * @see {@link https://nodejs.org/api/buffer.html#buffers-and-character-encodings}
     */
    decode(str) {
        return Buffer.from(str, 'base64').toString('latin1');
    }

    /**
     * Converts a PEM key in base64 format to DER format.
     * @param pem PEM key in base64 format
     * @returns DER equivalent of the PEM key
     */
    pemB64ToDer(pem) {
        const PEM_HEADER = '-----BEGIN PUBLIC KEY-----';
        const PEM_FOOTER = '-----END PUBLIC KEY-----';

        const pemString = this.decode(pem);
        const pemBody = pemString
            .substring(PEM_HEADER.length, pemString.length - PEM_FOOTER.length)
            .trim();

        const binaryDerString = this.decode(pemBody);
        const binaryDer = new Uint8Array(binaryDerString.length);
        for (let i = 0; i < binaryDerString.length; i++) {
            binaryDer[i] = binaryDerString.charCodeAt(i);
        }

        return binaryDer;
    }

    /**
     * Converts DER format key to a CryptoKey object.
     * @param key in DER format
     * @returns CryptoKey object of the passed DER key
     */
    async importKey(key) {
        return crypto.webcrypto.subtle.importKey(
            'spki',
            key,
            {
                name: 'RSA-OAEP',
                hash: 'SHA-256'
            },
            false,
            ['encrypt']
        );
    }
}
