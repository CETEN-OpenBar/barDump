<script lang="ts">
    import type { DumpInfo, User } from '$lib/types/dump';
    export let user: User;
    export let dumpInfo: DumpInfo;
</script>

<div class="text-center">
    <p class="text-xl md:text-2xl">En {dumpInfo.title}, tu as dépensé</p>
    <h2 class="text-6xl md:text-8xl font-bold my-4">{(user.total_depense / 100).toFixed(2)}€</h2>
</div>
