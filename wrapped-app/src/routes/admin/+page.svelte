<script lang="ts">
    import { onMount } from 'svelte';

    let config = { active_dump: '', dumps: {} };
    let availableLogos = [];
    let loading = true;
    let creating = false;
    let uploading = false;
    let logs: Array<{ text: string, type: 'info'|'success'|'error'|'done' }> = [];
    let currentEventSource: EventSource | null = null;
    let showLogsModal = false;

    // Formulaire de création
    let newDump = {
        id: '',
        title: '',
        type: 'civil',
        start_date: '',
        end_date: '',
        logo1: '/logo.png',
        logo2: ''
    };

    let fileInput;

    async function loadConfig() {
        const res = await fetch('/api/dumps');
        config = await res.json();
    }

    async function loadLogos() {
        const res = await fetch('/api/dumps', {
            method: 'POST',
            body: JSON.stringify({ action: 'list_logos' })
        });
        const data = await res.json();
        availableLogos = data.logos || [];
    }

    async function init() {
        await Promise.all([loadConfig(), loadLogos()]);
        loading = false;
    }

    async function setActive(id: string) {
        const res = await fetch('/api/dumps', {
            method: 'POST',
            body: JSON.stringify({ action: 'set_active', id })
        });
        if (res.ok) await loadConfig();
    }

    function handleSSE(url: string, onSuccess: () => void) {
        creating = true;
        showLogsModal = true;
        logs = [];
        
        const eventSource = new EventSource(url);
        currentEventSource = eventSource;
        
        eventSource.onmessage = (event) => {
            const data = JSON.parse(event.data);
            logs = [...logs, data];
            
            setTimeout(() => {
                const logContainer = document.getElementById('gen-logs');
                if (logContainer) logContainer.scrollTop = logContainer.scrollHeight;
            }, 10);

            if (data.type === 'done') {
                eventSource.close();
                currentEventSource = null;
                creating = false;
                onSuccess();
            }
            if (data.type === 'error') {
                eventSource.close();
                currentEventSource = null;
                creating = false;
            }
        };

        eventSource.onerror = () => {
            eventSource.close();
            currentEventSource = null;
            creating = false;
            logs = [...logs, { text: "Erreur de connexion au flux de génération.", type: 'error' }];
        };
    }

    async function createDump() {
        const params = new URLSearchParams({
            action: 'create',
            id: newDump.id,
            title: newDump.title,
            type: newDump.type,
            start_date: newDump.start_date,
            end_date: newDump.end_date,
            logo1: newDump.logo1,
            logo2: newDump.logo2
        });
        handleSSE(`/api/dumps/stream-generation?${params.toString()}`, async () => {
            await loadConfig();
            newDump = { id: '', title: '', type: 'civil', start_date: '', end_date: '', logo1: '/logo.png', logo2: '' };
        });
    }

    async function updateDump(id: string) {
        if (!confirm(`Voulez-vous vraiment actualiser les données du dump "${id}" ?`)) return;
        
        const params = new URLSearchParams({ action: 'update', id });
        handleSSE(`/api/dumps/stream-generation?${params.toString()}`, async () => {
            await loadConfig();
        });
    }

    function closeLogsModal() {
        showLogsModal = false;
        if (currentEventSource) {
            currentEventSource.close();
            currentEventSource = null;
            creating = false;
        }
    }

    async function uploadLogo() {
        const file = fileInput.files[0];
        if (!file) return;

        // Limite d'upload à 5Mo (5 * 1024 * 1024 bytes)
        const MAX_FILE_SIZE = 5 * 1024 * 1024;
        if (file.size > MAX_FILE_SIZE) {
            alert('Le fichier est trop volumineux. La taille maximum est de 5 Mo.');
            return;
        }

        uploading = true;
        const formData = new FormData();
        formData.append('logo', file);

        const res = await fetch('/api/dumps', {
            method: 'POST',
            body: formData
        });

        uploading = false;
        if (res.ok) {
            const data = await res.json();
            await loadLogos();
            newDump.logo1 = data.fileName;
            fileInput.value = '';
        } else {
            alert("Erreur lors de l'upload");
        }
    }

    onMount(init);
</script>

