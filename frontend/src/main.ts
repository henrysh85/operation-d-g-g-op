import { createApp } from 'vue';
import { createPinia } from 'pinia';
import App from './App.vue';
import router from './router';
import { useAuthStore } from '@/stores/auth';
import './assets/tokens.css';

const app = createApp(App);
app.use(createPinia());
app.use(router);

// Hydrate the user record from /auth/me on first load if a token already
// exists in localStorage — otherwise the sidebar shows an empty user until
// the next API call.
const auth = useAuthStore();
void auth.hydrate();

app.mount('#app');
