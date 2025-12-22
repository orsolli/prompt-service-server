// components/prompt-list.js
import { h } from 'https://esm.sh/preact';
import { useState, useEffect } from 'https://esm.sh/preact/hooks';
import { generateKeyPair } from '../utils/key-utils.js';
import { hashPublicKey } from '../utils/crypto-utils.js';
import { useKeyStore } from '../utils/storage-utils.js';

export function PromptList() {
    const [keys, addKey, removeKey, loading] = useKeyStore();
    const [activeKey, setActiveKey] = useState(null);
    const [challenge, setChallenge] = useState();
    const [prompts, setPrompts] = useState([]);
    const [error, setError] = useState('');
    const [loadingChallenge, setLoadingChallenge] = useState(false);
    const [loadingPrompts, setLoadingPrompts] = useState(false);
    const [sse, setSse] = useState(null);
    const [keySelected, setKeySelected] = useState(false);

    // Check for cookie key and select it
    useEffect(() => {
        if (!keys.length) return;
        const cookie = document.cookie;
        const cookieKey = cookie.match(/publicKey=([^;]+)/);
        if (cookieKey && cookieKey[1]) {
            const publicKey = cookieKey[1];
            hashPublicKey(publicKey).then((publicKeyHash) => {
                const matchingKey = keys.find(k => k.publicKeyHash === publicKeyHash);
                if (matchingKey) {
                    setActiveKey(matchingKey);
                    fetchChallenge(publicKeyHash).then((challenge) => {
                        signMessage(matchingKey, challenge).then((signature) => {
                            setKeySelected(true);
                            document.cookie = `CSRFChallenge=${signature}; path=/api`;
                            setupSSE(publicKeyHash);
                        })
                    });
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
    }, [keys]);

    // Setup SSE connection
    const setupSSE = (publicKeyHash) => {
        const eventSource = new EventSource(`/api/sse/${publicKeyHash}`, {
            withCredentials: true
        });
        
        eventSource.onmessage = function(event) {
            console.log('SSE Message:', event.data);
            const data = JSON.parse(event.data);
            if (data.type === 'connected') {
                fetchPrompts(publicKeyHash);
            } else if (data.type === 'challenge_updated') {
                console.log('Challenge updated, TODO: handle re-authentication');
            } else if (data.type === 'new_prompt') {
                // Add new prompt to the list
                fetchPrompts(publicKeyHash);
            }
        };
        
        eventSource.onerror = (error) => {
            console.error('SSE Error:', error);
        };
        
        setSse(eventSource);
    };

    // Fetch challenge from API
    const fetchChallenge = async (publicKeyHash) => {
        setLoadingChallenge(true);
        try {
            const response = await fetch(`/api/auth/${publicKeyHash}`, {
                method: 'GET',
                credentials: 'same-origin'
            });
            
            if (response.ok) {
                const data = await response.text();
                setChallenge(data);
                return data;
            } else {
                setError('Failed to fetch challenge');
                throw response;
            }
        } catch (err) {
            setError('Error fetching challenge');
            throw err;
        } finally {
            setLoadingChallenge(false);
        }
    };

    // Fetch prompts from API
    const fetchPrompts = async (publicKeyHash) => {
        setLoadingPrompts(true);
        try {
            const response = await fetch(`/api/prompts/${publicKeyHash}`, {
                method: 'GET',
                credentials: 'same-origin'
            });
            
            if (response.ok) {
                const data = await response.json();
                if (!Array.isArray(data)) throw new Error('Invalid prompts data');
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

    // Handle response submission
    const handleResponse = async (promptId, response) => {
        try {
            const responseHeaders = {
                'Content-Type': 'plain/text'
            };
            
            const res = await fetch(`/api/prompts/${promptId}`, {
                method: 'POST',
                headers: responseHeaders,
                body: response,
                credentials: 'same-origin'
            });
            
            if (res.ok) {
                // Update prompt with response
                setPrompts(prev => 
                    prev.map(prompt => 
                        prompt.id === promptId ? {...prompt, response: response || ' '} : prompt
                    )
                );
            } else {
                setError('Failed to submit response');
            }
        } catch (err) {
            setError('Error submitting response');
        }
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
    if (loading || loadingChallenge) {
        return h('div', { className: 'container' },
            h('h1', null, 'Prompt Service'),
            h('p', null, 'Loading...')
        );
    }

    // Copy key to clipboard handler
    const handleCopyKey = () => {
        if (activeKey?.publicKey) {
            navigator.clipboard.writeText(activeKey.publicKey);
        }
    };

    return h('div', { className: 'container' },
        h('h1', null, 'Prompt Service'),
        h('div', { className: 'active-key-row' },
            h('span', {
                className: 'active-key',
                title: activeKey?.publicKey,
                onClick: handleCopyKey,
                tabIndex: 0,
                style: {
                    cursor: 'pointer',
                    outline: 'none',
                }
            }, activeKey?.publicKey || ''),
            h('span', { className: 'copy-hint' }, ' (click to copy)')
        ),
        h('div', { className: 'key-actions' },
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
                        h('hr', null),
                        h('div', { className: 'prompt-message' },
                            h('p', null, prompt.message)
                        ),
                        h('div', { className: 'prompt-actions' },
                            prompt.response ? 
                                h('p', null, '-> ', prompt.response) :
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