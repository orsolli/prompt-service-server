import { h } from 'preact';
import { KeyItem } from './key-item.js';
export function KeyList({ keys, removeKey }) {
    if (keys.length === 0) {
        return h('p', null, 'No keys found. Generate a new key below.');
    }
    return h('div', { className: 'key-list' },
        keys.map((key, index) => {
            return h(KeyItem, {
                key: key.publicKeyHash,
                keyData: key,
                removeKey: () => removeKey(key.publicKeyHash)
            });
        })
    );
}