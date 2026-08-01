<script lang="ts">
    import type { User } from '$lib/types/dump';
    export let user: User;
</script>

<div class="text-center px-4">
    <h2 class="text-3xl md:text-5xl font-bold mb-8">Tes 3 produits les plus commandés</h2>
    <ul class="text-xl md:text-3xl space-y-4">
        {#each user.top_3_produits as product, i}
            <li>
                <span class="font-bold">{i + 1}.</span> {product.nom} <span class="text-base md:text-lg opacity-75">({product.quantite} fois)</span>
            </li>
        {/each}
    </ul>
</div>
