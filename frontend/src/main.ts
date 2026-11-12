import { mount } from 'svelte'
import './styles/index.sass'
import './lib/stores/theme'
import App from './App.svelte'

const app = mount(App, {
  target: document.getElementById('app')!,
})

export default app
