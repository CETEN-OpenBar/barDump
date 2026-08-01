<script lang="ts">
    import { enhance } from '$app/forms';
    export let form;
</script>

<div class="min-h-screen bg-gray-900 text-white flex items-center justify-center p-8 font-sans">
    <div class="w-full max-w-md bg-gray-800 p-8 rounded-xl shadow-lg border border-gray-700">
        <h1 class="text-3xl font-bold mb-6 text-center text-blue-400">Connexion Admin</h1>
        
        <form method="POST" use:enhance class="space-y-6">
            <div class="space-y-2">
                <label for="token" class="block text-sm font-medium text-gray-300">Token d'accès</label>
                <input 
                    type="password" 
                    id="token" 
                    name="token" 
                    required 
                    class="w-full bg-gray-700 border border-gray-600 p-3 rounded-lg focus:ring-2 focus:ring-blue-500 outline-none transition-all"
                    placeholder="Entrez votre token"
                >
            </div>

            {#if form?.incorrect}
                <p class="text-red-400 text-sm font-medium">Token invalide</p>
            {/if}

            <button 
                type="submit" 
                class="w-full bg-blue-600 hover:bg-blue-700 text-white font-bold py-3 px-4 rounded-lg transition-colors"
            >
                Se connecter
            </button>
        </form>
        
        <div class="mt-6 text-center">
            <a href="/" class="text-gray-400 hover:text-white text-sm transition-colors">Retour au site</a>
        </div>
    </div>
</div>
