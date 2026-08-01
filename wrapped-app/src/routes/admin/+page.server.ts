import { redirect } from '@sveltejs/kit';

export const actions = {
    logout: async ({ cookies }) => {
        cookies.delete('admin_token', { path: '/' });
        throw redirect(303, '/login');
    }
};
