import type { Router } from 'vue-router';
import { auth } from '@/globals';

import drip1 from '$/sounds/drip1.wav';
import drip2 from '$/sounds/drip2.wav';
import drip3 from '$/sounds/drip3.wav';
import drip4 from '$/sounds/drip4.wav';


let prev: number | undefined;
const drips = [
	drip1,
	drip2,
	drip3,
	drip4,
];
export function Notify() {
	for (; ;) {
		const r = Math.floor(Math.random() * 4);
		if (r !== prev) {
			new Audio(drips[r]).play();
			prev = r;
			return;
		}
	}
}

export function GetCookie(cookieName: string, default_value: any = null) {
	const name = cookieName + "=";
	const ca = document.cookie.split(";");
	let value: any = default_value;
	for (let i = 0; i < ca.length; i++) {
		const c = ca[i].trimStart();
		if (c.startsWith(name)) {
			value = decodeURIComponent(c.substring(name.length, c.length));
			break;
		}
	}

	if (value === "null") return null;
	if (value === "undefined") return undefined;
	// if (value !== default_value && type !== null) return ParserTypeMap[type](value);
	return value;
}

interface CetchOptions {
	attempts?: number,
	handlers?: { [statusCode: number]: { (r: Response): void; }; },
	method?: "GET" | "PUT" | "POST" | "PATCH" | "DELETE",
	credentials?: "include",
	headers?: { [header: string]: string; },
	body?: string | any,
	trace?: string,
	signal?: AbortSignal,
}

/**
 * 
 * @param url 
 * @param options interface CetchOptions {
 * 	attempts?:    number,
 * 	handlers?:    { [statusCode: number]: { (r: Response): void; }; },
 * 	method?:      "GET" | "PUT" | "POST" | "PATCH" | "DELETE",
 * 	credentials?: "include",
 * 	headers?:     { [header: string]: string; },
 * 	body?:        string | any,
 * 	trace?:       string,
 * 	signal?:      AbortSignal,
 * }
 * @returns 
 */
export async function cetch(url: string, options: CetchOptions = {}): Promise<Response> {
	const handlers = options?.handlers || {};
	options.headers = options?.headers || {};

	const a = GetCookie("ocm-auth");
	if (a) {
		options.credentials = "include";
		options.headers.authorization = "bearer " + a;
	}

	if (url.startsWith("/")) {
		url = window.location.origin + url;
	}

	let response: Response;
	try {
		response = await fetch(url, options);
	}
	catch (e) {
		throw e;
	}

	if (handlers[response.status]) {
		handlers[response.status](response);
		throw response;
	}
	else if (response.status < 400) {
		return response;
	}
	else if (response.status === 401) {
		// unset the auth cookie since it's no longer valid
		auth.value = undefined;
		document.cookie = `ocm-auth=nil; expires=${new Date(0)}; samesite=strict; path=/; secure`;
		throw response;
	}
	else if (response.status < 500) {
		const r = await response.json();
		// createToast({
		// 	title: errorMessage,
		// 	description: r?.error ?? apiErrorDescriptionToast,
		// 	dump: r,
		// });
		throw response;
	}
	return response;
}

export async function JsonPipeThrough(res: Response, abort: AbortController, func: ((json: any) => void | PromiseLike<void>)): Promise<void> {
	if (!res.body) return;
	const ro = res.body.pipeThrough(
		new TextDecoderStream("utf-8", { "fatal": false }),
		{ signal: abort.signal },
	).getReader();

	let ch = "";
	while (!abort.signal.aborted) {
		const r = await ro.read();
		if (r.done) return;
		ch += r.value;
		try {
			const rj: any = JSON.parse(ch).result;
			ch = "";
			func(rj);
		} catch {
			continue;
		}
	}
}
