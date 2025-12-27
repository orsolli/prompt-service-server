import { h } from 'preact';
export function ImportForm({
    importText,
    setImportText,
    publicKey,
    setPublicKey,
    importError,
    setImportError,
    handleImportKey
}) {
    return h('div', { className: 'import-form' },
        h('h3', null, 'Import Key'),
        h('div', { className: 'import-instructions' },
            h('h4', null, 'Instructions:'),
            h('ul', null,
                h('li', null, 'Paste your private key in the format shown below'),
                h('li', null, 'The key should be in PEM format with BEGIN/END markers'),
                h('li', null, 'Optionally, you can also paste the public key')
            )
        ),
        h('textarea', {
            value: importText,
            onChange: (e) => setImportText(e.target.value),
            placeholder: `-----BEGIN OPENSSH PRIVATE KEY-----
AAB3NzaC1yc2EAAAADAQABAAABAQDf4a9b1d2e7d6a2f3a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0
-----END OPENSSH PRIVATE KEY-----`
        }),
        h('div', { className: 'public-key-input' },
            h('label', null, 'Public Key (optional):'),
            h('input', {
                type: 'text',
                value: publicKey,
                onChange: (e) => setPublicKey(e.target.value),
                placeholder: 'ssh-ed25519 AAAAC4NzaC1yc2EA1d2e7d+G4a3b4c5d6e7f8a9b0c1 root@localhost',
                style: { width: '100%', marginTop: '0.5rem' }
            })
        ),
        importError && h('p', { className: 'error' }, importError),
        h('button', {
            onClick: handleImportKey,
            disabled: (!importText.trim() && !publicKey.trim())
        }, 'Import Key')
    );
}