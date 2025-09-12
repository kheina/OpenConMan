<template>
	<div class='logo'>
		<a href='https://github.com/kheina/openconman'>openconman</a>
	</div>
	<div class='sidebar'>
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
						<RouterLink to='/user/login'>Login</RouterLink>
					</div>
				</li>
			</ol>
		</nav>
	</div>
	<div class='view'>
		<RouterView/>
	</div>
</template>
<script setup lang='ts'>
import { RouterLink, RouterView } from 'vue-router';
</script>
<style>
body {
	display: grid;
	grid-template-columns: [sidebar-start] 15em [sidebar-end] 0 [view-start] 1fr [view-end];
}
.sidebar {
	grid-area: sidebar;
	max-width: 100%;
	margin: var(--half-margin) var(--border-size) var(--half-margin) 0;
	display: flex;
	align-items: center;
	border-right: var(--border-size) solid var(--border-color);
}
.view {
	grid-area: view;
	height: 100vh;
	max-height: 100vh;
	overflow: auto;
	position: relative;
}
.logo {
	position: absolute;
	top: 0;
	width: 15em;
	height: 3em;
	display: flex;
	align-items: center;
	justify-content: center;
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
