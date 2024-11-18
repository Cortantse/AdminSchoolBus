// main.js
import { createApp } from 'vue'; // Vue 3 的导入方式
import App from './App.vue';
import router from './router'; // 导入路由

const app = createApp(App);
app.use(router); // 使用路由
app.mount('#app'); // 挂载 Vue 应用
