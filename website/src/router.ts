import { createRouter, createWebHistory } from 'vue-router';
import Dashboard from '@/views/dashboard.vue';

export default createRouter({
	history: createWebHistory(import.meta.env.BASE_URL),
	routes: [
		{
			path: '/',
			name: 'dashboard',
			component: Dashboard,
		},
		{
			path: '/containers',
			name: 'containers',
			component: () => import("./views/containers.vue"),
		},
		{
			path: '/services',
			name: 'services',
			component: () => import("@/views/services.vue"),
		},
		{
			path: '/services/all',
			name: 'all services',
			component: () => import("@/views/all_services.vue"),
		},

		// test view
		{
			path: '/test',
			name: 'test',
			component: () => import("@/views/test.vue"),
		},
	],
});
