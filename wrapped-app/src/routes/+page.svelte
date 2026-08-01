<script lang="ts">
	export let data;
	const { activeDump, activeId } = data;
	// The user ID for the example link
	const sampleUserId = '01b62021-7c16-11ee-801c-3c219c2a5db';
</script>

<div class="bg-gray-900 text-white h-screen w-screen flex flex-col items-center justify-center text-center font-sans p-4">
	<main class="flex flex-col items-center justify-center">
		{#if activeDump}
			{#if activeDump.logo1}
				<img src={activeDump.logo1} alt="BarDump Logo" class="w-48 h-48 mb-8 object-contain">
			{/if}
			<h1 class="text-4xl md:text-6xl font-bold">Bienvenue sur BarDump</h1>
			<p class="text-xl md:text-2xl mt-4 max-w-2xl">
				Votre récapitulatif personnel {activeDump.title} est prêt.
			</p>
			<a 
				href="/dump/{activeId}/{sampleUserId}" 
				class="mt-10 px-8 py-4 text-lg font-semibold text-white bg-blue-600 rounded-full shadow-lg hover:bg-blue-700 transition-colors duration-300 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 focus:ring-offset-gray-900"
			>
				Voir mon {activeDump.title}
			</a>
		{:else}
			<h1 class="text-4xl md:text-6xl font-bold">Aucun récapitulatif actif</h1>
			<p class="text-xl md:text-2xl mt-4 max-w-2xl">
				Veuillez configurer un dump actif dans l'administration.
			</p>
		{/if}
	</main>
</div>