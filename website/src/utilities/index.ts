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

interface CetchOptions {
	attempts?: number,
	handlers?: { [statusCode: number]: { (r: Response): void; }; },
	method?: "GET" | "PUT" | "POST" | "PATCH" | "DELETE",
	credentials?: "include",
	headers?: { [header: string]: string; },
	body?: string | any,
	trace?: string,
}

/**
 * 
 * @param url 
 * @param options interface CetchOptions {
 * 	attempts?:      number,
 * 	handlers?: { [statusCode: number]: { (r: Response): void; }; },
 * 	method?:        "GET" | "PUT" | "POST" | "PATCH" | "DELETE",
 * 	credentials?:   "include",
 * 	headers?:       { [header: string]: string; },
 * 	body?:          string | any,
 * }
 * @returns 
 */
export async function cetch(url: string, options: CetchOptions = {}): Promise<Response> {
	const handlers = options?.handlers || {};
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
