<script lang="ts">
    import type { DumpInfo, User } from '$lib/types/dump';
    export let user: User;
    export let dumpInfo: DumpInfo;
</script>

<div class="text-center">
    {#if dumpInfo.logo1 || (dumpInfo.type === 'scolaire' && dumpInfo.logo2)}
        <div class="flex justify-center gap-4 mb-8">
            {#if dumpInfo.logo1}
                <img src={dumpInfo.logo1} alt="Logo 1" class="w-32 h-32 object-contain">
            {/if}
            {#if dumpInfo.type === 'scolaire' && dumpInfo.logo2}
                <img src={dumpInfo.logo2} alt="Logo 2" class="w-32 h-32 object-contain">
            {/if}
        </div>
    {/if}
    <h1 class="text-4xl md:text-6xl font-bold">Bonjour, {user.account_name}</h1>
    <p class="text-xl md:text-2xl mt-4">Voici ton {dumpInfo.title}!</p>
</div>
