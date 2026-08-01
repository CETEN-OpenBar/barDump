<script lang="ts">
    import type { User } from '$lib/types/dump';
    export let user: User;
</script>

<div class="text-center">
    <p class="text-xl md:text-2xl">Tu es le plus gros consommateur</p>
    {#if user.top_consommateur_produit}
        <h2 class="text-4xl md:text-7xl font-bold my-4 px-4">{user.top_consommateur_produit.nom}</h2>
        <p class="text-xl md:text-2xl px-4">Tu en as consommé {user.top_consommateur_produit.quantite} sur un total de {user.top_consommateur_produit.total} !</p>
    {/if}
</div>
