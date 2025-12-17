// main.js
import { h, render } from 'https://esm.sh/preact';
import { PromptList } from './components/prompt-list.js';

render(h(PromptList), document.getElementById('app'));