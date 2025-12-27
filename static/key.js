// main.js
import { h, render } from 'preact';
import { PromptList } from './components/prompt-list.js';

render(h(PromptList), document.getElementById('app'));