import { h } from 'https://esm.sh/preact';
import { useState } from 'https://esm.sh/preact/hooks';
import { KeyList } from './key-list.js';
import { ImportForm } from './import-form.js';
import { useKeyStore } from '../utils/storage-utils.js';
import { generateKeyPair } from '../utils/key-utils.js';
import { hashPublicKey } from '../utils/crypto-utils.js';
export function App() {
    const [loading, keys, addKey, removeKey] = useKeyStore();
    const [importOpen, setImportOpen] = useState(false);
    const [importText, setImportText] = useState('');
    const [publicKey, setPublicKey] = useState('');
    const [importError, setImportError] = useState('');
    // Generate new key pair
    const handleGenerateKey = async () => {
        try {
            const keyPair = await generateKeyPair();
            // Convert to base64 for storage
            const publicKeyB64 = arrayBufferToBase64(keyPair.publicKey);
            const privateKeyB64 = arrayBufferToBase64(keyPair.privateKey);
            // Hash the public key
            const publicKeyHash = await hashPublicKey(publicKeyB64);
            // Create key data
            const keyData = {
                publicKeyHash: publicKeyHash,
                publicKey: publicKeyB64,
                privateKey: privateKeyB64,
                timestamp: Date.now()
            };
            // Add to storage
            await addKey(keyData);
            // Set cookie with public key
            document.cookie = `publicKey=${publicKeyB64}; path=/`;
            // Redirect to key page
            window.location.href = `/key/${publicKeyHash}`;
        } catch (error) {
            console.error('Error generating key:', error);
        }
    };
    // Import key from text
    const handleImportKey = async () => {
        if (!importText.trim() && !publicKey.trim()) return;
        try {
            setImportError('');
            // Parse the import text
            let parsedData = {};
            try {
                // Try to parse as SSH key format
                const keyMatch = importText.match(/-----BEGIN OPENSSH PRIVATE KEY-----(.*?)-----END OPENSSH PRIVATE KEY-----(.*?)/s);
                if (keyMatch) {
                    parsedData['privateKey'] = keyMatch[1].replace(/\s/g, '');
                } else {
                    throw new Error('Invalid SSH key format');
                }
            } catch (e) {
                throw new Error('Failed to parse private key');
            }
            try {
                // Try to parse as SSH key format
                const keyMatch = importText.match(/ssh-.\ (.*?)\ (.*?)/s);
                if (keyMatch) {
                    parsedData['privateKey'] = keyMatch[1].replace(/\s/g, '');
                } else {
                    throw new Error('Invalid SSH key format');
                }
            } catch (e) {
                throw new Error('Failed to parse private key');
            }
            // Validate structure
            if (!parsedData.privateKey && !parsedData.publicKey) {
                throw new Error('Missing privateKey or publicKey in import data');
            }
            // Use the imported private key instead of generating a new one
            const publicKeyB64 = parsedData.publicKey;
            const privateKeyB64 = parsedData.privateKey;
            // Hash the public key
            const publicKeyHash = await hashPublicKey(publicKeyB64);
            // Add to storage
            const keyImportData = {
                publicKeyHash: publicKeyHash,
                publicKey: publicKeyB64,
                privateKey: privateKeyB64,
                timestamp: Date.now()
            };
            await addKey(keyImportData);
            // Close import form
            setImportOpen(false);
            setImportText('');
            setPublicKey('');
            // Set cookie with public key
            document.cookie = `publicKey=${publicKeyB64}; path=/`;
            // Redirect to key page
            window.location.href = `/key/${publicKeyHash}`;
        } catch (error) {
            console.error('Error importing key:', error);
            setImportError('Error importing key: ' + error.message);
        }
    };
    // Base64 encode/decode functions
    function arrayBufferToBase64(buffer) {
        const bytes = new Uint8Array(buffer);
        let binary = '';
        for (let i = 0; i < bytes.byteLength; i++) {
            binary += String.fromCharCode(bytes[i]);
        }
        return btoa(binary);
    }
    // Render main app
    return h('div', { className: 'container' },
        h('h1', null, 'Prompt Service'),
        h('p', null, 'Secure decentralized prompt service using asymmetric cryptography'),
        h('div', { className: 'generate-btn' },
            h('button', {
                onClick: handleGenerateKey,
                disabled: loading
            }, loading ? 'Generating...' : 'Generate New Key')
        ),
        h('div', { className: 'import-btn' },
            h('button', {
                onClick: () => setImportOpen(!importOpen)
            }, importOpen ? 'Cancel Import' : 'Import Existing Key')
        ),
        importOpen && h(ImportForm, {
            importText,
            setImportText,
            publicKey,
            setPublicKey,
            importError,
            setImportError,
            handleImportKey
        }),
        h('h2', null, 'Your Keys'),
        h(KeyList, { keys, removeKey }),
        h('p', null,
            'Keys are stored locally in your browser. Your private key is never sent to any server.'
        )
    );
}