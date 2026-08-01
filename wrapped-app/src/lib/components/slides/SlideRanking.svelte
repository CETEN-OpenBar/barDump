<script lang="ts">
    import type { User } from '$lib/types/dump';
    export let user: User;
</script>

<div class="text-center px-4">
    <p class="text-xl md:text-2xl">Tu es le</p>
    <h2 class="text-6xl md:text-8xl font-bold my-4">#{user.rang}</h2>
    <p class="text-xl md:text-2xl">plus grand consommateur, dans le top {user.rang_pourcentage}% !</p>
    <p class="text-xs opacity-50 mt-2">Sur un total de 417 consommateurs</p>
</div>
