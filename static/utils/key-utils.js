// Generate Ed25519 keypair
export async function generateKeyPair() {
    const keyPair = await crypto.subtle.generateKey(
        {
            name: "Ed25519",
        },
        true,
        ["sign", "verify"],
    );
    // Export public and private keys
    const publicKey = await crypto.subtle.exportKey("raw", keyPair.publicKey);
    const privateKey = await crypto.subtle.exportKey("pkcs8", keyPair.privateKey);
    return {
        publicKey: publicKey,
        privateKey: privateKey
    };
}

// Convert base64 string to ArrayBuffer
const base64ToArrayBuffer = (base64) => {
    const binaryString = atob(base64);
    const bytes = new Uint8Array(binaryString.length);
    for (let i = 0; i < binaryString.length; i++) {
        bytes[i] = binaryString.charCodeAt(i);
    }
    return bytes.buffer;
};

// Sign a message with the private key
export const signMessage = async (keyData, message) => {
    const encoder = new TextEncoder();
    const data = encoder.encode(message);
    
    // Get the private key from localStorage or sessionStorage
    const privateKey = keyData.privateKey;
    const key = await crypto.subtle.importKey(
        "pkcs8",
        base64ToArrayBuffer(privateKey),
        { name: "Ed25519" },
        false,
        ["sign"]
    );
    
    const signature = await crypto.subtle.sign("Ed25519", key, data);
    return btoa(String.fromCharCode(...new Uint8Array(signature)));
};