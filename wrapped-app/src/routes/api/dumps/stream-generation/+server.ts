import { dumpManager } from '$lib/server/dumpManager';
import { error } from '@sveltejs/kit';

export async function GET({ url, request }) {
    const action = url.searchParams.get('action');
    const id = url.searchParams.get('id');

    if (!action || !id) {
        throw error(400, 'Missing params');
    }

    const encoder = new TextEncoder();
    const stream = new ReadableStream({
        async start(controller) {
            let isClosed = false;
            
            // Keep-alive ping to prevent proxy/Ingress timeouts
            const pingInterval = setInterval(() => {
                if (request.signal.aborted || isClosed) {
                    clearInterval(pingInterval);
                    return;
                }
                try { controller.enqueue(encoder.encode(': ping\n\n')); } catch(e) {}
            }, 15000);

            const log = (msg: { text: string, type: 'info'|'success'|'error'|'done' }) => {
                if (request.signal.aborted || isClosed) return;
                try { controller.enqueue(encoder.encode(`data: ${JSON.stringify(msg)}\n\n`)); } catch(e) {}
            };

            try {
                const config = dumpManager.getConfig();
                
                let dumpToProcess: any = {};
                if (action === 'update') {
                    if (id !== 'all') {
                        throw new Error("Seul le dump 'all' peut être actualisé");
                    }
                    dumpToProcess = config.dumps[id];
                    if (!dumpToProcess) throw new Error('Dump not found');
                } else {
                    dumpToProcess = {
                        title: url.searchParams.get('title'),
                        type: url.searchParams.get('type'),
                        start_date: url.searchParams.get('start_date') || '',
                        end_date: url.searchParams.get('end_date') || '',
                        logo1: url.searchParams.get('logo1') || '',
                        logo2: url.searchParams.get('logo2') || ''
                    };
                }

                log({ text: 'Démarrage de la génération...', type: 'info' });
                
                log({ text: 'Export des données depuis MongoDB en cours... (cette étape peut prendre du temps)', type: 'info' });
                await dumpManager.exportData();
                log({ text: 'Export MongoDB terminé avec succès !', type: 'success' });

                log({ text: 'Analyse et traitement des transactions en cours... (cette étape peut prendre du temps)', type: 'info' });
                const fileName = await dumpManager.processStats({
                    id,
                    start_date: dumpToProcess.start_date,
                    end_date: dumpToProcess.end_date
                });
                log({ text: 'Traitement terminé avec succès !', type: 'success' });

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

                log({ text: 'Génération et mise à jour terminées avec succès !', type: 'done' });
                if (!request.signal.aborted && !isClosed) {
                    try { controller.close(); isClosed = true; } catch(e) {}
                }
            } catch(e: any) {
                if (e.name === 'AbortError') return;
                log({ text: 'Erreur: ' + e.message, type: 'error' });
                if (!request.signal.aborted && !isClosed) {
                    try { controller.close(); isClosed = true; } catch(e) {}
                }
            }
        }
    });

    return new Response(stream, {
        headers: {
            'Content-Type': 'text/event-stream',
            'Cache-Control': 'no-cache',
            'Connection': 'keep-alive',
            'X-Accel-Buffering': 'no',
            'Content-Encoding': 'none'
        }
    });
}