<div class="min-h-screen bg-gray-900 text-white p-8 font-sans">
    <div class="max-w-4xl mx-auto">
        <header class="mb-12 border-b border-gray-700 pb-6 flex justify-between items-center">
            <h1 class="text-4xl font-bold">Admin Dashboard - BarDump</h1>
            <div class="flex gap-4 items-center">
                <a href="/" class="text-blue-400 hover:underline">Retour au site</a>
                <form method="POST" action="?/logout">
                    <button type="submit" class="text-red-400 hover:underline">Déconnexion</button>
                </form>
            </div>
        </header>

        {#if loading}
            <p>Chargement...</p>
        {:else}
            <!-- Section Liste des Dumps & Actualisation -->
            <section class="mb-12 bg-gray-800 p-6 rounded-xl shadow-lg">
                <div class="flex justify-between items-center mb-6">
                    <h2 class="text-2xl font-semibold text-blue-400">Dumps Existants & Actif</h2>
                    <a 
                        href="/admin/email"
                        class="bg-purple-600 hover:bg-purple-700 px-4 py-2 rounded font-bold transition-colors flex items-center shadow-lg"
                    >
                        Gestion des E-mails
                    </a>
                </div>
                
                <div class="space-y-4">
                    {#each Object.entries(config.dumps) as [id, dump]}
                        <div class="flex items-center justify-between p-4 bg-gray-700/50 rounded-lg border {config.active_dump === id ? 'border-blue-500' : 'border-transparent'}">
                            <div>
                                <span class="font-bold text-lg">{dump.title}</span>
                                <span class="ml-2 text-xs text-gray-400 font-mono">ID: {id}</span>
                                <div class="text-sm text-gray-400">
                                    Type: {dump.type} | 
                                    Logos: {dump.logo1 ? '1' : '0'}{dump.logo2 ? ' + 1' : ''}
                                </div>
                            </div>
                            <div class="flex gap-2">
                                {#if id === 'all'}
                                    <button 
                                        on:click={() => updateDump(id)}
                                        disabled={creating}
                                        class="bg-amber-600 hover:bg-amber-700 disabled:bg-gray-600 px-3 py-1 rounded text-sm font-bold transition-colors"
                                        title="Relance l'export et le parsing pour l'intégralité de l'historique"
                                    >
                                        Actualiser
                                    </button>
                                {/if}

                                <button 
                                    on:click={() => setActive(id)}
                                    disabled={config.active_dump === id}
                                    class="bg-blue-600 hover:bg-blue-700 disabled:bg-blue-900/50 px-3 py-1 rounded text-sm font-bold transition-colors"
                                >
                                    {config.active_dump === id ? 'Actif' : 'Rendre Actif'}
                                </button>
                            </div>
                        </div>
                    {/each}
                </div>
            </section>

            <!-- Section Logos & Upload -->
            <section class="mb-12 bg-gray-800 p-6 rounded-xl shadow-lg">
                <h2 class="text-2xl font-semibold mb-6 text-purple-400">Gestion des Logos</h2>
                <div class="flex flex-col md:flex-row gap-6 items-end">
                    <div class="flex-grow">
                        <label for="file-upload" class="text-sm font-medium text-gray-300 block mb-2">Ajouter un nouveau logo (PNG, JPG, SVG...)</label>
                        <input 
                            id="file-upload" 
                            type="file" 
                            accept="image/*"
                            bind:this={fileInput}
                            class="block w-full text-sm text-gray-400
                                file:mr-4 file:py-2 file:px-4
                                file:rounded-full file:border-0
                                file:text-sm file:font-semibold
                                file:bg-purple-600 file:text-white
                                hover:file:bg-purple-700"
                        />
                    </div>
                    <button 
                        on:click={uploadLogo}
                        disabled={uploading}
                        class="bg-purple-600 hover:bg-purple-700 disabled:bg-gray-600 px-6 py-2 rounded font-bold transition-colors h-[38px]"
                    >
                        {uploading ? 'Upload...' : 'Uploader le logo'}
                    </button>
                </div>
                
                <div class="mt-6 flex flex-wrap gap-4">
                    {#each availableLogos as logo}
                        <div class="group relative">
                            <img src={logo} alt={logo} class="w-16 h-16 object-contain bg-black/20 p-1 rounded border border-gray-700" title={logo}>
                        </div>
                    {/each}
                </div>
            </section>

            <!-- Section Création -->
            <section class="bg-gray-800 p-6 rounded-xl shadow-lg border-2 border-green-900/30">
                <h2 class="text-2xl font-semibold mb-6 text-green-400">Créer un nouveau Dump</h2>
                <form on:submit|preventDefault={createDump} class="grid grid-cols-1 md:grid-cols-2 gap-6">
                    <div class="flex flex-col gap-2">
                        <label for="id" class="text-sm font-medium text-gray-300">ID unique (ex: 2024-2025)</label>
                        <input id="id" bind:value={newDump.id} required placeholder="2024-2025" class="bg-gray-700 border border-gray-600 p-2 rounded focus:ring-2 focus:ring-green-500 outline-none font-mono">
                    </div>
                    <div class="flex flex-col gap-2">
                        <label for="title" class="text-sm font-medium text-gray-300">Titre affiché (ex: BarDump 24/25)</label>
                        <input id="title" bind:value={newDump.title} required placeholder="BarDump 24/25" class="bg-gray-700 border border-gray-600 p-2 rounded focus:ring-2 focus:ring-green-500 outline-none">
                    </div>
                    <div class="flex flex-col gap-2">
                        <label for="type" class="text-sm font-medium text-gray-300">Type de Dump</label>
                        <select id="type" bind:value={newDump.type} class="bg-gray-700 border border-gray-600 p-2 rounded outline-none">
                            <option value="civil">Année Civile (1 Bar)</option>
                            <option value="scolaire">Année Scolaire (2 Bars)</option>
                        </select>
                    </div>
                    <div class="flex flex-col gap-2 invisible md:visible"></div>
                    
                    <div class="flex flex-col gap-2 border-l-2 border-green-600/30 pl-4">
                        <label for="start" class="text-sm font-medium text-gray-300">Date de début (optionnel)</label>
                        <input id="start" type="date" bind:value={newDump.start_date} class="bg-gray-700 border border-gray-600 p-2 rounded outline-none">
                    </div>
                    <div class="flex flex-col gap-2 border-l-2 border-green-600/30 pl-4">
                        <label for="end" class="text-sm font-medium text-gray-300">Date de fin (optionnel)</label>
                        <input id="end" type="date" bind:value={newDump.end_date} class="bg-gray-700 border border-gray-600 p-2 rounded outline-none">
                    </div>

                    <div class="flex flex-col gap-2">
                        <label for="logo1" class="text-sm font-medium text-gray-300">Logo 1</label>
                        <select id="logo1" bind:value={newDump.logo1} class="bg-gray-700 border border-gray-600 p-2 rounded outline-none">
                            <option value="">-- Aucun logo --</option>
                            {#each availableLogos as logo}
                                <option value={logo}>{logo}</option>
                            {/each}
                        </select>
                    </div>

                    {#if newDump.type === 'scolaire'}
                        <div class="flex flex-col gap-2">
                            <label for="logo2" class="text-sm font-medium text-gray-300">Logo 2</label>
                            <select id="logo2" bind:value={newDump.logo2} class="bg-gray-700 border border-gray-600 p-2 rounded outline-none">
                                <option value="">-- Aucun logo --</option>
                                {#each availableLogos as logo}
                                    <option value={logo}>{logo}</option>
                                {/each}
                            </select>
                        </div>
                    {/if}

                    <div class="md:col-span-2 mt-4">
                        <button 
                            type="submit" 
                            disabled={creating}
                            class="w-full bg-green-600 hover:bg-green-700 disabled:bg-gray-600 px-8 py-3 rounded-lg font-bold text-lg transition-all"
                        >
                            {#if creating}
                                Opération en cours...
                            {:else}
                                Lancer la génération
                            {/if}
                        </button>
                    </div>
                </form>
            </section>
        {/if}
    </div>
</div>

{#if showLogsModal}
<div class="fixed inset-0 z-50 flex items-center justify-center bg-black/80 backdrop-blur-sm p-4">
    <div class="bg-gray-900 border border-gray-700 rounded-xl shadow-2xl w-full max-w-3xl flex flex-col overflow-hidden">
        <div class="px-6 py-4 border-b border-gray-800 flex justify-between items-center bg-gray-800">
            <h3 class="text-xl font-bold text-blue-400">Génération en cours</h3>
            {#if !creating}
                <button on:click={closeLogsModal} class="text-gray-400 hover:text-white font-bold text-2xl leading-none">&times;</button>
            {/if}
        </div>
        <div id="gen-logs" class="p-6 bg-gray-950 h-96 overflow-y-auto font-mono text-sm shadow-inner flex flex-col gap-2">
            {#each logs as log}
                <div class="flex gap-2">
                    <span class="text-gray-600 shrink-0">[{new Date().toLocaleTimeString()}]</span>
                    {#if log.type === 'success'}
                        <span class="text-green-500 shrink-0">[SUCCÈS]</span>
                        <span class="text-green-300">{log.text}</span>
                    {:else if log.type === 'error'}
                        <span class="text-red-500 shrink-0">[ERREUR]</span>
                        <span class="text-red-300">{log.text}</span>
                    {:else if log.type === 'info'}
                        <span class="text-blue-500 shrink-0">[INFO]</span>
                        <span class="text-blue-200">{log.text}</span>
                    {:else if log.type === 'done'}
                        <span class="text-purple-500 shrink-0">[TERMINÉ]</span>
                        <span class="text-purple-300">{log.text}</span>
                    {/if}
                </div>
            {/each}
            {#if creating}
                <div class="text-gray-500 italic mt-2 animate-pulse flex items-center gap-2">
                    <svg class="animate-spin h-4 w-4 text-gray-500" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                        <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                        <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                    </svg>
                    Traitement en cours...
                </div>
            {/if}
        </div>
        <div class="px-6 py-4 bg-gray-800 border-t border-gray-700 flex justify-end">
            {#if creating}
                <button on:click={closeLogsModal} class="bg-red-600 hover:bg-red-700 px-6 py-2 rounded font-bold transition-colors">
                    Annuler la génération
                </button>
            {:else}
                <button on:click={closeLogsModal} class="bg-blue-600 hover:bg-blue-700 px-6 py-2 rounded font-bold transition-colors">
                    Fermer
                </button>
            {/if}
        </div>
    </div>
</div>
{/if}
