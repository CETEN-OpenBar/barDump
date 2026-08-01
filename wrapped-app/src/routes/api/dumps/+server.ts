import { json } from '@sveltejs/kit';
import { dumpManager } from '$lib/server/dumpManager';
import { env } from '$env/dynamic/private';

export async function GET() {
    try {
        const config = dumpManager.getConfig();
        const mail_template = dumpManager.getMailTemplate();
        const base_url = env.BASE_URL || 'https://dump.bar.telecomancy.net';
        return json({ ...config, mail_template, base_url });
    } catch (e) {
        return json({ error: 'Failed to load config' }, { status: 500 });
    }
}

export async function POST({ request }) {
    const contentType = request.headers.get('content-type') || '';

    // Gestion de l'upload de fichier
    if (contentType.includes('multipart/form-data')) {
        try {
            const formData = await request.formData();
            const file = formData.get('logo') as File;
            if (!file) return json({ error: 'No file uploaded' }, { status: 400 });

            const fileName = await dumpManager.saveLogo(file);
            return json({ success: true, fileName });
        } catch (e: any) {
            return json({ error: 'Upload failed', details: e.message }, { status: 500 });
        }
    }

    const body = await request.json();
    const { action, id, ...params } = body;

    if (action === 'set_active') {
        try {
            const config = dumpManager.getConfig();
            if (!config.dumps[id]) return json({ error: 'Dump not found' }, { status: 404 });
            config.active_dump = id;
            dumpManager.saveConfig(config);
            return json({ success: true });
        } catch (e) {
            return json({ error: 'Failed to update config' }, { status: 500 });
        }
    }

    if (action === 'create' || action === 'update') {
        if (action === 'update' && id !== 'all') {
            return json({ error: "Seul le dump 'all' peut être actualisé" }, { status: 403 });
        }

        try {
            const config = dumpManager.getConfig();
            let dumpToProcess = action === 'update' ? config.dumps[id] : params;
            if (!dumpToProcess && action === 'update') return json({ error: 'Dump not found' }, { status: 404 });

            await dumpManager.exportData();
            const fileName = await dumpManager.processStats({
                id,
                start_date: dumpToProcess.start_date,
                end_date: dumpToProcess.end_date
            });

            config.dumps[id] = {
                title: dumpToProcess.title || (config.dumps[id] ? config.dumps[id].title : id),
                type: dumpToProcess.type || (config.dumps[id] ? config.dumps[id].type : 'civil'),
                file: fileName,
                logo1: dumpToProcess.logo1,
                logo2: dumpToProcess.logo2,
                start_date: id === 'all' ? undefined : dumpToProcess.start_date,
                end_date: id === 'all' ? undefined : dumpToProcess.end_date
            };
            dumpManager.saveConfig(config);

            return json({ success: true });
        } catch (e: any) {
            console.error(e);
            return json({ error: 'Execution failed', details: e.message }, { status: 500 });
        }
    }

    if (action === 'list_logos') {
        try {
            const logos = await dumpManager.listAvailableLogos();
            return json({ logos });
        } catch (e) {
            return json({ error: 'Failed to list logos' }, { status: 500 });
        }
    }

    if (action === 'send_emails') {
        if (!id) return json({ error: 'Missing dump id' }, { status: 400 });
        try {
            await dumpManager.sendEmails(id);
            return json({ success: true });
        } catch (e: any) {
            console.error(e);
            return json({ error: 'Failed to send emails', details: e.message }, { status: 500 });
        }
    }

    if (action === 'save_mail_template') {
        try {
            dumpManager.saveMailTemplate(body.template);
            return json({ success: true });
        } catch (e: any) {
            console.error(e);
            return json({ error: 'Failed to save mail template', details: e.message }, { status: 500 });
        }
    }

    return json({ error: 'Invalid action' }, { status: 400 });
}
