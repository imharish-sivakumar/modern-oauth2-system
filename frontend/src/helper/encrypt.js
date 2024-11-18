import forge from 'node-forge';

const encrypt = (valueToEncrypt) => {
    // Import the public key (in PEM format)
    const publicKeyPem = atob(process.env.REACT_APP_PUBLIC_KEY);
    const publicKey = forge.pki.publicKeyFromPem(publicKeyPem);

    // Encrypt the value with RSA-OAEP and specify SHA-256 as the hash algorithm
    const encrypted = publicKey.encrypt(valueToEncrypt, 'RSA-OAEP', { md: forge.md.sha256.create() });

    // Return the encrypted value in Base64 format
    return forge.util.encode64(encrypted);
};

export default encrypt;
