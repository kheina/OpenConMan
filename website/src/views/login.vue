<template>
	<div class='centered'>
		<main>
			<div class='field'>
				<span>Username</span>
				<input placeholder='username' name='username' v-model='username' class='interactable text'>
			</div>
			<div class='field'>
				<span>Password</span>
				<div>
					<input placeholder='password' type='password' name='password' v-model='password' @keydown.enter='sendLogin' autocomplete='off' class='interactable text'>
					<button @click='sendLogin' @keydown.enter='sendLogin' class='interactable'>Submit »</button>
				</div>
			</div>
		</main>
	</div>
</template>
<script setup lang='ts'>
import { ref, type Ref } from 'vue';
import { auth } from '@/globals';
import { cetch } from '@/utilities';

const username: Ref<string> = ref("");
const password: Ref<string> = ref("");

function sendLogin() {
	cetch(`/v1/user/login`, {
		method: "POST",
		body: JSON.stringify({
			username: username.value,
			password: password.value,
		}),
	}).then((r: Response) => {
		switch (r.status) {
		case 200:
			return r.json();
		case 400:
		case 401:
		case 404:
		default:
			// TODO: create toast
			throw r;
		}
	}).then((r: { token: string, expires: string }) => {
		auth.value = r.token;
		const maxage = Math.round((new Date(r.expires).valueOf() - new Date().valueOf()) / 1000);
		document.cookie = `ocm-auth=${r.token}; max-age=${maxage}; samesite=strict; path=/; secure`;
	}).catch(
		console.error
	);
}
</script>
<style scoped>
.centered {
	display: flex;
	align-items: center;
	justify-content: center;
	height: 100%;
}

.field {
	display: flex;
	flex-direction: column;
	margin-bottom: var(--margin);

	span {
		margin-left: var(--margin);
	}
	&:last-child {
		margin: 0;
	}
}

button {
	color: inherit;
	cursor: pointer;
	border-radius: var(--border-radius);
	padding: 0.5em 1em;
	border: var(--border-size) solid var(--border-color);
	display: inline-block;
	font-weight: normal;
	background: var(--bg1);
	box-sizing: border-box;
	box-shadow: 0 2px 3px 1px var(--shadowcolor);
	margin-bottom: 1em;
	-webkit-transition: var(--transition) var(--fadetime);
	-moz-transition: var(--transition) var(--fadetime);
	-o-transition: var(--transition) var(--fadetime);
	transition: var(--transition) var(--fadetime);
	margin-left: var(--margin);
}
.service button {
	background: var(--bg2);
}
button:hover {
	border-color: var(--borderhover);
	color: var(--interact);
	box-shadow: 0 0 10px 3px var(--activeshadowcolor);
}

input {
	color: inherit;
	cursor: text;
	border-radius: var(--border-radius);
	padding: 0.5em 1em;
	border: var(--border-size) solid var(--border-color);
	display: inline-block;
	font-weight: normal;
	background: var(--bg1);
	box-sizing: border-box;
	box-shadow: 0 2px 3px 1px var(--shadowcolor);
	outline: none;

	&, &::placeholder {
		-webkit-transition: var(--transition) var(--fadetime);
		-moz-transition: var(--transition) var(--fadetime);
		-o-transition: var(--transition) var(--fadetime);
		transition: var(--transition) var(--fadetime);
	}
	&:hover::placeholder {
		color: var(--interact);
	}
	&:hover {
		border-color: var(--borderhover);
		color: var(--interact);
		box-shadow: 0 0 10px 3px var(--activeshadowcolor);
	}
	&:focus {
		color: var(--text);
		border-color: var(--interact);
		&::placeholder {
			color: #eeeeee20;
		}
	}
}
</style>
