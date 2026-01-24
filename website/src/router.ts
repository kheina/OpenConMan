import { createRouter, createWebHistory } from "vue-router";
import Dashboard from "@/views/dashboard.vue";

export default createRouter({
	history: createWebHistory(import.meta.env.BASE_URL),
	routes: [
		{
			path: "/",
			name: "dashboard",
			component: Dashboard,
		},
		{
			path: "/containers",
			name: "containers",
			component: () => import("./views/containers.vue"),
		},
		{
			path: "/services",
			name: "services",
			component: () => import("@/views/services.vue"),
		},
		{
			path: "/service/:name",
			name: "service",
			props: true,
			component: () => import("@/views/service.vue"),
		},
		{
			path: "/services/all",
			name: "all services",
			component: () => import("@/views/all_services.vue"),
		},
		{
			path: "/user/login",
			name: "login",
			component: () => import("@/views/login.vue"),
		},
		{
			path: "/user",
			name: "user",
			component: () => import("@/views/user.vue"),
		},

		// test view
		{
			path: "/test",
			name: "test",
			component: () => import("@/views/test.vue"),
		},
	],
});
