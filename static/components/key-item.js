import { h } from 'preact';
export function KeyItem({ keyData, removeKey }) {
    const keyHash = keyData.publicKeyHash;
    return h('div', { className: 'key-item' },
        [
            h('button', {
                className: 'key-button outline',
                onClick: () => {
                    document.cookie = `publicKey=${keyData.publicKey}; path=/`;
                    window.location.href = `/key/${keyHash}`;
                },
                key: `link-${keyData.publicKey}`
            }, `${keyData.publicKey}`),
            h('button', {
                className: 'outline secondary',
                onClick: removeKey,
                key: `remove-${keyData.publicKey}`
            }, 'Remove')
        ]
    );
}