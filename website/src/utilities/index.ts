import { useRouter, type Router } from 'vue-router';

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

export function GetCookie(cookieName: string, default_value: any = null, type: string | null = null) {
	const name = cookieName + "=";
	let ca = document.cookie.split(";");
	let value: any = default_value;
	for (let i = 0; i < ca.length; i++) {
		let c = ca[i];
		while (c.charAt(0) == " ") {
			c = c.substring(1);
		}

		if (c.indexOf(name) == 0) {
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
	router?: Router,
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
 *  trace?:       string,
 * }
 * @returns 
 */
export async function cetch(url: string, options: CetchOptions = {}): Promise<Response> {
	const handlers = options?.handlers || {};
	options.headers = options?.headers || {};

	const auth = GetCookie("ocm-auth");
	if (auth) {
		options.credentials = "include";
		options.headers.authorization = "bearer " + auth;
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
		document.cookie = `ocm-auth=nil; expires=${new Date(0)}; samesite=strict; path=/; secure`;
		options.router?.push("/user/login");
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
