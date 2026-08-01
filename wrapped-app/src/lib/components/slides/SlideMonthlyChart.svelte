<script lang="ts">
    import type { User } from '$lib/types/dump';
    export let user: User;

    const sortedMonths = Object.entries(user.articles_par_mois).sort(([a], [b]) => new Date(a).getTime() - new Date(b).getTime());
    const maxMonthlyValue = Math.max(...Object.values(user.articles_par_mois).map(m => m.nombre_transactions));
</script>

<div class="text-center w-full max-w-4xl px-4">
    <h2 class="text-5xl font-bold mb-8">Ton nombre de commande par mois</h2>
     <div class="flex justify-center items-end h-96 space-x-2 p-4 bg-white/5 rounded-lg">
         {#each sortedMonths as [month, data]}
             <div class="group flex-1 flex flex-col items-center justify-end h-full text-center">
                 <div class="flex flex-col items-center justify-end flex-grow w-full relative">
                     <span class="text-xs text-white/80 -top-5">{data.nombre_transactions}</span>
                     <div
                         class="w-full bg-blue-500 rounded-t-sm flex items-end justify-center pb-1 overflow-hidden transition-colors duration-300 group-hover:bg-blue-400"
                         style="height: {maxMonthlyValue > 0 ? (data.nombre_transactions / maxMonthlyValue) * 100 : 0}%;"
                         title="{data.produit_le_plus_achete}"
                     >
            
                     </div>
                 </div>
                 <span class="text-xs mt-2">{new Date(month).toLocaleString('fr-FR', { month: 'short' })}</span>
             </div>
         {/each}
     </div>
     <p class="mt-3">Tip : sur pc, passe ta souris sur les barres pour voir ton produit le plus consommé chaque mois</p>
</div>
