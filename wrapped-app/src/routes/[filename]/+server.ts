import { error } from '@sveltejs/kit';
import fs from 'fs';
import path from 'path';

export async function GET({ params }) {
    const filename = params.filename;

    // We only serve image files this way to avoid exposing sensitive files
    if (!filename.match(/\.(png|jpg|jpeg|svg|webp|gif)$/i)) {
        throw error(404, 'Not found');
    }

    const PROJECT_ROOT = path.resolve(process.cwd(), '..');
    const staticDir = path.resolve(PROJECT_ROOT, 'wrapped-app/static');
    const filePath = path.join(staticDir, filename);

    if (fs.existsSync(filePath)) {
        const file = fs.readFileSync(filePath);
        const ext = path.extname(filePath).toLowerCase();
        let contentType = 'image/png';
        if (ext === '.jpg' || ext === '.jpeg') contentType = 'image/jpeg';
        else if (ext === '.svg') contentType = 'image/svg+xml';
        else if (ext === '.webp') contentType = 'image/webp';
        else if (ext === '.gif') contentType = 'image/gif';

        return new Response(file, {
            headers: {
                'Content-Type': contentType,
                'Cache-Control': 'public, max-age=3600'
            }
        });
    }

    throw error(404, 'Not found');
}
