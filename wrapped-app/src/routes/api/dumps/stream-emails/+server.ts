import { dumpManager } from '$lib/server/dumpManager';
import { error } from '@sveltejs/kit';

export async function GET({ url, request }) {
    const id = url.searchParams.get('id');
    
    if (!id) {
        throw error(400, 'Missing dump id');
    }

    const stream = new ReadableStream({
        start(controller) {
            dumpManager.sendEmailsStream(id, (msg) => {
                if (request.signal.aborted) return;
                try { controller.enqueue(`data: ${JSON.stringify(msg)}\n\n`); } catch(e) {}
            }, request.signal).then(() => {
                if (request.signal.aborted) return;
                try {
                    controller.enqueue(`data: ${JSON.stringify({ text: 'Process finished', type: 'done' })}\n\n`);
                    controller.close();
                } catch(e) {}
            }).catch(e => {
                if (request.signal.aborted) return;
                if (e.name === 'AbortError') return;
                try {
                    controller.enqueue(`data: ${JSON.stringify({ text: 'Error: ' + e.message, type: 'error' })}\n\n`);
                    controller.close();
                } catch(e) {}
            });
        }
    });

    return new Response(stream, {
        headers: {
            'Content-Type': 'text/event-stream',
            'Cache-Control': 'no-cache',
            'Connection': 'keep-alive'
        }
    });
}
