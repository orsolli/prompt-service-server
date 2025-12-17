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