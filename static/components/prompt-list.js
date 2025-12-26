// components/prompt-list.js
import { h } from 'https://esm.sh/preact';
import { useState } from 'https://esm.sh/preact/hooks';
import { signMessage } from '../utils/key-utils.js';
import { hashPublicKey } from '../utils/crypto-utils.js';
import { useKeyStore } from '../utils/storage-utils.js';

export function PromptList() {
    const [activeKey, setActiveKey] = useState(null);
    const [prompts, setPrompts] = useState([]);
    const [error, setError] = useState('');
    const [loadingChallenge, setLoadingChallenge] = useState(false);
    const [loadingPrompts, setLoadingPrompts] = useState(false);
    const [sseConnection, setSSEConnection] = useState(null);

    // Fetch challenge from API
    const fetchChallenge = async (publicKeyHash) => {
        const response = await fetch(`/api/auth/${publicKeyHash}`, {
            method: 'GET',
            credentials: 'same-origin'
        });
        
        if (response.ok) {
            const data = await response.text();
            return data;
        } else {
            setError('Failed to fetch challenge');
            throw response;
        }
    };

    const signChallenge = async (activeKey) => {
        setLoadingChallenge(true);
        try {
            const challenge = await fetchChallenge(activeKey.publicKeyHash);
            const signature = await signMessage(activeKey, challenge);
            document.cookie = `CSRFChallenge=${signature}; path=/api; max-age=300`;
            if (!sseConnection || sseConnection.readyState === EventSource.CLOSED) {
                setSSEConnection(setupSSE(activeKey.publicKeyHash));
            }
        } catch (error) {
            setError('Failed to sign challenge');
        } finally {
            setLoadingChallenge(false);
        }
    };

    const initializeKey = async (keys) => {
        const cookie = document.cookie;
        const cookieKey = cookie.match(/publicKey=([^;]+)/);
        if (cookieKey && cookieKey[1]) {
            const publicKey = cookieKey[1];
            const publicKeyHash = await hashPublicKey(publicKey);
            const matchingKey = keys.find(k => k.publicKeyHash === publicKeyHash);
            if (matchingKey) {
                setActiveKey(matchingKey);
                signChallenge(matchingKey)
                setInterval(() => signChallenge(matchingKey), 4 * 60 * 1000); // Refresh every 4 minutes
            } else {
                // Redirect to root if no matching key
                window.location.href = '/?hash=' + publicKeyHash;
                console.error("Wrong publicKey cookie");
            }
        } else {
            // Redirect to root if no cookie
            window.location.href = '/';
            console.error("Missing publicKey cookie");
        }
    };

    const [loading] = useKeyStore(initializeKey);

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
                fetchPrompts(publicKeyHash);
            } else if (data.type === 'prompt_responded') {
                const [ promptId, ...response ] = data.content.split(':');
                setPrompts(prev => 
                    prev.map(prompt => 
                        prompt.id === promptId ? {...prompt, response: response.join(':') || ' '} : prompt
                    )
                );
            }
        };
        
        eventSource.onerror = (error) => {
            console.error('SSE Error:', error);
        };

        return eventSource;
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

    // Render UI
    if (loading || loadingChallenge) {
        return h('div', { className: 'container' },
            h('h1', null, 'Prompt Service'),
            h('p', null, 'Loading...')
        );
    }

    return h('div', { className: 'container' },
        h('h1', null, 'Prompt Service'),
        activeKey?.publicKey ? h('div', { className: 'active-key-row' },
            h('span', {
                className: 'active-key',
                title: activeKey.publicKey,
                onClick: () => navigator.clipboard.writeText(activeKey.publicKey),
                tabIndex: 0,
                style: {
                    cursor: 'pointer',
                    outline: 'none',
                }
            }, activeKey.publicKey || ''),
            h('span', { className: 'copy-hint' }, ' (click to copy)')
        ) : null,
        h('div', { className: 'key-actions' },
            h('button', {
                onClick: () => {
                    document.cookie = 'publicKey=; path=/; expires=Thu, 01 Jan 1970 00:00:00 UTC;';
                    window.location.href = '/';
                }
            }, 'Switch Key')
        ),
        h('div', null,
            h('p', null, error)
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