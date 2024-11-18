import { createRouter, createWebHistory } from 'vue-router'; // Vue Router 4 的导入方式
import LoginPage from '../views/login.vue'; // 确保路径正确
import HomePage from '../views/HomePage.vue';

const routes = [
    {
        path: '/',
        redirect: '/login', // 访问根路径时重定向到 /login
    },
    {
        path: '/login',
        name: 'Login',
        component: LoginPage,
    },
    {
        path: '/home',
        name: 'Home',
        component: HomePage,
    },
];

const router = createRouter({
    history: createWebHistory(), // 使用 history 模式
    routes,
});

// 添加导航守卫，确保用户只有在登录后才能访问主页
router.beforeEach((to, from, next) => {
    const isAuthenticated = !!localStorage.getItem('jwtToken');

    if (to.name === 'Login' && isAuthenticated) {
        // 如果用户已登录并且尝试访问登录页面，重定向到主页
        next({ name: 'Home' });
    } else if (to.name !== 'Login' && !isAuthenticated) {
        // 如果用户未登录并且尝试访问受保护的页面，则重定向到登录页面
        next({ name: 'Login' });
    } else {
        next(); // 继续导航
    }
});

export default router;
