<script lang="ts">
    import type { User } from '$lib/types/dump';
    export let user: User;
</script>

<div class="text-center">
    {#if user.nombre_transaction_remote > user.nombre_transaction_non_remote}
         <div class="text-center">
            <p class="text-xl md:text-2xl">Tu es plutôt du genre connecté</p>
            <h2 class="text-4xl md:text-7xl font-bold my-4 px-4">{user.nombre_transaction_remote}</h2>
            <p class="text-xl md:text-2xl">commandes faites en ligne contre</p>
            <p class="text-xl md:text-2xl"> </p>
            <h2 class="text-2xl md:text-5xl font-bold my-4 px-4">{user.nombre_transaction_non_remote}</h2>
            <p class="text-xl md:text-2xl px-4">commandes faites sur les bornes</p>
        </div>
    {:else}
        <div class="text-center">
            <p class="text-xl md:text-2xl">Tu es plutôt du genre physique</p>
            <h2 class="text-4xl md:text-7xl font-bold my-4 px-4">{user.nombre_transaction_non_remote}</h2>
            <p class="text-xl md:text-2xl">commandes faites sur les bornes contre</p>
            <p class="text-xl md:text-2xl"> </p>
            <h2 class="text-2xl md:text-5xl font-bold my-4 px-4">{user.nombre_transaction_remote}</h2>
            <p class="text-xl md:text-2xl px-4">commandes faites en ligne</p>
        </div>
    {/if}
</div>
