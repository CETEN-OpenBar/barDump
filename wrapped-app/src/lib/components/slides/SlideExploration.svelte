<script lang="ts">
    import type { User } from '$lib/types/dump';
    export let user: User;
</script>

<div class="text-center">
    <p class="text-xl md:text-2xl">Tu es un explorateur !</p>
    <h2 class="text-6xl md:text-8xl font-bold my-4">{user.nombre_produits_differents}</h2>
    <p class="text-xl md:text-2xl">produits différents testés.</p>
    <p class="text-lg md:text-xl mt-2">Ce qui te classe #{user.rang_produits_differents} !</p>
</div>
