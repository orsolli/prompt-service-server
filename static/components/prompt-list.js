// components/prompt-list.js
import { h } from 'https://esm.sh/preact';
import { useState, useEffect } from 'https://esm.sh/preact/hooks';
import { generateKeyPair } from '../utils/key-utils.js';
import { hashPublicKey } from '../utils/crypto-utils.js';
import { useKeyStore } from '../utils/storage-utils.js';

export function PromptList() {
    const [keys, addKey, removeKey, loading] = useKeyStore();
    const [activeKey, setActiveKey] = useState(null);
    const [csrfToken, setCsrfToken] = useState(document.headers);
    const [prompts, setPrompts] = useState([]);
    const [error, setError] = useState('');
    const [loadingPrompts, setLoadingPrompts] = useState(false);
    const [sse, setSse] = useState(null);
    const [keySelected, setKeySelected] = useState(false);

    // Check for cookie key and select it
    useEffect(() => {
        if (loading) return;
        const cookie = document.cookie;
        const cookieKey = cookie.match(/publicKey=([^;]+)/);
        if (cookieKey && cookieKey[1]) {
            const publicKey = cookieKey[1];
            hashPublicKey(publicKey).then((publicKeyHash) => {
                const matchingKey = keys.find(k => k.publicKeyHash === publicKeyHash);
                if (matchingKey) {
                    setActiveKey(matchingKey);
                    setKeySelected(true);
                } else {
                    // Redirect to root if no matching key
                    window.location.href = '/?hash=' + publicKeyHash;
                    console.error("Wrong publicKey cookie");
                }
            });
        } else {
            // Redirect to root if no cookie
            window.location.href = '/';
            console.error("Missing publicKey cookie");
        }
    }, [keys, loading]);

    // Fetch prompts when key is selected
    useEffect(() => {
        if (keySelected && activeKey) {
            fetchPrompts();
            setupSSE();
        }
    }, [keySelected, activeKey]);

    // Setup SSE connection
    const setupSSE = () => {
        if (!activeKey) return;
        
        const eventSource = new EventSource(`/api/sse/${activeKey.publicKeyHash}`, {
            withCredentials: true
        });
        
        eventSource.onmessage = (event) => {
            const data = JSON.parse(event.data);
            if (data.type === 'new_prompt') {
                // Add new prompt to the list
                setPrompts(prev => [data, ...prev]);
            } else if (data.type === 'response_received') {
                // Update prompt with response
                setPrompts(prev => 
                    prev.map(prompt => 
                        prompt.id === data.promptId ? {...prompt, response: data.response} : prompt
                    )
                );
            }
        };
        
        eventSource.onerror = (error) => {
            console.error('SSE Error:', error);
        };
        
        setSse(eventSource);
    };

    // Fetch prompts from API
    const fetchPrompts = async () => {
        if (!activeKey) return;
        
        setLoadingPrompts(true);
        try {
            const response = await fetch('/api/prompts', {
                method: 'GET',
                headers: {
                    'Authorization': await signMessage(activeKey, csrfToken)
                }
            });
            
            if (response.ok) {
                const data = await response.json();
                setPrompts(data);
            } else {
                setError('Failed to fetch prompts');
            }
        } catch (err) {
            setError('Error fetching prompts');
        } finally {
            setLoadingPrompts(false);
        }
    };

    // Sign a message with the private key
    const signMessage = async (keyData, message) => {
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

    // Handle key selection
    const handleKeySelect = async (keyData) => {
        setActiveKey(keyData);
        setKeySelected(true);
    };

    // Handle key removal
    const handleRemoveKey = (keyHash) => {
        removeKey(keyHash);
        if (activeKey && activeKey.publicKeyHash === keyHash) {
            setActiveKey(null);
            setKeySelected(false);
        }
    };

    // Handle response submission
    const handleResponse = async (promptId, response) => {
        try {
            const responseText = response.trim();
            if (!responseText) return;
            
            const responseJson = {
                promptId,
                response: responseText
            };
            
            const responseHeaders = {
                'Content-Type': 'application/json',
                'Authorization': await signMessage(activeKey, csrfToken)
            };
            
            const res = await fetch(`/api/prompts/${promptId}`, {
                method: 'POST',
                headers: responseHeaders,
                body: JSON.stringify(responseJson)
            });
            
            if (res.ok) {
                // Update prompt with response
                setPrompts(prev => 
                    prev.map(prompt => 
                        prompt.id === promptId ? {...prompt, response: responseText} : prompt
                    )
                );
            } else {
                setError('Failed to submit response');
            }
        } catch (err) {
            setError('Error submitting response');
        }
    };

    // Handle new key generation
    const handleGenerateKey = async () => {
        try {
            const keyPair = await generateKeyPair();
            const publicKeyB64 = arrayBufferToBase64(keyPair.publicKey);
            const privateKeyB64 = arrayBufferToBase64(keyPair.privateKey);
            const publicKeyHash = await hashPublicKey(publicKeyB64);
            
            const keyData = {
                publicKeyHash: publicKeyHash,
                publicKey: publicKeyB64,
                privateKey: privateKeyB64,
                timestamp: Date.now()
            };
            
            await addKey(keyData);
            document.cookie = `publicKey=${publicKeyB64}; path=/`;
            window.location.href = `/key/${publicKeyHash}`;
        } catch (err) {
            setError('Error generating key');
        }
    };

    // Helper functions
    const arrayBufferToBase64 = (buffer) => {
        const bytes = new Uint8Array(buffer);
        let binary = '';
        for (let i = 0; i < bytes.byteLength; i++) {
            binary += String.fromCharCode(bytes[i]);
        }
        return btoa(binary);
    };

    const base64ToArrayBuffer = (base64) => {
        const binaryString = atob(base64);
        const bytes = new Uint8Array(binaryString.length);
        for (let i = 0; i < binaryString.length; i++) {
            bytes[i] = binaryString.charCodeAt(i);
        }
        return bytes.buffer;
    };

    // Render UI
    if (loading || !keySelected) {
        return h('div', { className: 'container' },
            h('h1', null, 'Prompt Service'),
            h('p', null, 'Loading...')
        );
    }

    return h('div', { className: 'container' },
        h('h1', null, 'Prompt Service'),
        h('h2', null, `Active Key: ${activeKey.publicKey.substring(0, 20)}...`),
        h('div', { className: 'key-actions' },
            h('button', {
                onClick: handleGenerateKey
            }, 'Generate New Key'),
            h('button', {
                onClick: () => {
                    document.cookie = 'publicKey=; path=/; expires=Thu, 01 Jan 1970 00:00:00 UTC;';
                    window.location.href = '/';
                }
            }, 'Switch Key')
        ),
        h('div', { className: 'prompts-container' },
            h('h3', null, 'Your Prompts'),
            loadingPrompts ? h('p', null, 'Loading prompts...') : 
            prompts.length === 0 ? h('p', null, 'No prompts available') : 
            h('div', { className: 'prompts-list' },
                prompts.map(prompt => 
                    h('div', { className: 'prompt-item' },
                        h('div', { className: 'prompt-message' },
                            h('p', null, prompt.message)
                        ),
                        h('div', { className: 'prompt-sender' },
                            h('p', null, `From: ${prompt.senderPublicKey.substring(0, 20)}...`)
                        ),
                        h('div', { className: 'prompt-response' },
                            h('p', null, prompt.response || 'No response yet')
                        ),
                        h('div', { className: 'prompt-actions' },
                            prompt.response ? 
                                h('p', null, 'Response submitted') :
                                h('div', { className: 'response-form' },
                                    h('input', {
                                        type: 'text',
                                        placeholder: 'Enter your response',
                                        id: `response-${prompt.id}`
                                    }),
                                    h('button', {
                                        onClick: () => {
                                            const input = document.getElementById(`response-${prompt.id}`);
                                            handleResponse(prompt.id, input.value);
                                            input.value = '';
                                        }
                                    }, 'Submit')
                                )
                        )
                    )
                )
            )
        )
    );
}