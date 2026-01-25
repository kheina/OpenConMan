<template>
	<div class='sidebar' v-if='auth'>
		<div class='logo'>
			<a href='https://github.com/kheina/openconman'>openconman</a>
		</div>
		<nav>
			<ol>
				<li>
					<div>
						<div/>
						<RouterLink to='/'>Dashboard</RouterLink>
					</div>
				</li>
				<li>
					<div>
						<div/>
						<RouterLink to='/containers'>Containers</RouterLink>
						<i class='material-icons-round'>chevron_left</i>
					</div>
				</li>
				<li>
					<div>
						<div/>
						<RouterLink to='/services'>Services</RouterLink>
						<i class='material-icons-round'>chevron_left</i>
					</div>
					<ol>
						<li>
							<div>
								<div/>
								<RouterLink to='/services/all'>All Services</RouterLink>
							</div>
						</li>
					</ol>
				</li>
				<li>
					<div>
						<div/>
						<RouterLink to='/test'>test</RouterLink>
					</div>
				</li>
				<li>
					<div>
						<div/>
						<RouterLink to='/user'>User</RouterLink>
					</div>
				</li>
			</ol>
		</nav>
		<div v-if='update?.newer' class='update'>
			update available
		</div>
	</div>
	<div class='view'>
		<RouterView/>
	</div>
</template>
<script setup lang='ts'>
import { RouterLink, RouterView, useRoute, useRouter } from 'vue-router';
import { ref, watch, type Ref } from 'vue';
import { auth } from '@/globals';
import { cetch, GetCookie } from '@/utilities';

console.debug("mode:", import.meta.env.MODE);

interface Update {
	current: string,
	latest:  string,
	assert?: string,
	dev?:    boolean,
	newer?:  boolean,
}

const router = useRouter();
const route = useRoute();
const update: Ref<undefined | Update> = ref();

auth.value = GetCookie("ocm-auth");
if (auth.value) checkForUpdate();

function checkForUpdate() {
	cetch(
		"/v1/pkg/self"
	).then(
		r => r.json()
	).then((r: Update) =>
		update.value = r
	).catch(
		console.error
	);
}

watch(auth, (auth: string | undefined) => {
	if (auth) {
		router.replace(route.query?.path?.toString() ?? "/");
		checkForUpdate();
	} else {
		router.replace("/user/login?path=" + encodeURIComponent(route.fullPath));
	}
});
</script>
<style>
body {
	display: flex;
	flex-direction: row;
}
.sidebar {
	flex-shrink: 0;
	width: 15em;
	margin: var(--half-margin) var(--border-size) var(--half-margin) 0;
	display: flex;
	align-items: center;
	border-right: var(--border-size) solid var(--border-color);
	position: relative;
}
.view {
	height: 100vh;
	max-height: 100vh;
	overflow: auto;
	position: relative;
	flex-grow: 1;
	width: 100%;
}
.logo, .update {
	position: absolute;
	width: 100%;
	height: 3em;
	display: flex;
	align-items: center;
	justify-content: center;
}
.logo {
	top: var(--neg-half-margin);
}
.update {
	bottom: var(--neg-half-margin);
}
nav {
	width: 100%;
}
nav li > div {
	display: flex;
	height: 2.5em;
	align-items: center;
	position: relative;
}
nav a {
	display: flex;
	padding-left: var(--margin);
	text-decoration: none;
	height: 100%;
	align-items: center;
	flex-grow: 1;
	position: relative;
}
nav li > div > :first-child {
	width: 0.25em;
	height: 0;
	background: var(--interact);
	-webkit-transition: var(--transition) var(--fadetime);
	-moz-transition: var(--transition) var(--fadetime);
	-o-transition: var(--transition) var(--fadetime);
	transition: var(--transition) var(--fadetime);
	overflow: hidden;
	position: absolute;
}
nav li div > i {
	position: absolute;
	right: 0;
	pointer-events: none;
	font-size: 1em;
	-webkit-transition: var(--transition) var(--fadetime);
	-moz-transition: var(--transition) var(--fadetime);
	-o-transition: var(--transition) var(--fadetime);
	transition: var(--transition) var(--fadetime);
}
nav li > ol {
	max-height: 0;
	overflow: hidden;
	-webkit-transition: var(--transition) var(--fadetime);
	-moz-transition: var(--transition) var(--fadetime);
	-o-transition: var(--transition) var(--fadetime);
	transition: var(--transition) var(--fadetime);
	a {
		padding-left: calc(var(--margin) * 2);
	}
}
nav li:has(a.router-link-active) {
	& i {
		transform: rotate(-90deg);
	}
	&> ol {
		/* idk how to make this animate */
		max-height: unset;
	}
}
nav:not(:has(a:hover)) li div:has(a.router-link-active) > :first-child,
nav li div:has(a:hover) > :first-child {
	height: 2.5em;
}
</style>
