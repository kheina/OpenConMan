<template>
	<div class='logo'>
		<a href='https://github.com/kheina/openconman'>openconman</a>
	</div>
	<div class='sidebar'>
		<nav>
			<ol>
				<li>
					<div/>
					<RouterLink to='/'>Dashboard</RouterLink>
				</li>
				<li>
					<div/>
					<RouterLink to='/containers'>Containers</RouterLink>
					<i class='material-icons-round'>chevron_left</i>
				</li>
				<li>
					<div/>
					<RouterLink to='/services'>Services</RouterLink>
					<i class='material-icons-round'>chevron_left</i>
				</li>
				<li>
					<div/>
					<RouterLink to='/test'>test</RouterLink>
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
nav li {
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
nav li > :first-child {
	width: 5px;
	height: 0;
	background: var(--interact);
	-webkit-transition: var(--transition) var(--fadetime);
	-moz-transition: var(--transition) var(--fadetime);
	-o-transition: var(--transition) var(--fadetime);
	transition: var(--transition) var(--fadetime);
	overflow: hidden;
	position: absolute;
}
nav li > i {
	position: absolute;
	right: 0;
	pointer-events: none;
	font-size: 1em;
	-webkit-transition: var(--transition) var(--fadetime);
	-moz-transition: var(--transition) var(--fadetime);
	-o-transition: var(--transition) var(--fadetime);
	transition: var(--transition) var(--fadetime);
}
nav li:has(a.router-link-active) {
	&> i {
		transform: rotate(-90deg);
	}
	&> ol {
		height: auto;
	}
}
nav:not(:has(a:hover)) li:has(a.router-link-active) > div,
nav li:has(a:hover) > div{
	height: 100%;
}
</style>
