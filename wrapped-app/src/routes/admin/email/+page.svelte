<script lang="ts">
    import { onMount } from 'svelte';

    let config: any = { active_dump: '', dumps: {}, mail_template: '' };
    let loading = true;
    let sending = false;
    let savingTemplate = false;
    let successMessage = '';
    let errorMessage = '';
    let templateSuccessMessage = '';
    let currentEventSource: EventSource | null = null;
    let debugMode = false;
    let debugEmail = '';

    async function loadConfig() {
        const res = await fetch('/api/dumps');
        config = await res.json();
        loading = false;
    }

    async function saveTemplate() {
        savingTemplate = true;
        templateSuccessMessage = '';
        const res = await fetch('/api/dumps', {
            method: 'POST',
            body: JSON.stringify({ action: 'save_mail_template', template: config.mail_template })
        });
        savingTemplate = false;
        if (res.ok) {
            templateSuccessMessage = 'Template sauvegardé avec succès.';
            setTimeout(() => templateSuccessMessage = '', 3000);
        } else {
            const data = await res.json();
            alert('Erreur: ' + (data.details || data.error));
        }
    }

    let logs: Array<{ text: string, type: 'info'|'success'|'error'|'done' }> = [];

    function stopEmails() {
        if (currentEventSource) {
            currentEventSource.close();
            currentEventSource = null;
            sending = false;
            errorMessage = "Envoi des emails annulé par l'utilisateur.";
            // The browser closing the connection will trigger context cancellation on Go side.
        }
    }

    async function sendEmails(id: string) {
        if (!confirm(`Êtes-vous sûr de vouloir envoyer les emails pour le dump "${id}" ? Cette action est irréversible.`)) {
            return;
        }

        sending = true;
        successMessage = '';
        errorMessage = '';
        logs = [];
        
        const eventSource = new EventSource(`/api/dumps/stream-emails?id=${id}&debug_mode=${debugMode}&debug_email=${encodeURIComponent(debugEmail)}`);
        currentEventSource = eventSource;
        
        eventSource.onmessage = (event) => {
            const data = JSON.parse(event.data);
            logs = [...logs, data];
            
            // Auto scroll logs
            setTimeout(() => {
                const logContainer = document.getElementById('email-logs');
                if (logContainer) logContainer.scrollTop = logContainer.scrollHeight;
            }, 10);

            if (data.type === 'done') {
                eventSource.close();
                sending = false;
                successMessage = `Emails envoyés avec succès pour le dump ${id}.`;
            }
            if (data.type === 'error' && data.text.startsWith('Error:')) {
                eventSource.close();
                currentEventSource = null;
                sending = false;
                errorMessage = data.text;
            }
        };

        eventSource.onerror = (error) => {
            eventSource.close();
            currentEventSource = null;
            sending = false;
            if (!successMessage && !errorMessage) {
                errorMessage = "Erreur de connexion au flux d'envoi.";
            }
        };
    }

    onMount(loadConfig);
</script>

