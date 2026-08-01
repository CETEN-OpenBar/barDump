import { redirect, type Handle } from '@sveltejs/kit';
import { env } from '$env/dynamic/private';

export const handle: Handle = async ({ event, resolve }) => {
    const isAdminRoute = event.url.pathname.startsWith('/admin');
    const isApiDumpsRoute = event.url.pathname.startsWith('/api/dumps');

    if (isAdminRoute || isApiDumpsRoute) {
        const token = event.cookies.get('admin_token');
        if (token !== env.ADMIN_TOKEN) {
            if (isApiDumpsRoute) {
                return new Response(JSON.stringify({ error: 'Unauthorized' }), {
                    status: 401,
                    headers: { 'Content-Type': 'application/json' }
                });
            } else {
                throw redirect(303, '/login');
            }
        }
    }

    return resolve(event);
};
