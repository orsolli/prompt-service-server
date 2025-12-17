import { useState, useEffect } from 'https://esm.sh/preact/hooks';
import { hashPublicKey } from '../utils/crypto-utils.js';
// Key storage hook
export function useKeyStore() {
    const [keys, setKeys] = useState([]);
    const [loading, setLoading] = useState(true);
    // Initialize from storage
    useEffect(() => {
        const loadKeys = async () => {
            try {
                // Try localStorage first
                let storedKeys = [];
                const storedKeysStr = localStorage.getItem('promptServiceKeys');
                if (storedKeysStr) {
                    try {
                        storedKeys = JSON.parse(storedKeysStr);
                    } catch (e) {
                        console.warn('Failed to parse localStorage keys:', e);
                    }
                }
                // Check cookies for public key
                const cookie = document.cookie;
                const cookieKey = cookie.match(/publicKey=([^;]+)/);
                if (cookieKey && cookieKey[1]) {
                    const publicKey = cookieKey[1];
                    const publicKeyHash = await hashPublicKey(publicKey);
                    // Check if we already have this key in storage
                    const existingKey = storedKeys.find(key => key.publicKeyHash === publicKeyHash);
                    if (!existingKey) {
                        storedKeys.push({
                            publicKeyHash: publicKeyHash,
                            type: 'cookie',
                            publicKey: publicKey
                        });
                    }
                }
                setKeys(storedKeys);
                setLoading(false);
            } catch (error) {
                console.error('Error loading keys:', error);
                setLoading(false);
            }
        };
        loadKeys();
    }, []);
    // Add key to storage
    const addKey = async (keyData) => {
        // Update state
        setKeys(prev => [...prev, keyData]);
        try {
            // Update localStorage
            const storedKeys = localStorage.getItem('promptServiceKeys');
            let keyList = [];
            if (storedKeys) {
                try {
                    keyList = JSON.parse(storedKeys);
                } catch (e) {
                    console.warn('Failed to parse localStorage keys:', e);
                }
            }
            keyList.push(keyData);
            localStorage.setItem('promptServiceKeys', JSON.stringify(keyList));
        } catch (error) {
            console.error('Error adding key:', error);
        }
    };
    const removeKey = (keyHash) => {
        setKeys(prev => prev.filter(k => k.publicKeyHash !== keyHash));
    };
    return [keys, addKey, removeKey, loading];
}