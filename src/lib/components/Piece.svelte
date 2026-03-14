<script lang="ts">
    import Kitten from "./Kitten.svelte";
    import Cat from "./Cat.svelte";

    let {
        piece,
        placed,
        position,
        selectable,
        finalPosition = [0, 0, 0],
        booped,
    }: {
        piece: number;
        placed: boolean;
        position: any;
        selectable: boolean;
        finalPosition?: any[];
        booped: boolean;
    } = $props();

    const kitten = $derived(piece == 1 || piece == 8 ? true : piece == 2 || piece == 9 ? false : true);
    const color = $derived(piece == 1 || piece == 2 ? "orange" : piece == 8 || piece == 9 ? "lightblue" : "orange");
</script>

{#if kitten}
    <Kitten {color} {selectable} {placed} {position} {finalPosition} {booped} />
{:else}
    <Cat {color} {selectable} {placed} {position} {finalPosition} {booped} />
{/if}

