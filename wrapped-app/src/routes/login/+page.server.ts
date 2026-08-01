import { fail, redirect } from '@sveltejs/kit';
import { env } from '$env/dynamic/private';
import { dev } from '$app/environment';

export const load = ({ cookies }) => {
    const token = cookies.get('admin_token');
    if (token && token === env.ADMIN_TOKEN) {
        throw redirect(303, '/admin');
    }
};

export const actions = {
    default: async ({ request, cookies }) => {
        const data = await request.formData();
        const token = data.get('token');

        if (token === env.ADMIN_TOKEN) {
            cookies.set('admin_token', token as string, { 
                path: '/', 
                httpOnly: true, 
                maxAge: 60 * 60 * 24 * 30 // 30 days
            });
            throw redirect(303, '/admin');
        }

        return fail(401, { incorrect: true });
    }
};
