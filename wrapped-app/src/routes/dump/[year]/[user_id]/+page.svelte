<script lang="ts">
    import { fly } from 'svelte/transition';
    import SlideWelcome from '$lib/components/slides/SlideWelcome.svelte';
    import SlideTotalSpent from '$lib/components/slides/SlideTotalSpent.svelte';
    import SlideRanking from '$lib/components/slides/SlideRanking.svelte';
    import SlideTopProducts from '$lib/components/slides/SlideTopProducts.svelte';
    import SlideMonthlyChart from '$lib/components/slides/SlideMonthlyChart.svelte';
    import SlideTopConsumer from '$lib/components/slides/SlideTopConsumer.svelte';
    import SlideExploration from '$lib/components/slides/SlideExploration.svelte';
    import SlideOrderType from '$lib/components/slides/SlideOrderType.svelte';
    import SlideEnd from '$lib/components/slides/SlideEnd.svelte';

    export let data;

    const { user, dumpInfo } = data;

    // Configuration des slides
    const slides = [
        { id: 0, component: SlideWelcome },
        { id: 1, component: SlideTotalSpent },
        { id: 2, component: SlideRanking },
        { id: 3, component: SlideTopProducts },
        { id: 4, component: SlideMonthlyChart }
    ];

    if (user.top_consommateur_produit) {
        slides.push({ id: 5, component: SlideTopConsumer });
    }

    slides.push(
        { id: 6, component: SlideExploration },
        { id: 7, component: SlideOrderType },
        { id: 8, component: SlideEnd }
    );

    let currentSlideIndex = 0;
    const totalSlides = slides.length;

    function nextSlide() {
        if (currentSlideIndex < totalSlides - 1) currentSlideIndex++;
    }

    function prevSlide() {
        if (currentSlideIndex > 0) currentSlideIndex--;
    }
</script>

<div
    role="button"
    tabindex="0"
    on:click={nextSlide}
    on:keydown={(e) => { if (e.key === 'ArrowRight') nextSlide(); if (e.key === 'ArrowLeft') prevSlide(); }}
    class="bg-gray-900 text-white h-screen w-screen flex flex-col items-center justify-center font-sans overflow-hidden focus:outline-none focus:ring-2 focus:ring-blue-500"
>
    <div class="absolute top-4 left-4 text-xs opacity-50">Slide {currentSlideIndex + 1} / {totalSlides} (Click to advance)</div>
    <div class="absolute top-4 right-4 text-xs opacity-50">{user.account_name}</div>
    
    <main class="w-full h-full flex items-center justify-center">
        {#key currentSlideIndex}
            <div in:fly={{ y: 50, duration: 1000, delay: 300 }} class="w-full h-full flex items-center justify-center">
                <svelte:component 
                    this={slides[currentSlideIndex].component} 
                    {user} 
                    {dumpInfo} 
                />
            </div>
        {/key}
    </main>

    <div class="absolute bottom-4 w-full flex justify-between px-8">
        <button 
            on:click|stopPropagation={prevSlide} 
            class="px-4 py-2 rounded-full bg-white/10 backdrop-blur-sm disabled:opacity-25" 
            disabled={currentSlideIndex === 0}
        >
            Précédent
        </button>
        <button 
            on:click|stopPropagation={nextSlide} 
            class="px-4 py-2 rounded-full bg-white/10 backdrop-blur-sm disabled:opacity-25" 
            disabled={currentSlideIndex === totalSlides - 1}
        >
            Suivant
        </button>
    </div>
</div>