<div class="min-h-screen bg-gray-900 text-white p-8 font-sans">
    <div class="max-w-4xl mx-auto">
        <header class="mb-12 border-b border-gray-700 pb-6 flex justify-between items-center">
            <h1 class="text-4xl font-bold">Envoi d'e-mails - BarDump</h1>
            <div class="flex gap-4 items-center">
                <a href="/admin" class="text-blue-400 hover:underline">Retour à l'Admin</a>
                <a href="/" class="text-gray-400 hover:underline">Retour au site</a>
            </div>
        </header>

        {#if loading}
            <p>Chargement...</p>
        {:else}
            <section class="mb-12 bg-gray-800 p-6 rounded-xl shadow-lg">
                <h2 class="text-2xl font-semibold mb-2 text-blue-400">Éditeur de Template HTML</h2>
                <p class="text-sm text-gray-400 mb-6">
                    Vous pouvez utiliser les variables suivantes dans votre code HTML : 
                    <code class="bg-gray-700 px-1 rounded">{'{name}'}</code>, 
                    <code class="bg-gray-700 px-1 rounded">{'{dump_link}'}</code>, 
                    <code class="bg-gray-700 px-1 rounded">{'{base_url}'}</code>, 
                    <code class="bg-gray-700 px-1 rounded">{'{logos_html}'}</code>.
                </p>

                <div class="mb-4">
                    <textarea 
                        bind:value={config.mail_template}
                        class="w-full h-96 bg-gray-900 text-gray-300 p-4 rounded-lg font-mono text-sm border border-gray-700 focus:border-blue-500 focus:outline-none"
                    ></textarea>
                </div>
                
                <div class="flex justify-between items-center mb-8">
                    <button 
                        on:click={saveTemplate}
                        disabled={savingTemplate}
                        class="bg-blue-600 hover:bg-blue-700 disabled:bg-blue-900/50 px-6 py-2 rounded font-bold transition-colors"
                    >
                        {savingTemplate ? 'Sauvegarde...' : 'Sauvegarder le template'}
                    </button>
                    {#if templateSuccessMessage}
                        <span class="text-green-400 font-bold">{templateSuccessMessage}</span>
                    {/if}
                </div>

                <div>
                    <h3 class="text-xl font-semibold mb-4 text-gray-300">Aperçu en direct</h3>
                    <div class="bg-white rounded-lg overflow-hidden border border-gray-700 h-[600px] shadow-inner">
                        <iframe 
                            srcdoc={config.mail_template ? config.mail_template.replace(/\{name\}/g, 'Jean Dupont').replace(/\{dump_link\}/g, `${config.base_url || 'https://dump.bar.telecomancy.net'}/dump/2025/example-user-id`).replace(/\{base_url\}/g, config.base_url || 'https://dump.bar.telecomancy.net').replace(/\{logos_html\}/g, `<div style="margin-bottom: 25px; text-align: center;"><img src="${config.base_url || 'https://dump.bar.telecomancy.net'}/logo.png" width="150" style="display: inline-block; margin: 0 10px; max-width: 45%;"><img src="${config.base_url || 'https://dump.bar.telecomancy.net'}/logo.png" width="150" style="display: inline-block; margin: 0 10px; max-width: 45%;"></div>`) : ''} 
                            class="w-full h-full border-none"
                            title="Aperçu de l'email"
                        ></iframe>
                    </div>
                </div>
            </section>

            <section class="mb-12 bg-gray-800 p-6 rounded-xl shadow-lg">
                <div class="flex justify-between items-center mb-6">
                    <h2 class="text-2xl font-semibold text-blue-400">Sélectionnez le Dump pour l'envoi</h2>
                    {#if sending}
                        <button 
                            on:click={stopEmails}
                            class="bg-red-600 hover:bg-red-700 px-6 py-2 rounded font-bold transition-colors shadow-lg"
                        >
                            Stopper l'envoi
                        </button>
                    {/if}
                </div>
                
                {#if successMessage}
                    <div class="mb-6 p-4 bg-green-900/50 border border-green-500 rounded text-green-400">
                        {successMessage}
                    </div>
                {/if}
                
                {#if errorMessage}
                    <div class="mb-6 p-4 bg-red-900/50 border border-red-500 rounded text-red-400 whitespace-pre-wrap">
                        {errorMessage}
                    </div>
                {/if}

                <div class="mb-6 p-4 bg-gray-700/50 rounded-lg border border-gray-600">
                    <label class="flex items-center space-x-3 mb-3 cursor-pointer">
                        <input type="checkbox" bind:checked={debugMode} class="form-checkbox h-5 w-5 text-blue-600 bg-gray-800 border-gray-600 rounded focus:ring-blue-500 focus:ring-opacity-25" />
                        <span class="text-gray-200 font-semibold">Activer le Mode Debug</span>
                    </label>
                    
                    {#if debugMode}
                        <div class="ml-8">
                            <label class="block text-sm text-gray-400 mb-1" for="debug-email">Adresse e-mail de debug :</label>
                            <input type="email" id="debug-email" bind:value={debugEmail} placeholder="dev@example.com" class="w-full bg-gray-900 text-gray-300 p-2 rounded border border-gray-700 focus:border-blue-500 focus:outline-none" />
                            <p class="text-xs text-yellow-400 mt-2">Attention : Tous les emails générés seront envoyés à cette adresse au lieu des vrais destinataires.</p>
                        </div>
                    {/if}
                </div>

                <div class="space-y-4">
                    {#each Object.entries(config.dumps) as [id, dump]}
                        <div class="flex items-center justify-between p-4 bg-gray-700/50 rounded-lg border border-transparent">
                            <div>
                                <span class="font-bold text-lg">{dump.title}</span>
                                <span class="ml-2 text-xs text-gray-400 font-mono">ID: {id}</span>
                                <div class="text-sm text-gray-400 mt-1">
                                    Fichier source : {dump.file}
                                </div>
                            </div>
                            <div class="flex gap-2">
                                <button 
                                    on:click={() => sendEmails(id)}
                                    disabled={sending}
                                    class="bg-purple-600 hover:bg-purple-700 disabled:bg-gray-600 px-4 py-2 rounded text-sm font-bold transition-colors"
                                >
                                    {sending ? 'Envoi en cours...' : 'Envoyer les mails'}
                                </button>
                            </div>
                        </div>
                    {/each}
                </div>
                
                {#if Object.keys(config.dumps).length === 0}
                    <p class="text-gray-400 italic">Aucun dump disponible.</p>
                {/if}

                {#if logs.length > 0}
                    <div class="mt-8">
                        <h3 class="text-xl font-semibold mb-4 text-gray-300">Journal d'envoi en direct</h3>
                        <div id="email-logs" class="bg-gray-900 rounded-lg border border-gray-700 h-[300px] overflow-y-auto p-4 font-mono text-sm shadow-inner">
                            {#each logs as log}
                                <div class="mb-1 pb-1 border-b border-gray-800/50 flex gap-2">
                                    <span class="text-gray-500">[{new Date().toLocaleTimeString()}]</span>
                                    {#if log.type === 'success'}
                                        <span class="text-green-400 font-bold">[SUCCÈS]</span>
                                        <span class="text-green-300">{log.text}</span>
                                    {:else if log.type === 'error'}
                                        <span class="text-red-400 font-bold">[ERREUR]</span>
                                        <span class="text-red-300">{log.text}</span>
                                    {:else if log.type === 'info'}
                                        <span class="text-blue-400 font-bold">[INFO]</span>
                                        <span class="text-blue-200">{log.text}</span>
                                    {:else if log.type === 'done'}
                                        <span class="text-purple-400 font-bold">[TERMINÉ]</span>
                                        <span class="text-purple-300">{log.text}</span>
                                    {/if}
                                </div>
                            {/each}
                            {#if sending}
                                <div class="mt-2 text-gray-400 italic animate-pulse">En attente...</div>
                            {/if}
                        </div>
                    </div>
                {/if}
            </section>
            
            <div class="mt-8 p-4 bg-yellow-900/30 border border-yellow-700 rounded-lg">
                <h3 class="text-lg font-bold text-yellow-500 mb-2">⚠️ Attention</h3>
                <p class="text-sm text-yellow-300">
                    L'envoi d'e-mails peut prendre du temps (une pause de 2 secondes est appliquée par e-mail pour ne pas surcharger le serveur SMTP). 
                    Veuillez ne pas quitter la page pendant le traitement si vous voulez voir le message de succès.
                </p>
            </div>
        {/if}
    </div>
</div>
